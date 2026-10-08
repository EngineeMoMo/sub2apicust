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

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/stretchr/testify/require"
)

func TestCustomCollectionPostgres(t *testing.T) {
	dsn := os.Getenv("DEDICATED_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	schema := fmt.Sprintf("collection_%d", time.Now().UnixNano())
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
	exec(`CREATE TABLE settings(key TEXT PRIMARY KEY,value TEXT NOT NULL);
 CREATE TABLE subscription_plans(id BIGINT PRIMARY KEY,for_sale BOOLEAN NOT NULL DEFAULT TRUE);
 CREATE TABLE payment_orders(id BIGSERIAL PRIMARY KEY,plan_id BIGINT,order_type TEXT DEFAULT 'balance',status TEXT DEFAULT 'PENDING',paid_at TIMESTAMPTZ,pay_amount NUMERIC(20,2),provider_snapshot JSONB);
 INSERT INTO subscription_plans(id) VALUES(1);`)
	for _, name := range []string{"245_custom_subscription_plan_stock.sql", "247_custom_payment_collection.sql"} {
		b, e := os.ReadFile(filepath.Join("..", "..", "migrations", name))
		require.NoError(t, e)
		exec(string(b))
	}
	reset := func() {
		t.Helper()
		exec("TRUNCATE payment_orders,custom_payment_collection_ledger")
		exec(`UPDATE settings SET value='{"enabled":true,"single_max":50,"daily_max":1000,"timezone":"Asia/Shanghai"}'`)
		exec("UPDATE subscription_plans SET sales_mode='online',stock_limit=-1,stock_used=0")
	}
	t.Run("线下登记幂等且影响在线额度", func(t *testing.T) {
		reset()
		client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
		svc := service.NewPaymentConfigService(client, nil, nil)
		require.NoError(t, svc.CustomRecordCollection(context.Background(), 980, "receipt-unique"))
		require.NoError(t, svc.CustomRecordCollection(context.Background(), 980, "receipt-unique"))
		require.Error(t, svc.CustomRecordCollection(context.Background(), 981, "receipt-unique"))
		_, e := db.Exec("INSERT INTO payment_orders(pay_amount) VALUES(21)")
		require.ErrorContains(t, e, "CUSTOM_COLLECTION_DAILY_LIMIT")
		exec("INSERT INTO payment_orders(pay_amount) VALUES(20)")
		usage, e := svc.CustomCollectionUsage(context.Background(), &service.CustomCollectionPolicy{Timezone: "Asia/Shanghai", DailyMax: 1000})
		require.NoError(t, e)
		require.Equal(t, 980.0, usage.Paid)
		require.Equal(t, 20.0, usage.Held)
		require.Zero(t, usage.Remaining)
	})
	t.Run("并发充值与订阅共用1000元", func(t *testing.T) {
		reset()
		var ok atomic.Int32
		var wg sync.WaitGroup
		errors := make(chan error, 32)
		for i := range 32 {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				kind := "balance"
				if i%2 == 0 {
					kind = "subscription"
				}
				_, e := db.Exec("INSERT INTO payment_orders(pay_amount,order_type,plan_id) VALUES(50,$1,1)", kind)
				if e == nil {
					ok.Add(1)
				} else {
					errors <- e
				}
			}(i)
		}
		wg.Wait()
		close(errors)
		require.Equal(t, int32(20), ok.Load())
		for e := range errors {
			require.ErrorContains(t, e, "CUSTOM_COLLECTION_DAILY_LIMIT")
		}
	})
	t.Run("单笔包含边界并拒绝其他币种", func(t *testing.T) {
		reset()
		exec("INSERT INTO payment_orders(pay_amount) VALUES(50)")
		_, e := db.Exec("INSERT INTO payment_orders(pay_amount) VALUES(50.01)")
		require.ErrorContains(t, e, "CUSTOM_COLLECTION_SINGLE_LIMIT")
		_, e = db.Exec(`INSERT INTO payment_orders(pay_amount,provider_snapshot) VALUES(10,'{"currency":"USD"}')`)
		require.ErrorContains(t, e, "CUSTOM_COLLECTION_CURRENCY")
	})
	t.Run("人工开通拒绝在线订单且不扣库存", func(t *testing.T) {
		reset()
		exec("UPDATE subscription_plans SET sales_mode='contact_admin'")
		_, e := db.Exec("INSERT INTO payment_orders(pay_amount,order_type,plan_id) VALUES(10,'subscription',1)")
		require.ErrorContains(t, e, "CUSTOM_PLAN_CONTACT_ADMIN")
		var used int
		require.NoError(t, db.QueryRow("SELECT stock_used FROM subscription_plans WHERE id=1").Scan(&used))
		require.Zero(t, used)
	})
	t.Run("取消超时跨日继续占用直到确认关闭", func(t *testing.T) {
		reset()
		exec(`UPDATE settings SET value=jsonb_set(value::jsonb,'{daily_max}','50')::text`)
		exec("INSERT INTO payment_orders(pay_amount,status) VALUES(50,'PENDING')")
		for _, status := range []string{"CANCELLED", "EXPIRED", "FAILED"} {
			exec("UPDATE payment_orders SET status=$1", status)
			_, e := db.Exec("INSERT INTO payment_orders(pay_amount) VALUES(1)")
			require.ErrorContains(t, e, "CUSTOM_COLLECTION_DAILY_LIMIT")
		}
		exec("UPDATE custom_payment_collection_ledger SET released=true")
		exec("INSERT INTO payment_orders(pay_amount) VALUES(50)")
	})
	t.Run("付款重放退款删除不重复计数或释放", func(t *testing.T) {
		reset()
		exec(`UPDATE settings SET value=jsonb_set(value::jsonb,'{daily_max}','50')::text`)
		exec("INSERT INTO payment_orders(pay_amount) VALUES(50)")
		exec("UPDATE payment_orders SET paid_at=now(),status='PAID'")
		exec("UPDATE payment_orders SET status='COMPLETED'")
		exec("UPDATE payment_orders SET status='REFUNDED'")
		exec("DELETE FROM payment_orders")
		_, e := db.Exec("INSERT INTO payment_orders(pay_amount) VALUES(1)")
		require.ErrorContains(t, e, "CUSTOM_COLLECTION_DAILY_LIMIT")
		exec("UPDATE custom_payment_collection_ledger SET paid_at=now()-interval '2 days'")
		exec("INSERT INTO payment_orders(pay_amount) VALUES(50)")
	})
	t.Run("订单事务回滚不留下预占", func(t *testing.T) {
		reset()
		tx, e := db.Begin()
		require.NoError(t, e)
		_, e = tx.Exec("INSERT INTO payment_orders(pay_amount) VALUES(50)")
		require.NoError(t, e)
		require.NoError(t, tx.Rollback())
		var count int
		require.NoError(t, db.QueryRow("SELECT count(*) FROM custom_payment_collection_ledger").Scan(&count))
		require.Zero(t, count)
	})
}
