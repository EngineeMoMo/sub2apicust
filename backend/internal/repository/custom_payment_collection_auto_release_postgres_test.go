//go:build unit

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestCustomCollectionAutoMissingProviderRealSchemaPostgres(t *testing.T) {
	dsn := os.Getenv("DEDICATED_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func() { _ = admin.Close() }()
	name := fmt.Sprintf("collection_auto_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE DATABASE " + name)
	require.NoError(t, err)
	defer func() { _, err := admin.Exec("DROP DATABASE " + name + " WITH (FORCE)"); require.NoError(t, err) }()
	endpoint, err := url.Parse(dsn)
	require.NoError(t, err)
	endpoint.Path = "/" + name
	db, err := sql.Open("postgres", endpoint.String())
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	require.NoError(t, ApplyMigrations(ctx, db))
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	config := service.NewPaymentConfigService(client, NewSettingRepository(client), nil)
	paymentService := service.NewPaymentService(client, payment.NewRegistry(), nil, nil, nil, config, nil, nil, nil)
	_, err = db.Exec(`UPDATE settings SET value='{"enabled":true,"single_max":50,"daily_max":10000,"user_hold_max":1000,"timezone":"Asia/Shanghai","unproven_release_hours":24,"missing_provider_auto_release":true}' WHERE key='CUSTOM_PAYMENT_COLLECTION'`)
	require.NoError(t, err)
	user := client.User.Create().SetEmail(name + "@test.invalid").SetPasswordHash("test").SaveX(ctx)
	sequence := 0
	newInstance := func() *dbent.PaymentProviderInstance {
		return client.PaymentProviderInstance.Create().SetProviderKey(payment.TypeAlipay).SetConfig("{}").SaveX(ctx)
	}
	newOrder := func(instanceID int64) *dbent.PaymentOrder {
		sequence++
		return client.PaymentOrder.Create().SetUserID(user.ID).SetUserEmail(user.Email).SetUserName("test").
			SetAmount(10).SetPayAmount(10).SetRechargeCode(fmt.Sprintf("AUTO-%d", sequence)).SetOutTradeNo(fmt.Sprintf("%s-%d", name, sequence)).
			SetPaymentType(payment.TypeAlipay).SetPaymentTradeNo("").SetOrderType(payment.OrderTypeBalance).
			SetStatus(service.OrderStatusCancelled).SetExpiresAt(time.Now().Add(-25 * time.Hour)).SetClientIP("127.0.0.1").SetSrcHost("test.invalid").
			SetProviderInstanceID(strconv.FormatInt(instanceID, 10)).SetProviderSnapshot(map[string]any{
			"schema_version": 2, "provider_key": payment.TypeAlipay, "provider_instance_id": strconv.FormatInt(instanceID, 10),
		}).SaveX(ctx)
	}
	readLedger := func(orderID int64) (bool, sql.NullTime) {
		t.Helper()
		var released bool
		var paid sql.NullTime
		require.NoError(t, db.QueryRow("SELECT released,paid_at FROM custom_payment_collection_ledger WHERE order_id=$1", orderID).Scan(&released, &paid))
		return released, paid
	}

	t.Run("积压分批自动清理且迟到付款恢复已收账本", func(t *testing.T) {
		instance := newInstance()
		orders := make([]*dbent.PaymentOrder, 0, 45)
		for index := 0; index < 45; index++ {
			orders = append(orders, newOrder(instance.ID))
		}
		require.NoError(t, client.PaymentProviderInstance.DeleteOneID(instance.ID).Exec(ctx))
		for _, expected := range []int{20, 20, 5, 0} {
			released, err := paymentService.CustomReleaseUnpayableCollectionHolds(ctx)
			require.NoError(t, err)
			require.Equal(t, expected, released)
		}
		for _, order := range orders {
			released, paid := readLedger(order.ID)
			require.True(t, released)
			require.False(t, paid.Valid)
		}
		client.PaymentOrder.UpdateOneID(orders[0].ID).SetStatus(service.OrderStatusPaid).SetPaidAt(time.Now()).SaveX(ctx)
		released, paid := readLedger(orders[0].ID)
		require.False(t, released)
		require.True(t, paid.Valid)
		require.Zero(t, client.User.GetX(ctx, user.ID).Balance)
	})

	t.Run("审计写入失败时回滚占用释放", func(t *testing.T) {
		instance := newInstance()
		order := newOrder(instance.ID)
		require.NoError(t, client.PaymentProviderInstance.DeleteOneID(instance.ID).Exec(ctx))
		_, err := db.Exec(`CREATE FUNCTION custom_test_block_auto_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
 IF NEW.action='COLLECTION_HOLD_AUTO_RELEASED' THEN RAISE EXCEPTION 'auto audit unavailable'; END IF;
 RETURN NEW; END $$;
 CREATE TRIGGER custom_test_block_auto_audit BEFORE INSERT ON payment_audit_logs FOR EACH ROW EXECUTE FUNCTION custom_test_block_auto_audit()`)
		require.NoError(t, err)
		released, err := paymentService.CustomRecheckCollectionHolds(ctx, 9)
		require.NoError(t, err)
		require.Zero(t, released)
		wasReleased, _ := readLedger(order.ID)
		require.False(t, wasReleased)
		_, err = db.Exec("DROP TRIGGER custom_test_block_auto_audit ON payment_audit_logs; DROP FUNCTION custom_test_block_auto_audit()")
		require.NoError(t, err)
		released, err = paymentService.CustomRecheckCollectionHolds(ctx, 9)
		require.NoError(t, err)
		require.Equal(t, 1, released)
	})

	t.Run("付款更新与自动释放并发时保留已收账本", func(t *testing.T) {
		for iteration := 0; iteration < 10; iteration++ {
			instance := newInstance()
			order := newOrder(instance.ID)
			require.NoError(t, client.PaymentProviderInstance.DeleteOneID(instance.ID).Exec(ctx))
			var wait sync.WaitGroup
			wait.Add(2)
			start := make(chan struct{})
			errors := make(chan error, 2)
			go func() {
				defer wait.Done()
				<-start
				_, err := paymentService.CustomRecheckCollectionHolds(ctx, 9)
				errors <- err
			}()
			go func() {
				defer wait.Done()
				<-start
				_, err := client.PaymentOrder.UpdateOneID(order.ID).SetStatus(service.OrderStatusPaid).SetPaidAt(time.Now()).Save(ctx)
				errors <- err
			}()
			close(start)
			wait.Wait()
			close(errors)
			for err := range errors {
				require.NoError(t, err)
			}
			released, paid := readLedger(order.ID)
			require.False(t, released)
			require.True(t, paid.Valid)
		}
	})
}
