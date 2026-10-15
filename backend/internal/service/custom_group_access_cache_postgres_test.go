//go:build unit

package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type customAccessSQLUsers struct {
	UserRepository
	db *sql.DB
}

func (r customAccessSQLUsers) GetByID(ctx context.Context, id int64) (*User, error) {
	user := &User{ID: id}
	err := r.db.QueryRowContext(ctx, "SELECT balance,status FROM users WHERE id=$1", id).Scan(&user.Balance, &user.Status)
	return user, err
}

type customAccessSQLCountStore struct {
	customGroupAccessStore
	factsLoads atomic.Int64
}

func (s *customAccessSQLCountStore) Facts(ctx context.Context, id int64, start, end time.Time) (customGroupAccessFacts, error) {
	s.factsLoads.Add(1)
	return s.customGroupAccessStore.Facts(ctx, id, start, end)
}

func TestCustomGroupAccessPostgresMultiInstanceCache(t *testing.T) {
	dsn := os.Getenv("CUSTOM_GROUP_ACCESS_TEST_DSN")
	if dsn == "" {
		t.Skip("CUSTOM_GROUP_ACCESS_TEST_DSN 未配置，未验证跨实例事务失效")
	}
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		var err error
		dsn, err = pq.ParseURL(dsn)
		require.NoError(t, err)
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	var name string
	require.NoError(t, db.QueryRow("SELECT current_database()").Scan(&name))
	require.Equal(t, "group_access_test", name)
	schema := fmt.Sprintf("group_access_cache_%d", time.Now().UnixNano())
	_, err = db.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = db.Exec("DROP SCHEMA " + schema + " CASCADE") })
	schemaDSN := dsn + " search_path=" + schema
	storeDB, err := sql.Open("postgres", schemaDSN)
	require.NoError(t, err)
	t.Cleanup(func() { _ = storeDB.Close() })
	_, err = storeDB.Exec(`CREATE TABLE groups(id BIGINT PRIMARY KEY,deleted_at TIMESTAMPTZ);
CREATE TABLE users(id BIGINT PRIMARY KEY,balance NUMERIC(20,8),status TEXT,deleted_at TIMESTAMPTZ);
CREATE TABLE payment_orders(id BIGINT PRIMARY KEY,user_id BIGINT,order_type TEXT,amount NUMERIC(20,2),bonus_amount NUMERIC(20,2),refund_amount NUMERIC(20,2),status TEXT,paid_at TIMESTAMPTZ);
CREATE TABLE user_allowed_groups(user_id BIGINT,group_id BIGINT);
CREATE TABLE user_group_rate_multipliers(user_id BIGINT,group_id BIGINT,rate_multiplier NUMERIC);
CREATE TABLE user_subscriptions(id BIGINT PRIMARY KEY,user_id BIGINT,group_id BIGINT);
INSERT INTO groups VALUES(1),(2),(3); INSERT INTO users VALUES(7,100,'active',NULL);
INSERT INTO payment_orders VALUES(1,7,'balance',100,0,0,'COMPLETED','2026-10-15T00:00:00Z');`)
	require.NoError(t, err)
	migration, err := os.ReadFile("../../migrations/257_custom_group_access_policy.sql")
	require.NoError(t, err)
	_, err = storeDB.Exec(string(migration))
	require.NoError(t, err)
	store := &customGroupAccessSQLStore{db: storeDB}
	_, err = store.Save(context.Background(), CustomGroupAccessPolicy{GroupID: 1, MonthlyRechargeEnabled: true, MonthlyRechargeMin: 80, BalanceEnabled: true, BalanceMin: 80})
	require.NoError(t, err)
	var services []*CustomGroupAccessService
	var counters []*customAccessSQLCountStore
	var keys []*APIKey
	var stops []func()
	for instance := 0; instance < 2; instance++ {
		svc, _, key := customAccessTestFixture()
		counter := &customAccessSQLCountStore{customGroupAccessStore: store}
		svc.store, svc.users, svc.cache = counter, customAccessSQLUsers{db: storeDB}, newCustomGroupAccessCache()
		stop := svc.startChangeListener(schemaDSN)
		t.Cleanup(stop)
		require.Eventually(t, func() bool {
			svc.cache.mu.Lock()
			defer svc.cache.mu.Unlock()
			return time.Now().Before(svc.cache.healthyUntil)
		}, 3*time.Second, 10*time.Millisecond)
		services = append(services, svc)
		counters = append(counters, counter)
		keys = append(keys, key)
		stops = append(stops, stop)
		_, err = svc.PrepareKey(context.Background(), key)
		require.NoError(t, err)
	}
	check := func(allowed bool) bool {
		for index, svc := range services {
			_, err := svc.PrepareKey(context.Background(), keys[index])
			if (err == nil) != allowed {
				return false
			}
		}
		return true
	}
	tx, err := storeDB.Begin()
	require.NoError(t, err)
	_, err = tx.Exec("UPDATE payment_orders SET status='REFUND_PROCESSING' WHERE id=1; UPDATE users SET balance=50 WHERE id=7")
	require.NoError(t, err)
	require.True(t, check(true), "事务提交前不能收到退款失效通知")
	require.NoError(t, tx.Commit())
	require.Eventually(t, func() bool { return check(false) }, 3*time.Second, 10*time.Millisecond)
	for _, counter := range counters {
		require.EqualValues(t, 2, counter.factsLoads.Load())
	}
	_, err = storeDB.Exec("UPDATE payment_orders SET status='COMPLETED' WHERE id=1; UPDATE users SET balance=100 WHERE id=7")
	require.NoError(t, err)
	require.Eventually(t, func() bool { return check(true) }, 3*time.Second, 10*time.Millisecond)
	tx, err = storeDB.Begin()
	require.NoError(t, err)
	_, err = tx.Exec("UPDATE users SET balance=0 WHERE id=7; UPDATE payment_orders SET status='REFUNDED' WHERE id=1")
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())
	require.True(t, check(true))
	previousCounts := []int64{counters[0].factsLoads.Load(), counters[1].factsLoads.Load()}
	_, err = storeDB.Exec("UPDATE users SET balance=50 WHERE id=7")
	require.NoError(t, err)
	require.Eventually(t, func() bool { return check(false) }, 3*time.Second, 10*time.Millisecond)
	for index, counter := range counters {
		require.Equal(t, previousCounts[index], counter.factsLoads.Load(), "消费只更新余额，不重新聚合月充值")
	}
	stops[0]()
	_, err = services[0].PrepareKey(context.Background(), keys[0])
	require.ErrorIs(t, err, ErrGroupAccessRequirements)
	_, err = storeDB.Exec("UPDATE users SET balance=100 WHERE id=7")
	require.NoError(t, err)
	require.Eventually(t, func() bool { return check(true) }, 3*time.Second, 10*time.Millisecond)
	policies, err := store.List(context.Background())
	require.NoError(t, err)
	policy := policies[1]
	now := services[0].now()
	policy.RateScheduleEnabled = true
	policy.RateSchedule = []CustomGroupRatePeriod{{StartsAt: now.Add(-time.Hour), EndsAt: now.Add(time.Hour), RateMultiplier: 0.15}}
	saved, err := store.Save(context.Background(), policy)
	require.NoError(t, err)
	require.Equal(t, policy.RateSchedule, saved.RateSchedule)
	require.Eventually(t, func() bool {
		for index, svc := range services {
			prepared, err := svc.PrepareKey(context.Background(), keys[index])
			if err != nil || prepared.Group.RateMultiplier != 0.15 {
				return false
			}
		}
		return true
	}, 3*time.Second, 10*time.Millisecond)
}
