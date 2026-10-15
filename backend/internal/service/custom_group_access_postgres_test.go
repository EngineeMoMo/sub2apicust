//go:build unit

package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestCustomGroupAccessPostgresRefundAndConcurrentPolicy(t *testing.T) {
	dsn := os.Getenv("CUSTOM_GROUP_ACCESS_TEST_DSN")
	if dsn == "" {
		t.Skip("CUSTOM_GROUP_ACCESS_TEST_DSN 未配置，未运行真实 PostgreSQL 门槛验证")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	db.SetMaxOpenConns(1)
	ctx := context.Background()
	var database string
	require.NoError(t, db.QueryRowContext(ctx, "SELECT current_database()").Scan(&database))
	require.Equal(t, "group_access_test", database, "仅允许专用隔离测试库")
	schema := fmt.Sprintf("group_access_%d", time.Now().UnixNano())
	_, err = db.ExecContext(ctx, "CREATE SCHEMA "+schema)
	require.NoError(t, err)
	defer func() { _, _ = db.ExecContext(context.Background(), "DROP SCHEMA "+schema+" CASCADE") }()
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		dsn, err = pq.ParseURL(dsn)
		require.NoError(t, err)
	}
	schemaDB, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	defer func() { _ = schemaDB.Close() }()
	schemaDB.SetMaxOpenConns(4)
	_, err = schemaDB.ExecContext(ctx, "SELECT 1")
	require.NoError(t, err)
	_, err = schemaDB.ExecContext(ctx, `CREATE TABLE groups(id BIGINT PRIMARY KEY,deleted_at TIMESTAMPTZ);
 CREATE TABLE users(id BIGINT PRIMARY KEY, balance NUMERIC(20,8) NOT NULL, deleted_at TIMESTAMPTZ);
 CREATE TABLE payment_orders(id BIGSERIAL PRIMARY KEY, user_id BIGINT, order_type TEXT DEFAULT 'balance',
 amount NUMERIC(20,2), bonus_amount NUMERIC(20,2) DEFAULT 0, pay_amount NUMERIC(20,2), refund_amount NUMERIC(20,2) DEFAULT 0,
 status TEXT, paid_at TIMESTAMPTZ);
 INSERT INTO groups VALUES(1),(2),(3);
 INSERT INTO users(id,balance) VALUES(7,200);`)
	require.NoError(t, err)
	migration, err := os.ReadFile("../../migrations/257_custom_group_access_policy.sql")
	require.NoError(t, err)
	_, err = schemaDB.ExecContext(ctx, string(migration))
	require.NoError(t, err)
	store := &customGroupAccessSQLStore{db: schemaDB}
	_, err = store.Save(ctx, CustomGroupAccessPolicy{GroupID: 1, MonthlyRechargeEnabled: true, MonthlyRechargeMin: 80})
	require.NoError(t, err)
	access, _, existingKey := customAccessTestFixture()
	access.store = store
	start, end := customGroupAccessMonth(time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC))
	var orderID int64
	err = schemaDB.QueryRowContext(ctx, `INSERT INTO payment_orders(user_id,amount,bonus_amount,pay_amount,status,paid_at)
 VALUES(7,120,20,999,'COMPLETED',$1) RETURNING id`, start).Scan(&orderID)
	require.NoError(t, err)
	_, err = schemaDB.ExecContext(ctx, `INSERT INTO payment_orders(user_id,amount,bonus_amount,pay_amount,status,paid_at,order_type) VALUES
 (7,500,0,500,'COMPLETED',$1,'balance'),
 (7,500,0,500,'COMPLETED',$2,'balance'),
 (7,500,0,500,'PENDING',$3,'balance'),
 (7,500,0,500,'COMPLETED',$3,'subscription')`, start.Add(-time.Nanosecond*1000), end, start.Add(time.Hour))
	require.NoError(t, err)
	for _, test := range []struct {
		status             string
		refunded, expected float64
	}{
		{"COMPLETED", 0, 100},
		{"PARTIALLY_REFUNDED", 60, 50},
		{"REFUND_REQUESTED", 120, 0},
		{"REFUND_PENDING", 120, 0},
		{"REFUND_FAILED", 120, 0},
		{"REFUNDED", 120, 0},
		{"COMPLETED", 120, 100},
	} {
		t.Run(test.status+fmt.Sprint(test.refunded), func(t *testing.T) {
			_, err := schemaDB.ExecContext(ctx, "UPDATE payment_orders SET status=$1,refund_amount=$2 WHERE id=$3", test.status, test.refunded, orderID)
			require.NoError(t, err)
			facts, err := store.Facts(ctx, 7, start, end)
			require.NoError(t, err)
			require.InDelta(t, test.expected, facts.MonthlyRecharge, 1e-8)
			require.Equal(t, float64(200), facts.Balance)
			_, eligibilityErr := access.PrepareKey(ctx, existingKey)
			if test.expected >= 80 {
				require.NoError(t, eligibilityErr)
			} else {
				require.ErrorIs(t, eligibilityErr, ErrGroupAccessRequirements)
			}
		})
	}
	_, err = schemaDB.ExecContext(ctx, "DELETE FROM custom_group_access_policies")
	require.NoError(t, err)
	expires := time.Now().Add(time.Hour)
	target, source := int64(2), int64(1)
	first, err := store.Save(ctx, CustomGroupAccessPolicy{GroupID: 1, ExpiresAt: &expires, FallbackGroupID: &target})
	require.NoError(t, err)
	require.NotNil(t, first.UpdatedAt)
	_, err = store.Save(ctx, CustomGroupAccessPolicy{GroupID: 1})
	require.ErrorIs(t, err, ErrGroupAccessPolicyStale)
	_, err = store.Save(ctx, CustomGroupAccessPolicy{GroupID: 2, ExpiresAt: &expires, FallbackGroupID: &source})
	require.Error(t, err)
	_, err = schemaDB.ExecContext(ctx, "DELETE FROM custom_group_access_policies")
	require.NoError(t, err)
	var wait sync.WaitGroup
	results := make(chan error, 2)
	for _, policy := range []CustomGroupAccessPolicy{
		{GroupID: 1, ExpiresAt: &expires, FallbackGroupID: &target},
		{GroupID: 2, ExpiresAt: &expires, FallbackGroupID: &source},
	} {
		wait.Add(1)
		go func(policy CustomGroupAccessPolicy) {
			defer wait.Done()
			_, err := store.Save(ctx, policy)
			results <- err
		}(policy)
	}
	wait.Wait()
	close(results)
	passed, failed := 0, 0
	for err := range results {
		if err == nil {
			passed++
		} else {
			failed++
		}
	}
	require.Equal(t, 1, passed)
	require.Equal(t, 1, failed)
}
