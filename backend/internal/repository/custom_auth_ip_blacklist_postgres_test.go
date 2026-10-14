//go:build unit

package repository

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestAuthIPBlacklistRealPostgres(t *testing.T) {
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
	config, err := repo.Config(ctx)
	require.NoError(t, err)
	require.False(t, config.Enabled)
	require.Equal(t, 10, config.Threshold)
	require.Equal(t, 7200, config.WindowSeconds)
	require.Equal(t, service.AuthIPBlockModeTimed, config.BlockMode)
	config.Enabled = true
	config, err = repo.UpdateConfig(ctx, *config)
	require.NoError(t, err)
	_, err = repo.UpdateConfig(ctx, service.AuthIPBlacklistConfig{Enabled: true, Threshold: 10, WindowSeconds: 7200, Version: 1, BlockMode: service.AuthIPBlockModeTimed})
	require.ErrorIs(t, err, service.ErrAuthIPBlacklistConflict)
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	counter := NewAuthIPWindowCounter(client)
	guard := service.NewAuthIPBlacklistService(repo, counter)
	var decision *service.AuthIPBlacklistDecision
	for index, clock := range authIPReplayTimes {
		server.SetTime(authIPReplayInstant(t, clock))
		path := "/api/v1/auth/login"
		if index%2 == 1 {
			path = "/api/v1/auth/register"
		}
		decision, err = guard.Check(ctx, "61.120.9.25", path, true)
		if index < 9 {
			require.NoError(t, err)
		} else {
			require.ErrorIs(t, err, service.ErrAuthIPBlocked)
		}
	}
	require.True(t, decision.NewlyBlocked)
	entry := decision.Entry
	require.Equal(t, authIPReplayInstant(t, "20:32:32"), entry.BlockedAt)
	require.Equal(t, authIPReplayInstant(t, "19:12:39"), *entry.FirstEventAt)
	require.Equal(t, 10, entry.TriggerCount)
	items, total, err := repo.List(ctx, "61.120.9.25", 1, 20)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, items, 1)
	items, total, err = repo.List(ctx, "192.0.2.10", 1, 20)
	require.NoError(t, err)
	require.Zero(t, total)
	require.Empty(t, items)
	server.FlushAll()
	restarted := service.NewAuthIPBlacklistService(NewAuthIPBlacklistRepository(db), nil)
	_, err = restarted.Check(ctx, "61.120.9.25", "/api/v1/auth/oauth/google/start", false)
	require.ErrorIs(t, err, service.ErrAuthIPBlocked, "重启且 Redis 清空仍封禁")
	oldSnapshot, err := repo.Snapshot(ctx, "61.120.9.25")
	require.NoError(t, err)
	oldSnapshot.Entry = nil
	released, err := guard.Unblock(ctx, entry.ID, entry.Generation)
	require.NoError(t, err)
	require.False(t, released.Blocked)
	require.Greater(t, released.Generation, entry.Generation)
	_, err = guard.Unblock(ctx, entry.ID, entry.Generation)
	require.ErrorIs(t, err, service.ErrAuthIPBlacklistConflict)
	oldEvents := service.AuthIPWindowCount{Count: 10, Now: *entry.LastEventAt, First: *entry.FirstEventAt, Last: *entry.LastEventAt}
	stale, err := repo.Block(ctx, entry.IP, "/api/v1/auth/login", *oldSnapshot, oldEvents)
	require.NoError(t, err)
	require.Nil(t, stale.Entry, "旧在途请求不能重新封禁")
	for attempt := range 9 {
		_, err = guard.Check(ctx, entry.IP, "/api/v1/auth/login", true)
		require.NoError(t, err, "解封后第 %d 次", attempt+1)
	}
	decision, err = guard.Check(ctx, entry.IP, "/api/v1/auth/login", true)
	require.ErrorIs(t, err, service.ErrAuthIPBlocked)
	require.True(t, decision.NewlyBlocked)
	config.Enabled = false
	config, err = guard.UpdateConfig(ctx, *config)
	require.NoError(t, err)
	_, err = guard.Check(ctx, entry.IP, "/api/v1/auth/login", true)
	require.ErrorIs(t, err, service.ErrAuthIPBlocked, "关闭开关仍检查已有名单")
	_, err = guard.Check(ctx, "192.0.2.10", "/api/v1/auth/login", true)
	require.NoError(t, err)
	config.Enabled = true
	config, err = guard.UpdateConfig(ctx, *config)
	require.NoError(t, err)

	snapshot, err := repo.Snapshot(ctx, "192.0.2.20")
	require.NoError(t, err)
	config, err = guard.UpdateConfig(ctx, *config)
	require.NoError(t, err)
	stale, err = repo.Block(ctx, "192.0.2.20", "/api/v1/auth/login", *snapshot, oldEvents)
	require.NoError(t, err)
	require.Nil(t, stale.Entry, "规则修改后旧请求不能封禁")
	snapshot, err = repo.Snapshot(ctx, "192.0.2.30")
	require.NoError(t, err)
	var wait sync.WaitGroup
	var newBlocks atomic.Int64
	failures := make(chan error, 30)
	for range 30 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, blockErr := repo.Block(ctx, "192.0.2.30", "/api/v1/auth/login", *snapshot, oldEvents)
			if blockErr != nil {
				failures <- blockErr
				return
			}
			if result.NewlyBlocked {
				newBlocks.Add(1)
			}
		}()
	}
	wait.Wait()
	close(failures)
	for failure := range failures {
		require.NoError(t, failure)
	}
	require.EqualValues(t, 1, newBlocks.Load(), "跨实例并发只能创建一次封禁")
	snapshot, err = repo.Snapshot(ctx, "192.0.2.30")
	require.NoError(t, err)
	failures = make(chan error, 30)
	for range 30 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, blockErr := repo.Block(ctx, "192.0.2.30", "/api/v1/auth/login", *snapshot, oldEvents)
			if blockErr != nil {
				failures <- blockErr
			}
		}()
	}
	_, err = repo.Unblock(ctx, snapshot.Entry.ID, snapshot.Entry.Generation)
	require.NoError(t, err)
	wait.Wait()
	close(failures)
	for failure := range failures {
		require.NoError(t, failure)
	}
	after, err := repo.Snapshot(ctx, "192.0.2.30")
	require.NoError(t, err)
	require.False(t, after.Entry.Blocked, "解封与旧封禁请求竞争后仍保持解封")
	current, err := repo.Snapshot(ctx, entry.IP)
	require.NoError(t, err)
	require.True(t, current.Entry.Blocked)
	require.NotNil(t, current.Entry.BlockedUntil)
	_, err = db.ExecContext(ctx, `UPDATE custom_auth_ip_blacklist SET blocked_until = NOW() - INTERVAL '1 second' WHERE ip = $1::inet`, entry.IP)
	require.NoError(t, err)
	items, _, err = repo.List(ctx, entry.IP, 1, 20)
	require.NoError(t, err)
	require.Empty(t, items, "过期记录不再出现在活跃名单")
	expired, err := repo.Snapshot(ctx, entry.IP)
	require.NoError(t, err)
	require.False(t, expired.Entry.Blocked)
	require.Greater(t, expired.Entry.Generation, current.Entry.Generation)
	stale, err = repo.Block(ctx, entry.IP, "/api/v1/auth/login", *current, oldEvents)
	require.NoError(t, err)
	require.Nil(t, stale.Entry, "到期后旧请求不能重封")
	_, err = guard.Check(ctx, entry.IP, "/api/v1/auth/login", true)
	require.NoError(t, err, "自动到期后从新窗口重新计数")
	_, err = db.ExecContext(ctx, `UPDATE custom_auth_ip_blacklist SET blocked = TRUE, blocked_until = NULL WHERE ip = $1::inet`, entry.IP)
	require.NoError(t, err)
	_, err = guard.Check(ctx, entry.IP, "/api/v1/auth/login", false)
	require.ErrorIs(t, err, service.ErrAuthIPBlocked, "旧永久记录保留管理员解封要求")

	config.Threshold = 1
	config.BlockMode = service.AuthIPBlockModePermanent
	previousConfig := *config
	config, err = guard.UpdateConfig(ctx, *config)
	require.NoError(t, err)
	require.Equal(t, service.AuthIPBlockModePermanent, config.BlockMode)
	_, err = guard.UpdateConfig(ctx, previousConfig)
	require.ErrorIs(t, err, service.ErrAuthIPBlacklistConflict, "旧版本不能覆盖封禁模式")
	observation, _, err := guard.Begin(ctx, "192.0.2.40", "/api/v1/auth/login", true)
	require.NoError(t, err)
	decision, err = guard.Failed(ctx, observation)
	require.ErrorIs(t, err, service.ErrAuthIPBlocked)
	require.True(t, decision.NewlyBlocked)
	permanent := decision.Entry
	require.Nil(t, permanent.BlockedUntil, "永久模式写入真正的无到期时间记录")
	server.FlushAll()
	server.SetTime(time.Now().Add(48 * time.Hour))
	_, err = restarted.Check(ctx, permanent.IP, "/api/v1/auth/login", false)
	require.ErrorIs(t, err, service.ErrAuthIPBlocked, "永久封禁在重启、计数清空与两天后仍生效")
	config.BlockMode = service.AuthIPBlockModeTimed
	config.BlockSeconds = 60
	config, err = guard.UpdateConfig(ctx, *config)
	require.NoError(t, err)
	unchanged, err := repo.Snapshot(ctx, permanent.IP)
	require.NoError(t, err)
	require.True(t, unchanged.Entry.Blocked)
	require.Nil(t, unchanged.Entry.BlockedUntil, "切换规则不修改已有永久记录")
	_, err = guard.Unblock(ctx, permanent.ID, permanent.Generation)
	require.NoError(t, err)
	stale, err = repo.Block(ctx, permanent.IP, "/api/v1/auth/login", observation.Snapshot, oldEvents)
	require.NoError(t, err)
	require.Nil(t, stale.Entry)
	decision, err = guard.Check(ctx, permanent.IP, "/api/v1/auth/login", true)
	require.ErrorIs(t, err, service.ErrAuthIPBlocked)
	require.True(t, decision.NewlyBlocked)
	require.NotNil(t, decision.Entry.BlockedUntil, "解封后新请求采用临时模式")
	require.WithinDuration(t, time.Now().Add(time.Minute), *decision.Entry.BlockedUntil, 5*time.Second)
	config.BlockMode = service.AuthIPBlockModePermanent
	_, err = guard.UpdateConfig(ctx, *config)
	require.NoError(t, err)
	unchanged, err = repo.Snapshot(ctx, permanent.IP)
	require.NoError(t, err)
	require.Equal(t, decision.Entry.BlockedUntil, unchanged.Entry.BlockedUntil, "切回永久不延长已存在的临时封禁")
}
