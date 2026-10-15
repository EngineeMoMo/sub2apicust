//go:build unit

package service

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type customAccessCountingStore struct {
	*customAccessTestStore
	lists      atomic.Int64
	factsLoads atomic.Int64
}

func (s *customAccessCountingStore) List(ctx context.Context) (map[int64]CustomGroupAccessPolicy, error) {
	s.lists.Add(1)
	policies, err := s.customAccessTestStore.List(ctx)
	copy := make(map[int64]CustomGroupAccessPolicy, len(policies))
	for id, policy := range policies {
		copy[id] = policy
	}
	return copy, err
}

func (s *customAccessCountingStore) Facts(ctx context.Context, userID int64, start, end time.Time) (customGroupAccessFacts, error) {
	s.factsLoads.Add(1)
	return s.customAccessTestStore.Facts(ctx, userID, start, end)
}

func TestCustomGroupAccessCacheWarmRequests(t *testing.T) {
	for _, mode := range []string{"no-rules", "balance-only", "monthly-and-balance"} {
		t.Run(mode, func(t *testing.T) {
			svc, store, key := customAccessTestFixture()
			if mode != "no-rules" {
				store.policies[1] = CustomGroupAccessPolicy{GroupID: 1, BalanceEnabled: true, BalanceMin: 80, MonthlyRechargeEnabled: mode == "monthly-and-balance", MonthlyRechargeMin: 80}
			}
			counting := &customAccessCountingStore{customAccessTestStore: store}
			svc.store, svc.cache = counting, newCustomGroupAccessCache()
			svc.cache.health(true)
			for request := 0; request < 100; request++ {
				_, err := svc.PrepareKey(context.Background(), key)
				require.NoError(t, err)
			}
			require.EqualValues(t, 1, counting.lists.Load())
			if mode == "monthly-and-balance" {
				require.EqualValues(t, 1, counting.factsLoads.Load())
			} else {
				require.Zero(t, counting.factsLoads.Load())
			}
		})
	}
}

func TestCustomGroupAccessCacheInvalidValues(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	start, _ := customGroupAccessMonth(now)
	monthlyKey := customGroupAccessCacheKey("monthly", 1) + ":" + strconv.FormatInt(start.Unix(), 10)
	cases := []struct {
		name  string
		key   string
		value any
		load  func(*CustomGroupAccessService) error
	}{
		{"policies-type", "policies", "invalid", func(svc *CustomGroupAccessService) error {
			_, err := svc.loadPolicies(context.Background())
			return err
		}},
		{"user-type", "user:1", "invalid", func(svc *CustomGroupAccessService) error { _, err := svc.loadUser(context.Background(), 1); return err }},
		{"user-nil", "user:1", (*User)(nil), func(svc *CustomGroupAccessService) error { _, err := svc.loadUser(context.Background(), 1); return err }},
		{"group-type", "group:1", "invalid", func(svc *CustomGroupAccessService) error {
			_, err := svc.loadGroup(context.Background(), 1)
			return err
		}},
		{"group-nil", "group:1", (*Group)(nil), func(svc *CustomGroupAccessService) error {
			_, err := svc.loadGroup(context.Background(), 1)
			return err
		}},
		{"monthly-type", monthlyKey, "invalid", func(svc *CustomGroupAccessService) error {
			_, err := svc.loadFacts(context.Background(), 1, true)
			return err
		}},
		{"subscription-type", "subscription:1:1", "invalid", func(svc *CustomGroupAccessService) error {
			_, err := svc.loadSubscription(context.Background(), 1, 1)
			return err
		}},
		{"subscription-nil", "subscription:1:1", (*UserSubscription)(nil), func(svc *CustomGroupAccessService) error {
			_, err := svc.loadSubscription(context.Background(), 1, 1)
			return err
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			svc, _, _ := customAccessTestFixture()
			svc.now = func() time.Time { return now }
			svc.cache = newCustomGroupAccessCache()
			svc.cache.health(true)
			svc.cache.entries[test.key] = &customGroupAccessCacheEntry{value: test.value, expires: now.Add(time.Minute)}
			require.Error(t, test.load(svc))
		})
	}
}

func TestCustomGroupAccessCacheBalanceAndRefundInvalidation(t *testing.T) {
	svc, store, key := customAccessTestFixture()
	store.policies[1] = CustomGroupAccessPolicy{GroupID: 1, BalanceEnabled: true, BalanceMin: 80, MonthlyRechargeEnabled: true, MonthlyRechargeMin: 80}
	counting := &customAccessCountingStore{customAccessTestStore: store}
	svc.store, svc.cache = counting, newCustomGroupAccessCache()
	svc.cache.health(true)
	_, err := svc.PrepareKey(context.Background(), key)
	require.NoError(t, err)
	require.NoError(t, svc.cache.changed(`{"kind":"balance","id":7,"balance":50}`))
	_, err = svc.PrepareKey(context.Background(), key)
	require.ErrorIs(t, err, ErrGroupAccessRequirements)
	require.EqualValues(t, 1, counting.factsLoads.Load(), "消费不应重新汇总月充值")
	require.Equal(t, float64(100), key.User.Balance, "认证缓存原对象不得被修改")
	require.NoError(t, svc.cache.changed(`{"kind":"balance","id":7,"balance":100}`))
	store.facts.MonthlyRecharge = 0
	require.NoError(t, svc.cache.changed(`{"kind":"monthly","id":7}`))
	_, err = svc.PrepareKey(context.Background(), key)
	require.ErrorIs(t, err, ErrGroupAccessRequirements)
	require.EqualValues(t, 2, counting.factsLoads.Load())
}

func TestCustomGroupAccessCacheConcurrentMissAndInvalidation(t *testing.T) {
	t.Run("coalesce", func(t *testing.T) {
		cache := newCustomGroupAccessCache()
		cache.health(true)
		var loads atomic.Int64
		var wait sync.WaitGroup
		results := make(chan error, 100)
		start := make(chan struct{})
		for request := 0; request < 100; request++ {
			wait.Add(1)
			go func() {
				defer wait.Done()
				<-start
				_, err := cache.get(context.Background(), "monthly:7", time.Now(), time.Now().Add(time.Minute), func(context.Context) (any, error) { loads.Add(1); return float64(100), nil })
				results <- err
			}()
		}
		close(start)
		wait.Wait()
		close(results)
		for err := range results {
			require.NoError(t, err)
		}
		require.EqualValues(t, 1, loads.Load())
	})
	t.Run("refund-during-load", func(t *testing.T) {
		cache := newCustomGroupAccessCache()
		cache.health(true)
		started, release := make(chan struct{}), make(chan struct{})
		var loads atomic.Int64
		result := make(chan any, 1)
		go func() {
			value, err := cache.get(context.Background(), "monthly:7", time.Now(), time.Now().Add(time.Minute), func(context.Context) (any, error) {
				if loads.Add(1) == 1 {
					close(started)
					<-release
					return float64(100), nil
				}
				return float64(0), nil
			})
			if err != nil {
				result <- err
			} else {
				result <- value
			}
		}()
		<-started
		require.NoError(t, cache.changed(`{"kind":"monthly","id":7}`))
		close(release)
		require.Equal(t, float64(0), <-result)
		require.EqualValues(t, 2, loads.Load())
	})
}

func TestCustomGroupAccessCacheDisconnectFailsClosed(t *testing.T) {
	cache := newCustomGroupAccessCache()
	cache.health(true)
	now := time.Now()
	_, err := cache.get(context.Background(), "monthly:7", now, now.Add(time.Minute), func(context.Context) (any, error) { return 100, nil })
	require.NoError(t, err)
	cache.health(false)
	_, err = cache.get(context.Background(), "monthly:7", now, now.Add(time.Minute), func(context.Context) (any, error) { return nil, errors.New("database unavailable") })
	require.ErrorContains(t, err, "database unavailable")
	cache.health(true)
	value, err := cache.get(context.Background(), "monthly:7", now, now.Add(time.Minute), func(context.Context) (any, error) { return 0, nil })
	require.NoError(t, err)
	require.Equal(t, 0, value)
}

func TestCustomGroupAccessCacheMonthAndExpiryBoundaries(t *testing.T) {
	svc, store, key := customAccessTestFixture()
	now := time.Date(2026, 10, 31, 15, 59, 59, 0, time.UTC)
	svc.now = func() time.Time { return now }
	svc.cache = newCustomGroupAccessCache()
	svc.cache.health(true)
	store.policies[1] = CustomGroupAccessPolicy{GroupID: 1, MonthlyRechargeEnabled: true, MonthlyRechargeMin: 80}
	_, err := svc.PrepareKey(context.Background(), key)
	require.NoError(t, err)
	store.facts.MonthlyRecharge = 0
	now = now.Add(time.Second)
	_, err = svc.PrepareKey(context.Background(), key)
	require.ErrorIs(t, err, ErrGroupAccessRequirements)
	expires := now.Add(time.Second)
	fallback := int64(2)
	store.policies[1] = CustomGroupAccessPolicy{GroupID: 1, ExpiresAt: &expires, FallbackGroupID: &fallback}
	require.NoError(t, svc.cache.changed(`{"kind":"policies"}`))
	prepared, err := svc.PrepareKey(context.Background(), key)
	require.NoError(t, err)
	require.Equal(t, int64(1), *prepared.GroupID)
	now = expires
	prepared, err = svc.PrepareKey(context.Background(), key)
	require.NoError(t, err)
	require.Equal(t, fallback, *prepared.GroupID)
}

func BenchmarkCustomGroupAccessCacheWarm(b *testing.B) {
	svc, store, key := customAccessTestFixture()
	store.policies[1] = CustomGroupAccessPolicy{GroupID: 1, MonthlyRechargeEnabled: true, MonthlyRechargeMin: 80, BalanceEnabled: true, BalanceMin: 80}
	counting := &customAccessCountingStore{customAccessTestStore: store}
	svc.store, svc.cache = counting, newCustomGroupAccessCache()
	svc.cache.health(true)
	if _, err := svc.PrepareKey(context.Background(), key); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for request := 0; request < b.N; request++ {
		if _, err := svc.PrepareKey(context.Background(), key); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(counting.factsLoads.Load()-1)/float64(b.N), "monthly-loads/op")
	b.ReportMetric(float64(counting.lists.Load()-1)/float64(b.N), "policy-loads/op")
}
