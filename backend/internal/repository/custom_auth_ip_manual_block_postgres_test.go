//go:build unit

package repository

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAuthIPManualBlockRealPostgres(t *testing.T) {
	dsn := os.Getenv("AUTH_IP_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()
	require.NoError(t, ApplyMigrations(ctx, db))
	require.NoError(t, ApplyMigrations(ctx, db))
	_, err = db.ExecContext(ctx, `TRUNCATE custom_auth_ip_blacklist RESTART IDENTITY;
UPDATE custom_auth_ip_blacklist_config SET enabled = FALSE, threshold = 10, window_seconds = 7200, block_seconds = 900, block_mode = 'timed', version = 1 WHERE id = 1`)
	require.NoError(t, err)
	repo := NewAuthIPBlacklistRepository(db)
	guard := service.NewAuthIPBlacklistService(repo, nil)
	input := service.AuthIPManualBlock{IP: "::ffff:192.0.2.40", BlockMode: service.AuthIPBlockModePermanent, Reason: "持续恶意扫描"}
	entry, err := guard.ManualBlock(ctx, input)
	require.NoError(t, err)
	require.Equal(t, "192.0.2.40", entry.IP)
	require.Equal(t, "manual", entry.Source)
	require.Equal(t, input.Reason, entry.Reason)
	require.Nil(t, entry.BlockedUntil)
	require.Nil(t, entry.FirstEventAt)
	require.Nil(t, entry.LastEventAt)
	require.Zero(t, entry.TriggerCount)
	require.Zero(t, entry.WindowSeconds)
	_, err = service.NewAuthIPBlacklistService(NewAuthIPBlacklistRepository(db), nil).Check(ctx, input.IP, "/api/v1/auth/login", true)
	require.ErrorIs(t, err, service.ErrAuthIPBlocked, "自动功能关闭且无 Redis 仍拦截持久手动名单")
	input.BlockMode = service.AuthIPBlockModeTimed
	input.BlockSeconds = 60
	_, err = guard.ManualBlock(ctx, input)
	require.ErrorIs(t, err, service.ErrAuthIPAlreadyBlocked)
	snapshot, err := repo.Snapshot(ctx, entry.IP)
	require.NoError(t, err)
	require.Nil(t, snapshot.Entry.BlockedUntil, "重复添加不得缩短永久封禁")
	_, err = guard.Unblock(ctx, entry.ID, entry.Generation)
	require.NoError(t, err)
	timed, err := guard.ManualBlock(ctx, input)
	require.NoError(t, err)
	require.Greater(t, timed.Generation, entry.Generation)
	require.WithinDuration(t, timed.BlockedAt.Add(time.Minute), *timed.BlockedUntil, time.Millisecond)
	_, err = guard.Unblock(ctx, entry.ID, entry.Generation)
	require.ErrorIs(t, err, service.ErrAuthIPBlacklistConflict)
	_, err = db.ExecContext(ctx, `UPDATE custom_auth_ip_blacklist SET blocked_until = NOW() - INTERVAL '1 minute' WHERE id = $1`, timed.ID)
	require.NoError(t, err)
	items, total, err := repo.List(ctx, entry.IP, 1, 20)
	require.NoError(t, err)
	require.Zero(t, total)
	require.Empty(t, items)
	_, err = guard.Check(ctx, entry.IP, "/api/v1/auth/login", true)
	require.NoError(t, err, "到期自动解除，关闭自动计数时无需 Redis")
	snapshot, err = repo.Snapshot(ctx, entry.IP)
	require.NoError(t, err)
	require.Greater(t, snapshot.Entry.Generation, timed.Generation)

	config, err := repo.Config(ctx)
	require.NoError(t, err)
	config.Enabled = true
	config, err = guard.UpdateConfig(ctx, *config)
	require.NoError(t, err)
	concurrentIP := "2001:db8::40"
	oldSnapshot, err := repo.Snapshot(ctx, concurrentIP)
	require.NoError(t, err)
	results := make(chan error, 8)
	var workers sync.WaitGroup
	for range 8 {
		workers.Go(func() {
			_, addErr := guard.ManualBlock(ctx, service.AuthIPManualBlock{IP: concurrentIP, BlockMode: service.AuthIPBlockModePermanent, Reason: "并发人工封禁"})
			results <- addErr
		})
	}
	workers.Wait()
	close(results)
	successes := 0
	for addErr := range results {
		if addErr == nil {
			successes++
		} else {
			require.True(t, errors.Is(addErr, service.ErrAuthIPAlreadyBlocked), "%v", addErr)
		}
	}
	require.Equal(t, 1, successes)
	items, total, err = repo.List(ctx, concurrentIP, 1, 20)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	manual := items[0]
	events := service.AuthIPWindowCount{Count: config.Threshold, Now: time.Now(), First: time.Now(), Last: time.Now()}
	decision, err := repo.Block(ctx, concurrentIP, "/api/v1/auth/login", *oldSnapshot, events)
	require.NoError(t, err)
	require.False(t, decision.NewlyBlocked)
	require.Equal(t, "manual", decision.Entry.Source)
	_, err = guard.Unblock(ctx, manual.ID, manual.Generation)
	require.NoError(t, err)
	decision, err = repo.Block(ctx, concurrentIP, "/api/v1/auth/login", *oldSnapshot, events)
	require.NoError(t, err)
	require.Nil(t, decision.Entry, "人工解封后旧在途请求不能重新拉黑")
	freshSnapshot, err := repo.Snapshot(ctx, concurrentIP)
	require.NoError(t, err)
	decision, err = repo.Block(ctx, concurrentIP, "/api/v1/auth/login", *freshSnapshot, events)
	require.NoError(t, err)
	require.True(t, decision.NewlyBlocked)
	require.Equal(t, "auto", decision.Entry.Source)
	require.Empty(t, decision.Entry.Reason, "重新自动封禁不得沿用旧人工原因")
	require.NotNil(t, decision.Entry.FirstEventAt)
}
