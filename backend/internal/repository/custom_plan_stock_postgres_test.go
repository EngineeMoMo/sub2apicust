//go:build unit

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// 真实 PostgreSQL：验证数据库事务裁决，而非单线程模拟计数。
func TestCustomPlanStockPostgres(t *testing.T) {
	dsn := os.Getenv("DEDICATED_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	schema := fmt.Sprintf("plan_stock_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, e := admin.Exec("DROP SCHEMA " + schema + " CASCADE")
		require.NoError(t, e)
		_ = admin.Close()
	})
	u, err := url.Parse(dsn)
	require.NoError(t, err)
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	db, err := sql.Open("postgres", u.String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	exec := func(query string, args ...any) { t.Helper(); _, e := db.Exec(query, args...); require.NoError(t, e) }
	used := func(id int) int {
		t.Helper()
		var n int
		require.NoError(t, db.QueryRow("SELECT stock_used FROM subscription_plans WHERE id=$1", id).Scan(&n))
		return n
	}
	exec(`CREATE TABLE subscription_plans(id BIGINT PRIMARY KEY, for_sale BOOLEAN NOT NULL DEFAULT TRUE);
 CREATE TABLE payment_orders(id BIGSERIAL PRIMARY KEY, plan_id BIGINT, order_type TEXT NOT NULL DEFAULT 'subscription', status TEXT NOT NULL DEFAULT 'PENDING', paid_at TIMESTAMPTZ);
 INSERT INTO subscription_plans(id) VALUES (1),(2),(3),(4);
 INSERT INTO payment_orders(plan_id,status,paid_at) VALUES (1,'COMPLETED',now()),(1,'PENDING',NULL),(1,'CANCELLED',NULL);`)
	migration, err := os.ReadFile(filepath.Join("..", "..", "migrations", "245_custom_subscription_plan_stock.sql"))
	require.NoError(t, err)
	exec(string(migration))
	require.Equal(t, 2, used(1))
	var limit int
	require.NoError(t, db.QueryRow("SELECT stock_limit FROM subscription_plans WHERE id=1").Scan(&limit))
	require.Equal(t, -1, limit)

	t.Run("最后一份并发不能超卖", func(t *testing.T) {
		exec("UPDATE subscription_plans SET stock_limit=1 WHERE id=2")
		var success atomic.Int32
		var wg sync.WaitGroup
		errs := make(chan error, 24)
		for range 24 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, e := db.Exec("INSERT INTO payment_orders(plan_id) VALUES(2)")
				if e == nil {
					success.Add(1)
				} else {
					errs <- e
				}
			}()
		}
		wg.Wait()
		close(errs)
		require.Equal(t, int32(1), success.Load())
		for e := range errs {
			require.Contains(t, e.Error(), "CUSTOM_PLAN_OUT_OF_STOCK")
		}
		require.Equal(t, 1, used(2))
	})
	t.Run("取消释放且重放不重复释放", func(t *testing.T) {
		exec("UPDATE payment_orders SET status='CANCELLED' WHERE plan_id=2")
		require.Equal(t, 0, used(2))
		exec("UPDATE payment_orders SET status='CANCELLED' WHERE plan_id=2")
		require.Equal(t, 0, used(2))
	})
	t.Run("晚付重新占用与售罄拒绝", func(t *testing.T) {
		exec("INSERT INTO payment_orders(plan_id) VALUES(2)")
		_, e := db.Exec("UPDATE payment_orders SET status='PAID',paid_at=now() WHERE plan_id=2 AND status='CANCELLED'")
		require.ErrorContains(t, e, "CUSTOM_PLAN_OUT_OF_STOCK")
		require.Equal(t, 1, used(2))
		exec("UPDATE subscription_plans SET stock_limit=2 WHERE id=2")
		exec("UPDATE payment_orders SET status='PAID',paid_at=now() WHERE plan_id=2 AND status='CANCELLED'")
		require.Equal(t, 2, used(2))
		exec("UPDATE payment_orders SET status='PAID',paid_at=now() WHERE plan_id=2")
		require.Equal(t, 2, used(2))
	})
	t.Run("已付失败重试退款删除均不补货", func(t *testing.T) {
		for _, status := range []string{"FAILED", "RECHARGING", "COMPLETED", "REFUNDING", "REFUNDED"} {
			exec("UPDATE payment_orders SET status=$1 WHERE plan_id=2", status)
			require.Equal(t, 2, used(2))
		}
		exec("DELETE FROM payment_orders WHERE plan_id=2")
		require.Equal(t, 2, used(2))
		exec(string(migration))
		require.Equal(t, 2, used(2))
		require.Equal(t, 2, used(1))
	})
	t.Run("超时失败事务回滚", func(t *testing.T) {
		exec("UPDATE subscription_plans SET stock_limit=1 WHERE id=3")
		tx, e := db.Begin()
		require.NoError(t, e)
		_, e = tx.Exec("INSERT INTO payment_orders(plan_id) VALUES(3)")
		require.NoError(t, e)
		require.NoError(t, tx.Rollback())
		require.Equal(t, 0, used(3))
		for _, status := range []string{"EXPIRED", "FAILED"} {
			exec("INSERT INTO payment_orders(plan_id) VALUES(3)")
			require.Equal(t, 1, used(3))
			exec("UPDATE payment_orders SET status=$1 WHERE plan_id=3 AND status='PENDING'", status)
			require.Equal(t, 0, used(3))
		}
	})
	t.Run("管理员改限额不覆盖占用与零库存", func(t *testing.T) {
		exec("UPDATE subscription_plans SET stock_limit=0 WHERE id=1")
		require.Equal(t, 2, used(1))
		_, e := db.Exec("INSERT INTO payment_orders(plan_id) VALUES(1)")
		require.ErrorContains(t, e, "CUSTOM_PLAN_OUT_OF_STOCK")
		exec("UPDATE subscription_plans SET stock_limit=3 WHERE id=1")
		exec("INSERT INTO payment_orders(plan_id) VALUES(1)")
		require.Equal(t, 3, used(1))
	})
	t.Run("不限量与余额订单不受影响", func(t *testing.T) {
		exec("INSERT INTO payment_orders(plan_id) SELECT 4 FROM generate_series(1,30)")
		require.Equal(t, 30, used(4))
		exec("INSERT INTO payment_orders(plan_id,order_type) VALUES(1,'balance')")
		require.Equal(t, 3, used(1))
		exec("UPDATE subscription_plans SET for_sale=FALSE WHERE id=4")
		_, e := db.Exec("INSERT INTO payment_orders(plan_id) VALUES(4)")
		require.ErrorContains(t, e, "CUSTOM_PLAN_OUT_OF_STOCK")
	})
	t.Run("完整迁移链新库与重跑", func(t *testing.T) {
		// 既有迁移检查 public schema；完整迁移必须使用独立数据库，不能更换 schema 冒充新库。
		fullName := schema + "_full"
		_, e := admin.Exec("CREATE DATABASE " + fullName)
		require.NoError(t, e)
		t.Cleanup(func() { _, e := admin.Exec("DROP DATABASE " + fullName); require.NoError(t, e) })
		fullURL := *u
		query := fullURL.Query()
		query.Del("search_path")
		fullURL.RawQuery = query.Encode()
		fullURL.Path = "/" + fullName
		fullDB, e := sql.Open("postgres", fullURL.String())
		require.NoError(t, e)
		defer func() { _ = fullDB.Close() }()
		require.NoError(t, ApplyMigrations(context.Background(), fullDB))
		require.NoError(t, ApplyMigrations(context.Background(), fullDB))
		var count int
		require.NoError(t, fullDB.QueryRow("SELECT count(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='subscription_plans' AND column_name IN ('stock_limit','stock_used')").Scan(&count))
		require.Equal(t, 2, count)
	})
}
