package service

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"

	"github.com/lib/pq"
)

func (s *CustomGroupAccessService) wakeGroupExpiry() {
	select {
	case s.expiryWake <- struct{}{}:
	default:
	}
}

func (s *CustomGroupAccessService) startGroupExpiry(keys *APIKeyService) func() {
	store, ok := s.store.(*customGroupAccessSQLStore)
	if !ok {
		return func() {}
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		for ctx.Err() == nil {
			delay := s.runGroupExpiry(ctx, store, keys)
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
			case <-s.expiryWake:
			case <-timer.C:
			}
			timer.Stop()
		}
	}()
	var once sync.Once
	return func() { once.Do(func() { cancel(); <-done }) }
}

func (s *CustomGroupAccessService) runGroupExpiry(parent context.Context, store *customGroupAccessSQLStore, keys *APIKeyService) time.Duration {
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	now := s.now()
	ids, err := store.dueGroupExpiries(ctx, now)
	cancel()
	if err != nil {
		if parent.Err() == nil {
			slog.Error("读取到期分组失败，稍后重试", "error", err)
		}
		return 30 * time.Second
	}
	failed := make([]int64, 0)
	for _, groupID := range ids {
		if parent.Err() != nil {
			return 30 * time.Second
		}
		ctx, cancel := context.WithTimeout(parent, 10*time.Second)
		move, err := store.moveExpiredGroup(ctx, groupID, now)
		cancel()
		if err != nil {
			failed = append(failed, groupID)
			slog.Error("分组到期平移未完成，保留原分组并稍后重试", "group_id", groupID, "error", err)
			continue
		}
		if move == nil {
			continue
		}
		if s.cache != nil {
			_ = s.cache.changed(`{"kind":"reset"}`)
		}
		if keys != nil {
			ctx, cancel := context.WithTimeout(parent, 5*time.Second)
			for _, key := range move.authKeys {
				keys.InvalidateAuthCacheByKey(ctx, key)
			}
			cancel()
		}
		slog.Info("到期分组已平移并删除", "group_id", groupID, "target_group_id", move.targetID, "users", move.users, "keys", move.keys)
	}
	ctx, cancel = context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	var next sql.NullTime
	err = store.db.QueryRowContext(ctx, `SELECT MIN(p.expires_at) FROM custom_group_access_policies p
 JOIN groups g ON g.id=p.group_id WHERE g.deleted_at IS NULL AND p.expires_at IS NOT NULL
 AND NOT (p.group_id=ANY($1::bigint[]))`, pq.Array(failed)).Scan(&next)
	if err != nil {
		return 30 * time.Second
	}
	return customGroupExpiryDelay(next, s.now())
}

func customGroupExpiryDelay(next sql.NullTime, now time.Time) time.Duration {
	if !next.Valid {
		return 30 * time.Second
	}
	delay := next.Time.Sub(now)
	if delay < 0 {
		return 0
	}
	if delay > 30*time.Second {
		return 30 * time.Second
	}
	return delay
}
