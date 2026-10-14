//go:build unit

package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// 真实 PostgreSQL：248 按用户封顶、249 限量套餐单笔待付，以及到期复查粗筛 SQL 的时间比较。
func TestCustomCollectionHoldPostgres(t *testing.T) {
	dsn := os.Getenv("DEDICATED_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("DEDICATED_TEST_POSTGRES_DSN is not set; requires an isolated test database")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	schema := fmt.Sprintf("collection_hold_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = admin.Exec("DROP SCHEMA " + schema + " CASCADE")
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
	exec(`CREATE TABLE settings(key TEXT PRIMARY KEY, value TEXT NOT NULL);
 CREATE TABLE subscription_plans(id BIGINT PRIMARY KEY, for_sale BOOLEAN NOT NULL DEFAULT TRUE);
 CREATE TABLE payment_orders(id BIGSERIAL PRIMARY KEY, user_id BIGINT, plan_id BIGINT, order_type TEXT DEFAULT 'balance',
  status TEXT DEFAULT 'PENDING', paid_at TIMESTAMPTZ, pay_amount NUMERIC(20,2), provider_snapshot JSONB,
  expires_at TIMESTAMPTZ DEFAULT now() + interval '30 minutes', pay_url TEXT, qr_code TEXT);
 INSERT INTO subscription_plans(id) VALUES (1),(2);`)
	for _, name := range []string{"245_custom_subscription_plan_stock.sql", "247_custom_payment_collection.sql", "248_custom_payment_collection_user_hold.sql", "249_custom_plan_stock_pending_guard.sql"} {
		b, e := os.ReadFile("../../migrations/" + name)
		require.NoError(t, e)
		exec(string(b))
	}
	policy := func(raw string) {
		t.Helper()
		exec("TRUNCATE payment_orders, custom_payment_collection_ledger")
		exec("UPDATE settings SET value=$1 WHERE key='CUSTOM_PAYMENT_COLLECTION'", raw)
		exec("UPDATE subscription_plans SET stock_limit=-1, stock_used=0, sales_mode='online'")
	}
	insert := func(user int64, amount float64) error {
		_, e := db.Exec("INSERT INTO payment_orders(user_id, pay_amount) VALUES($1, $2)", user, amount)
		return e
	}

	t.Run("同一用户并发下单只占到封顶额度", func(t *testing.T) {
		policy(`{"enabled":true,"single_max":50,"daily_max":1000,"user_hold_max":150,"timezone":"Asia/Shanghai"}`)
		var ok atomic.Int32
		var wg sync.WaitGroup
		errs := make(chan error, 10)
		for range 10 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if e := insert(7, 50); e == nil {
					ok.Add(1)
				} else {
					errs <- e
				}
			}()
		}
		wg.Wait()
		close(errs)
		require.Equal(t, int32(3), ok.Load())
		for e := range errs {
			require.ErrorContains(t, e, "CUSTOM_COLLECTION_USER_HOLD_LIMIT")
		}
		require.NoError(t, insert(8, 50), "封顶按用户计算，不影响其他用户")
	})
	t.Run("取消后仍占用直到释放或付款", func(t *testing.T) {
		policy(`{"enabled":true,"single_max":50,"daily_max":1000,"user_hold_max":100,"timezone":"Asia/Shanghai"}`)
		require.NoError(t, insert(7, 50))
		require.NoError(t, insert(7, 50))
		exec("UPDATE payment_orders SET status='CANCELLED'")
		require.ErrorContains(t, insert(7, 1), "CUSTOM_COLLECTION_USER_HOLD_LIMIT", "取消不等于渠道确认失效")
		exec("UPDATE custom_payment_collection_ledger SET released=true WHERE order_id=(SELECT min(id) FROM payment_orders)")
		require.NoError(t, insert(7, 50))
		exec("UPDATE payment_orders SET paid_at=now(), status='PAID' WHERE id=(SELECT max(id) FROM payment_orders)")
		require.NoError(t, insert(7, 50), "已付款转为已收款，不再占用户额度")
	})
	t.Run("旧配置缺字段按单笔3倍封顶", func(t *testing.T) {
		policy(`{"enabled":true,"single_max":20,"daily_max":1000,"timezone":"Asia/Shanghai"}`)
		for range 3 {
			require.NoError(t, insert(7, 20))
		}
		require.ErrorContains(t, insert(7, 1), "CUSTOM_COLLECTION_USER_HOLD_LIMIT")
	})
	t.Run("未启用限额不封顶", func(t *testing.T) {
		policy(`{"enabled":false,"single_max":50,"daily_max":1000,"user_hold_max":50,"timezone":"Asia/Shanghai"}`)
		for range 5 {
			require.NoError(t, insert(7, 50))
		}
	})

	t.Run("到期复查粗筛按真实时间戳比较", func(t *testing.T) {
		policy(`{"enabled":false,"single_max":50,"daily_max":1000,"timezone":"Asia/Shanghai"}`)
		now := time.Now()
		order := func(status string, expires time.Time, payURL string) int64 {
			t.Helper()
			var id int64
			require.NoError(t, db.QueryRow("INSERT INTO payment_orders(user_id,pay_amount,status,expires_at,pay_url) VALUES(7,10,$1,$2,NULLIF($3,'')) RETURNING id", status, expires, payURL).Scan(&id))
			return id
		}
		wapURL := "https://openapi.alipay.com/gateway.do?biz_content=%7B%22time_expire%22%3A%222026-01-01+10%3A30%22%7D"
		failedNoEntry := order(OrderStatusFailed, now.Add(-11*time.Minute), "")
		wapDue := order(OrderStatusCancelled, now.Add(-11*time.Minute), wapURL)
		wapRecent := order(OrderStatusCancelled, now.Add(-5*time.Minute), wapURL)
		unprovenRecent := order(OrderStatusCancelled, now.Add(-2*time.Hour), "https://openapi.alipay.com/gateway.do?method=alipay.trade.page.pay")
		unprovenDue := order(OrderStatusExpired, now.Add(-25*time.Hour), "https://openapi.alipay.com/gateway.do?method=alipay.trade.page.pay")
		pendingOld := order(OrderStatusPending, now.Add(-48*time.Hour), "")
		checkedRecently := order(OrderStatusExpired, now.Add(-25*time.Hour), "")
		released := order(OrderStatusExpired, now.Add(-25*time.Hour), "")
		exec("UPDATE custom_payment_collection_ledger SET release_checked_at=$1 WHERE order_id=$2", now.Add(-10*time.Minute), checkedRecently)
		exec("UPDATE custom_payment_collection_ledger SET released=true WHERE order_id=$1", released)
		ids, e := customDueCollectionHolds(context.Background(), db, now, 24*time.Hour, 20)
		require.NoError(t, e)
		require.ElementsMatch(t, []int64{failedNoEntry, wapDue, unprovenDue}, ids)
		for _, id := range []int64{wapRecent, unprovenRecent, pendingOld, checkedRecently, released} {
			require.NotContains(t, ids, id)
		}
		require.NoError(t, customMarkCollectionChecked(context.Background(), db, unprovenDue, now))
		ids, e = customDueCollectionHolds(context.Background(), db, now, 24*time.Hour, 20)
		require.NoError(t, e)
		require.NotContains(t, ids, unprovenDue, "刚复查过的订单进入退避")
	})
	t.Run("限量套餐每人同时一笔待付", func(t *testing.T) {
		policy(`{"enabled":false,"single_max":50,"daily_max":1000,"timezone":"Asia/Shanghai"}`)
		exec("UPDATE subscription_plans SET stock_limit=5 WHERE id=1")
		sub := func(user, plan int64) error {
			_, e := db.Exec("INSERT INTO payment_orders(user_id,plan_id,order_type,pay_amount) VALUES($1,$2,'subscription',10)", user, plan)
			return e
		}
		require.NoError(t, sub(7, 1))
		require.ErrorContains(t, sub(7, 1), "CUSTOM_PLAN_PENDING_EXISTS")
		require.NoError(t, sub(8, 1), "其他用户不受影响")
		require.NoError(t, sub(7, 2), "不限量套餐不受影响")
		require.NoError(t, sub(7, 2))
		exec("UPDATE payment_orders SET status='CANCELLED' WHERE user_id=7 AND plan_id=1")
		require.NoError(t, sub(7, 1), "取消后可以重新下单")
		var used int
		require.NoError(t, db.QueryRow("SELECT stock_used FROM subscription_plans WHERE id=1").Scan(&used))
		require.Equal(t, 2, used, "被拒绝的待付订单不扣库存")

		exec("TRUNCATE payment_orders, custom_payment_collection_ledger")
		exec("UPDATE subscription_plans SET stock_limit=5, stock_used=0 WHERE id=2")
		var ok atomic.Int32
		var wg sync.WaitGroup
		for range 5 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if sub(9, 2) == nil {
					ok.Add(1)
				}
			}()
		}
		wg.Wait()
		require.Equal(t, int32(1), ok.Load(), "同一用户并发下单只保留一笔待付")
		require.NoError(t, db.QueryRow("SELECT stock_used FROM subscription_plans WHERE id=2").Scan(&used))
		require.Equal(t, 1, used)
	})
}
