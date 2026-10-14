//go:build unit

package repository

import (
	"context"
	"database/sql"
	"fmt"
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

func TestAuthIPNetworkRealPostgres(t *testing.T) {
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
	var defaultMode string
	require.NoError(t, db.QueryRowContext(ctx, `SELECT column_default FROM information_schema.columns
WHERE table_schema = 'public' AND table_name = 'custom_auth_ip_blacklist_config' AND column_name = 'block_mode'`).Scan(&defaultMode))
	require.Contains(t, defaultMode, "permanent")
	address := os.Getenv("AUTH_IP_TEST_REDIS_ADDR")
	require.NotEmpty(t, address, "network regression requires isolated real Redis")
	client := redis.NewClient(&redis.Options{Addr: address})
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.Ping(ctx).Err())
	repo := NewAuthIPBlacklistRepository(db)
	guard := service.NewAuthIPBlacklistService(repo, NewAuthIPWindowCounter(client))
	reset := func(threshold int) {
		require.NoError(t, client.FlushDB(ctx).Err())
		_, err := db.ExecContext(ctx, `TRUNCATE custom_auth_ip_blacklist RESTART IDENTITY`)
		require.NoError(t, err)
		_, err = db.ExecContext(ctx, `UPDATE custom_auth_ip_blacklist_config SET enabled = TRUE, threshold = $1, window_seconds = 7200,
block_seconds = 900, block_mode = 'permanent', aggregation_mode = 'network', network_ipv4_prefix = 24,
network_ipv6_prefix = 64, version = version + 1 WHERE id = 1`, threshold)
		require.NoError(t, err)
	}
	fail := func(ip, path string) (*service.AuthIPBlacklistDecision, error) {
		observation, _, err := guard.Begin(ctx, ip, path, true)
		if err != nil {
			return nil, err
		}
		return guard.Failed(ctx, observation)
	}

	t.Run("不同IP共享计数成功不计入邻网段不受影响", func(t *testing.T) {
		reset(3)
		for range 12 {
			observation, _, err := guard.Begin(ctx, "210.230.206.1", "/api/v1/auth/login", true)
			require.NoError(t, err)
			require.NotNil(t, observation)
		}
		_, err := fail("210.230.207.1", "/api/v1/auth/login")
		require.NoError(t, err)
		_, err = fail("210.230.206.27", "/api/v1/auth/login")
		require.NoError(t, err)
		_, err = fail("210.230.206.23", "/api/v1/auth/register")
		require.NoError(t, err)
		decision, err := fail("210.230.206.18", "/api/v1/auth/send-email-code")
		require.ErrorIs(t, err, service.ErrAuthIPBlocked)
		require.True(t, decision.NewlyBlocked)
		require.Equal(t, "210.230.206.0/24", decision.Entry.IP)
		require.Equal(t, 3, decision.Entry.TriggerCount)
		_, _, err = guard.Begin(ctx, "210.230.206.254", "/api/v1/auth/login", true)
		require.ErrorIs(t, err, service.ErrAuthIPBlocked)
		_, _, err = guard.Begin(ctx, "210.230.207.254", "/api/v1/auth/login", true)
		require.NoError(t, err)
		require.NoError(t, client.FlushDB(ctx).Err())
		restarted := service.NewAuthIPBlacklistService(NewAuthIPBlacklistRepository(db), nil)
		_, _, err = restarted.Begin(ctx, "210.230.206.254", "/api/v1/auth/login", true)
		require.ErrorIs(t, err, service.ErrAuthIPBlocked)
		_, err = guard.Unblock(ctx, decision.Entry.ID, decision.Entry.Generation)
		require.NoError(t, err)
		_, err = fail("210.230.206.26", "/api/v1/auth/login")
		require.NoError(t, err)
		items, total, err := repo.List(ctx, "210.230.206.26", 1, 20)
		require.NoError(t, err)
		require.Empty(t, items)
		require.Zero(t, total)
	})

	t.Run("同网段并发只持久封禁一次", func(t *testing.T) {
		reset(3)
		var observations []*service.AuthIPObservation
		for index := 1; index <= 24; index++ {
			observation, _, err := guard.Begin(ctx, fmt.Sprintf("192.0.2.%d", index), "/api/v1/auth/login", true)
			require.NoError(t, err)
			observations = append(observations, observation)
		}
		var wait sync.WaitGroup
		var newly atomic.Int64
		errors := make(chan error, len(observations))
		for _, observation := range observations {
			wait.Go(func() {
				decision, err := guard.Failed(ctx, observation)
				if err != nil && err != service.ErrAuthIPBlocked {
					errors <- err
				}
				if decision != nil && decision.NewlyBlocked {
					newly.Add(1)
				}
			})
		}
		wait.Wait()
		close(errors)
		for err := range errors {
			require.NoError(t, err)
		}
		require.EqualValues(t, 1, newly.Load())
		items, total, err := guard.List(ctx, "192.0.2.*", 1, 20)
		require.NoError(t, err)
		require.EqualValues(t, 1, total)
		require.Len(t, items, 1)
	})

	t.Run("手动通配符覆盖匹配及旧请求解封竞争", func(t *testing.T) {
		reset(1)
		observation, _, err := guard.Begin(ctx, "198.51.100.18", "/api/v1/auth/login", true)
		require.NoError(t, err)
		manual, err := guard.ManualBlock(ctx, service.AuthIPManualBlock{IP: "198.51.100.*", BlockMode: "permanent", Reason: "人工确认"})
		require.NoError(t, err)
		require.Equal(t, "198.51.100.0/24", manual.IP)
		_, err = guard.Unblock(ctx, manual.ID, manual.Generation)
		require.NoError(t, err)
		decision, err := guard.Failed(ctx, observation)
		require.NoError(t, err)
		require.Nil(t, decision.Entry)
		fresh, _, err := guard.Begin(ctx, "198.51.100.19", "/api/v1/auth/login", true)
		require.NoError(t, err)
		_, err = guard.Failed(ctx, fresh)
		require.ErrorIs(t, err, service.ErrAuthIPBlocked)
		_, err = guard.Unblock(ctx, manual.ID, manual.Generation)
		require.ErrorIs(t, err, service.ErrAuthIPBlacklistConflict)
	})

	t.Run("窄规则变更使网段旧统计失效且重叠规则独立", func(t *testing.T) {
		reset(1)
		observation, _, err := guard.Begin(ctx, "198.51.100.18", "/api/v1/auth/login", true)
		require.NoError(t, err)
		narrow, err := guard.ManualBlock(ctx, service.AuthIPManualBlock{IP: "198.51.100.200", BlockMode: "permanent", Reason: "独立恶意地址"})
		require.NoError(t, err)
		decision, err := guard.Failed(ctx, observation)
		require.NoError(t, err)
		require.Nil(t, decision.Entry)
		broad, err := guard.ManualBlock(ctx, service.AuthIPManualBlock{IP: "198.51.*.*", BlockMode: "permanent", Reason: "人工网段"})
		require.NoError(t, err)
		_, err = guard.ManualBlock(ctx, service.AuthIPManualBlock{IP: "198.51.100.*", BlockMode: "permanent", Reason: "重复覆盖"})
		require.ErrorIs(t, err, service.ErrAuthIPAlreadyBlocked)
		items, total, err := repo.List(ctx, "198.51.100.200", 1, 20)
		require.NoError(t, err)
		require.EqualValues(t, 2, total)
		require.Len(t, items, 2)
		_, err = guard.Unblock(ctx, broad.ID, broad.Generation)
		require.NoError(t, err)
		_, _, err = guard.Begin(ctx, narrow.IP, "/api/v1/auth/login", true)
		require.ErrorIs(t, err, service.ErrAuthIPBlocked)
		_, _, err = guard.Begin(ctx, "198.51.100.18", "/api/v1/auth/login", true)
		require.NoError(t, err)
	})

	t.Run("升级前自动单IP名单解封重置新的网段计数", func(t *testing.T) {
		reset(1)
		config, err := repo.Config(ctx)
		require.NoError(t, err)
		config.AggregationMode = "ip"
		config, err = guard.UpdateConfig(ctx, *config)
		require.NoError(t, err)
		host, err := fail("192.0.2.5", "/api/v1/auth/login")
		require.ErrorIs(t, err, service.ErrAuthIPBlocked)
		require.Equal(t, "192.0.2.5", host.Entry.IP)
		config.AggregationMode = "network"
		config.Threshold = 3
		_, err = guard.UpdateConfig(ctx, *config)
		require.NoError(t, err)
		_, err = fail("192.0.2.6", "/api/v1/auth/login")
		require.NoError(t, err)
		_, err = fail("192.0.2.7", "/api/v1/auth/login")
		require.NoError(t, err)
		_, err = guard.Unblock(ctx, host.Entry.ID, host.Entry.Generation)
		require.NoError(t, err)
		_, err = fail("192.0.2.5", "/api/v1/auth/login")
		require.NoError(t, err, "旧单IP名单解封后网段须从第一次失败重新累计")
	})

	t.Run("IPv6网段和临时到期重置", func(t *testing.T) {
		reset(2)
		config, err := repo.Config(ctx)
		require.NoError(t, err)
		config.BlockMode = "timed"
		_, err = guard.UpdateConfig(ctx, *config)
		require.NoError(t, err)
		_, err = fail("2001:db8:1:2::1", "/api/v1/auth/login")
		require.NoError(t, err)
		decision, err := fail("2001:db8:1:2::ffff", "/api/v1/auth/register")
		require.ErrorIs(t, err, service.ErrAuthIPBlocked)
		require.Equal(t, "2001:db8:1:2::/64", decision.Entry.IP)
		_, _, err = guard.Begin(ctx, "2001:db8:1:3::ffff", "/api/v1/auth/login", true)
		require.NoError(t, err)
		_, err = db.ExecContext(ctx, `UPDATE custom_auth_ip_blacklist SET blocked_until = NOW() - INTERVAL '1 second' WHERE id = $1`, decision.Entry.ID)
		require.NoError(t, err)
		_, err = fail("2001:db8:1:2::abcd", "/api/v1/auth/login")
		require.NoError(t, err, "到期后的第一次失败从新窗口开始")
	})

	t.Run("切换范围旧请求失效单IP分别累计", func(t *testing.T) {
		reset(2)
		observation, _, err := guard.Begin(ctx, "203.0.113.1", "/api/v1/auth/login", true)
		require.NoError(t, err)
		config, err := repo.Config(ctx)
		require.NoError(t, err)
		config.AggregationMode = "ip"
		config.Threshold = 1
		_, err = guard.UpdateConfig(ctx, *config)
		require.NoError(t, err)
		decision, err := guard.Failed(ctx, observation)
		require.NoError(t, err)
		require.Nil(t, decision)
		decision, err = fail("203.0.113.1", "/api/v1/auth/login")
		require.ErrorIs(t, err, service.ErrAuthIPBlocked)
		require.Equal(t, "203.0.113.1", decision.Entry.IP)
		_, _, err = guard.Begin(ctx, "203.0.113.2", "/api/v1/auth/login", true)
		require.NoError(t, err)
	})

	t.Run("截图时间点不同IP回放", func(t *testing.T) {
		reset(10)
		server := miniredis.RunT(t)
		mockClient := redis.NewClient(&redis.Options{Addr: server.Addr()})
		t.Cleanup(func() { _ = mockClient.Close() })
		replay := service.NewAuthIPBlacklistService(repo, NewAuthIPWindowCounter(mockClient))
		for index, clock := range authIPReplayTimes {
			server.SetTime(authIPReplayInstant(t, clock))
			decision, err := replay.Check(ctx, fmt.Sprintf("61.120.9.%d", index+1), "/api/v1/auth/login", true)
			if index < 9 {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, service.ErrAuthIPBlocked)
				require.Equal(t, "61.120.9.0/24", decision.Entry.IP)
				require.Equal(t, authIPReplayInstant(t, "20:32:32"), decision.Entry.BlockedAt)
				require.Equal(t, 79*time.Minute+53*time.Second, decision.Entry.LastEventAt.Sub(*decision.Entry.FirstEventAt))
			}
		}
	})
}
