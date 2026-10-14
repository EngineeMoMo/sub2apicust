//go:build unit

package repository

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

var authIPReplayTimes = []string{"19:12:39", "19:19:01", "19:35:39", "19:43:23", "19:57:26", "20:05:11", "20:19:10", "20:21:41", "20:31:28", "20:32:32"}

func authIPReplayInstant(t *testing.T, clock string) time.Time {
	t.Helper()
	instant, err := time.Parse(time.RFC3339, "2026-10-13T"+clock+"+08:00")
	require.NoError(t, err)
	return instant.UTC()
}

func TestAuthIPCounterScreenshotAndBoundaries(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	counter := NewAuthIPWindowCounter(client)
	config := service.AuthIPBlacklistConfig{Enabled: true, Threshold: 10, WindowSeconds: 7200, Version: 1}
	ctx := context.Background()
	for index, clock := range authIPReplayTimes {
		server.SetTime(authIPReplayInstant(t, clock))
		count, err := counter.Count(ctx, "61.120.9.25", 0, config)
		require.NoError(t, err)
		require.Equal(t, index+1, count.Count)
		require.Equal(t, authIPReplayInstant(t, "19:12:39"), count.First)
		require.Equal(t, authIPReplayInstant(t, clock), count.Last)
	}
	server.SetTime(authIPReplayInstant(t, "21:12:39"))
	count, err := counter.Count(ctx, "61.120.9.25", 0, config)
	require.NoError(t, err)
	require.Equal(t, 10, count.Count)
	require.Equal(t, authIPReplayInstant(t, "19:19:01"), count.First)
	require.Equal(t, 2*time.Hour, server.TTL(service.AuthIPCounterKey("61.120.9.25", 0, 1)))
	require.NoError(t, client.ZAdd(ctx, service.AuthIPCounterKey("192.0.2.99", 0, 1), redis.Z{Score: float64(authIPReplayInstant(t, "22:00:00").UnixMilli()), Member: "future"}).Err())
	count, err = counter.Count(ctx, "192.0.2.99", 0, config)
	require.NoError(t, err)
	require.Equal(t, 1, count.Count, "时钟回拨时未来事件不属于窗口")
	require.Equal(t, count.Now, count.First)
	require.Equal(t, count.Now, count.Last)
	for repeat := range 15 {
		count, err = counter.Count(ctx, "192.0.2.5", 0, config)
		require.NoError(t, err)
		require.Equal(t, repeat+1, count.Count, "同毫秒请求均保留独立事件")
	}
	count, err = counter.Count(ctx, "61.120.9.25", 2, config)
	require.NoError(t, err)
	require.Equal(t, 1, count.Count, "解封状态版本重置计数")
	config.Version++
	count, err = counter.Count(ctx, "192.0.2.5", 0, config)
	require.NoError(t, err)
	require.Equal(t, 1, count.Count, "规则版本重置计数")
	server.FastForward(2 * time.Hour)
	require.False(t, server.Exists(service.AuthIPCounterKey("61.120.9.25", 0, 1)))
}

func TestAuthIPCounterUsesEntryTimeForFailedRequests(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	counter := NewAuthIPWindowCounter(client).(service.AuthIPRequestCounter)
	config := service.AuthIPBlacklistConfig{Enabled: true, Threshold: 10, WindowSeconds: 7200, Version: 1}
	ctx := context.Background()
	server.SetTime(authIPReplayInstant(t, "19:12:39"))
	entered, err := counter.Time(ctx)
	require.NoError(t, err)
	server.SetTime(authIPReplayInstant(t, "19:19:01"))
	count, err := counter.CountAt(ctx, "192.0.2.1", 0, config, entered)
	require.NoError(t, err)
	require.Equal(t, entered, count.First)
	server.SetTime(authIPReplayInstant(t, "21:12:39"))
	count, err = counter.CountAt(ctx, "192.0.2.1", 0, config, entered)
	require.NoError(t, err)
	require.Zero(t, count.Count, "处理结束时恰好过期的入场事件不再计入")
}

func TestAuthIPCounterRealRedisConcurrency(t *testing.T) {
	address := os.Getenv("AUTH_IP_TEST_REDIS_ADDR")
	if address == "" {
		t.Skip("requires isolated Redis")
	}
	client := redis.NewClient(&redis.Options{Addr: address, DB: 3})
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	require.NoError(t, client.FlushDB(ctx).Err())
	counter := NewAuthIPWindowCounter(client)
	config := service.AuthIPBlacklistConfig{Enabled: true, Threshold: 1000, WindowSeconds: 7200, Version: 1}
	var wait sync.WaitGroup
	counts := make(chan int, 80)
	failures := make(chan error, 80)
	for range 80 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			count, err := counter.Count(ctx, "2001:db8::5", 0, config)
			if err != nil {
				failures <- err
				return
			}
			counts <- count.Count
		}()
	}
	wait.Wait()
	close(counts)
	close(failures)
	for err := range failures {
		require.NoError(t, err)
	}
	observed := map[int]bool{}
	for count := range counts {
		observed[count] = true
	}
	require.Len(t, observed, 80)
	key := service.AuthIPCounterKey("2001:db8::5", 0, 1)
	require.EqualValues(t, 80, client.ZCard(ctx, key).Val())
	require.Greater(t, client.PTTL(ctx, key).Val(), 119*time.Minute)
	clock, err := client.Time(ctx).Result()
	require.NoError(t, err)
	require.NoError(t, client.ZAdd(ctx, key, redis.Z{Score: float64(clock.Add(-2 * time.Hour).UnixMilli()), Member: "exact-boundary"}).Err())
	count, err := counter.Count(ctx, "2001:db8::5", 0, config)
	require.NoError(t, err)
	require.Equal(t, 81, count.Count)
	require.WithinDuration(t, clock, count.Now, 5*time.Second)
}
