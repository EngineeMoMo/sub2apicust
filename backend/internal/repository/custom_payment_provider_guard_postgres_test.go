//go:build unit

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestCustomPaymentProviderDeletionRealSchemaPostgres(t *testing.T) {
	dsn := os.Getenv("DEDICATED_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func() { _ = admin.Close() }()
	name := fmt.Sprintf("provider_guard_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE DATABASE " + name)
	require.NoError(t, err)
	defer func() {
		_, err := admin.Exec("DROP DATABASE " + name + " WITH (FORCE)")
		require.NoError(t, err)
	}()
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
	user := client.User.Create().SetEmail(name + "@test.invalid").SetPasswordHash("test").SaveX(ctx)
	for _, status := range []string{service.OrderStatusCancelled, service.OrderStatusExpired, service.OrderStatusFailed} {
		for _, binding := range []string{"column", "snapshot", "both", "snapshot_whitespace"} {
			t.Run(status+"/"+binding, func(t *testing.T) {
				instance := client.PaymentProviderInstance.Create().SetProviderKey(payment.TypeAlipay).
					SetName("collection-guard").SetConfig("{}").SetSupportedTypes(payment.TypeAlipay).SaveX(ctx)
				instanceID := strconv.FormatInt(instance.ID, 10)
				snapshot := map[string]any{"schema_version": 2, "provider_key": payment.TypeAlipay}
				if binding != "column" {
					snapshotID := instanceID
					if binding == "snapshot_whitespace" {
						snapshotID = " " + instanceID + " "
					}
					snapshot["provider_instance_id"] = snapshotID
				}
				builder := client.PaymentOrder.Create().SetUserID(user.ID).SetUserEmail(user.Email).
					SetUserName("test").SetAmount(10).SetPayAmount(10).SetRechargeCode("GUARD-" + instanceID).
					SetOutTradeNo(name + "-" + instanceID).SetPaymentType(payment.TypeAlipay).SetPaymentTradeNo("").
					SetOrderType(payment.OrderTypeBalance).SetStatus(status).SetExpiresAt(time.Now().Add(-25 * time.Hour)).
					SetClientIP("127.0.0.1").SetSrcHost("test.invalid").SetProviderSnapshot(snapshot)
				if binding == "column" || binding == "both" {
					builder.SetProviderInstanceID(instanceID)
				}
				order := builder.SaveX(ctx)
				var released bool
				require.NoError(t, db.QueryRow("SELECT released FROM custom_payment_collection_ledger WHERE order_id=$1", order.ID).Scan(&released))
				require.False(t, released)

				err := config.DeleteProviderInstance(ctx, instance.ID)
				require.Error(t, err)
				require.Equal(t, "PROVIDER_COLLECTION_HOLDS", infraerrors.Reason(err))
				_, err = client.PaymentProviderInstance.Get(ctx, instance.ID)
				require.NoError(t, err)
				_, err = db.Exec("UPDATE custom_payment_collection_ledger SET released=true WHERE order_id=$1", order.ID)
				require.NoError(t, err)
				require.NoError(t, config.DeleteProviderInstance(ctx, instance.ID))
				_, err = client.PaymentProviderInstance.Get(ctx, instance.ID)
				require.True(t, dbent.IsNotFound(err))
			})
		}
	}
}
