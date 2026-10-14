//go:build unit

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

// 完整迁移链生成的真实表结构上，验证 248 封顶触发器、到期复查、待确认列表与人工释放能执行，
// 避免只在简化夹具上通过而真实列名或类型不符。
func TestCustomCollectionReleaseRealSchemaPostgres(t *testing.T) {
	dsn := os.Getenv("DEDICATED_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func() { _ = admin.Close() }()
	name := fmt.Sprintf("collection_release_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE DATABASE " + name)
	require.NoError(t, err)
	defer func() { _, e := admin.Exec("DROP DATABASE " + name + " WITH (FORCE)"); require.NoError(t, e) }()
	u, err := url.Parse(dsn)
	require.NoError(t, err)
	u.Path = "/" + name
	db, err := sql.Open("postgres", u.String())
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	require.NoError(t, ApplyMigrations(ctx, db))
	require.NoError(t, ApplyMigrations(ctx, db), "重跑迁移按 checksum 跳过")

	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	cfg := service.NewPaymentConfigService(client, NewSettingRepository(client), nil)
	svc := service.NewPaymentService(client, payment.NewRegistry(), nil, nil, nil, cfg, nil, nil, nil)
	user := client.User.Create().SetEmail(fmt.Sprintf("release-%d@test.invalid", time.Now().UnixNano())).SetPasswordHash("test").SaveX(ctx)
	seq := 0
	order := func(amount float64, status string, expires time.Time) (*dbent.PaymentOrder, error) {
		seq++
		return client.PaymentOrder.Create().SetUserID(user.ID).SetUserEmail(user.Email).SetUserName("test").
			SetAmount(amount).SetPayAmount(amount).SetRechargeCode(fmt.Sprintf("REAL-%d", seq)).SetOutTradeNo(fmt.Sprintf("real-release-%d-%d", time.Now().UnixNano(), seq)).
			SetPaymentType(payment.TypeAlipay).SetPaymentTradeNo("").SetOrderType(payment.OrderTypeBalance).SetStatus(status).
			SetExpiresAt(expires).SetClientIP("127.0.0.1").SetSrcHost("test.invalid").
			// 快照没有实例编号：相当于渠道实例已删除，无法向渠道复查。
			SetProviderSnapshot(map[string]any{"schema_version": 2, "provider_key": payment.TypeAlipay}).Save(ctx)
	}

	stuck, err := order(10, service.OrderStatusCancelled, time.Now().Add(-25*time.Hour))
	require.NoError(t, err)
	var held int
	require.NoError(t, db.QueryRow("SELECT count(*) FROM custom_payment_collection_ledger WHERE order_id=$1 AND paid_at IS NULL AND NOT released", stuck.ID).Scan(&held))
	require.Equal(t, 1, held, "真实触发器为人民币订单写入预占")

	// 到期复查：真实表上 SQL 可执行；渠道无法解析时继续占用，只记复查时间。
	n, err := svc.CustomReleaseUnpayableCollectionHolds(ctx)
	require.NoError(t, err)
	require.Zero(t, n)
	var checked sql.NullTime
	require.NoError(t, db.QueryRow("SELECT release_checked_at FROM custom_payment_collection_ledger WHERE order_id=$1", stuck.ID).Scan(&checked))
	require.True(t, checked.Valid, "复查时间写入 248 新列")

	usage, err := cfg.CustomCollectionUsage(ctx, &service.CustomCollectionPolicy{Timezone: "Asia/Shanghai", DailyMax: 1000})
	require.NoError(t, err)
	require.Len(t, usage.Pending, 1)
	require.WithinDuration(t, stuck.ExpiresAt, usage.Pending[0].ExpiresAt, time.Millisecond)
	require.Equal(t, 24*time.Hour+10*time.Minute, usage.Pending[0].ManualReleaseAfter.Sub(usage.Pending[0].ExpiresAt))

	// 人工释放：写真实审计表，释放后不再计入占用。
	require.NoError(t, svc.CustomForceReleaseCollection(ctx, stuck.ID, 9, "渠道实例已删除，已核对商户账单无此笔收款", true))
	var released bool
	require.NoError(t, db.QueryRow("SELECT released FROM custom_payment_collection_ledger WHERE order_id=$1", stuck.ID).Scan(&released))
	require.True(t, released)
	var audits int
	require.NoError(t, db.QueryRow("SELECT count(*) FROM payment_audit_logs WHERE order_id=$1 AND action='COLLECTION_HOLD_FORCE_RELEASED' AND operator='admin:9'", fmt.Sprint(stuck.ID)).Scan(&audits))
	require.Equal(t, 1, audits)

	// 248 按用户封顶在真实 payment_orders 上生效。
	_, err = db.Exec(`UPDATE settings SET value='{"enabled":true,"single_max":50,"daily_max":1000,"user_hold_max":15,"timezone":"Asia/Shanghai","quick_amounts":[10],"allow_custom_amount":true}' WHERE key='CUSTOM_PAYMENT_COLLECTION'`)
	require.NoError(t, err)
	_, err = order(10, service.OrderStatusPending, time.Now().Add(30*time.Minute))
	require.NoError(t, err)
	_, err = order(10, service.OrderStatusPending, time.Now().Add(30*time.Minute))
	require.ErrorContains(t, err, "CUSTOM_COLLECTION_USER_HOLD_LIMIT")
}
