//go:build unit

package service

import (
	"context"
	"strconv"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func customProviderGuardInstance(t *testing.T, ctx context.Context, client *dbent.Client) *dbent.PaymentProviderInstance {
	t.Helper()
	return client.PaymentProviderInstance.Create().SetProviderKey(payment.TypeAlipay).
		SetName("collection-guard").SetConfig("{}").SetSupportedTypes(payment.TypeAlipay).SaveX(ctx)
}

func TestCustomPaymentProviderDeletionPreservesCollectionHolds(t *testing.T) {
	for _, status := range []string{OrderStatusCancelled, OrderStatusExpired, OrderStatusFailed} {
		for _, binding := range []string{"column", "snapshot", "both", "snapshot_whitespace"} {
			t.Run(status+"/"+binding, func(t *testing.T) {
				fixture := newCustomReleaseFixture(t)
				instance := customProviderGuardInstance(t, fixture.ctx, fixture.client)
				instanceID := strconv.FormatInt(instance.ID, 10)
				order := fixture.order(payment.TypeAlipay, status, time.Now().Add(-25*time.Hour), false, "")
				update := fixture.client.PaymentOrder.UpdateOneID(order.ID)
				if binding == "column" || binding == "both" {
					update.SetProviderInstanceID(instanceID)
				}
				if binding != "column" {
					snapshotID := instanceID
					if binding == "snapshot_whitespace" {
						snapshotID = " " + instanceID + " "
					}
					update.SetProviderSnapshot(map[string]any{"schema_version": 2, "provider_key": payment.TypeAlipay, "provider_instance_id": snapshotID})
				}
				order = update.SaveX(fixture.ctx)
				config := &PaymentConfigService{entClient: fixture.client}

				err := config.DeleteProviderInstance(fixture.ctx, instance.ID)
				require.Error(t, err)
				require.Equal(t, "PROVIDER_COLLECTION_HOLDS", infraerrors.Reason(err))
				require.ErrorContains(t, err, "收款占用")
				require.False(t, fixture.released(order.ID))
				provider, err := (&PaymentService{entClient: fixture.client}).getOrderProviderInstance(fixture.ctx, order)
				require.NoError(t, err)
				require.Equal(t, instance.ID, provider.ID)
			})
		}
	}
}

func TestCustomPaymentProviderDeletionAllowsSettledHolds(t *testing.T) {
	for _, state := range []string{"released", "paid", "other_instance", "without_ledger", "without_orders"} {
		t.Run(state, func(t *testing.T) {
			fixture := newCustomReleaseFixture(t)
			instance := customProviderGuardInstance(t, fixture.ctx, fixture.client)
			if state != "without_orders" {
				boundID := instance.ID
				if state == "other_instance" {
					boundID = customProviderGuardInstance(t, fixture.ctx, fixture.client).ID
				}
				order := fixture.order(payment.TypeAlipay, OrderStatusExpired, time.Now().Add(-25*time.Hour), false, "")
				fixture.client.PaymentOrder.UpdateOneID(order.ID).SetProviderInstanceID(strconv.FormatInt(boundID, 10)).SaveX(fixture.ctx)
				var err error
				switch state {
				case "released":
					_, err = fixture.client.ExecContext(fixture.ctx, "UPDATE custom_payment_collection_ledger SET released=true WHERE order_id=$1", order.ID)
				case "paid":
					_, err = fixture.client.ExecContext(fixture.ctx, "UPDATE custom_payment_collection_ledger SET paid_at=$1 WHERE order_id=$2", time.Now(), order.ID)
				case "without_ledger":
					_, err = fixture.client.ExecContext(fixture.ctx, "DELETE FROM custom_payment_collection_ledger WHERE order_id=$1", order.ID)
				}
				require.NoError(t, err)
			}
			config := &PaymentConfigService{entClient: fixture.client}
			require.NoError(t, config.DeleteProviderInstance(fixture.ctx, instance.ID))
			_, err := fixture.client.PaymentProviderInstance.Get(fixture.ctx, instance.ID)
			require.True(t, dbent.IsNotFound(err))
		})
	}
}

func TestCustomPaymentProviderDeletionPreservesPendingProtection(t *testing.T) {
	for _, status := range []string{OrderStatusPending, OrderStatusPaid, OrderStatusRecharging} {
		t.Run(status, func(t *testing.T) {
			fixture := newCustomReleaseFixture(t)
			instance := customProviderGuardInstance(t, fixture.ctx, fixture.client)
			order := fixture.order(payment.TypeAlipay, status, time.Now().Add(time.Hour), false, "")
			fixture.client.PaymentOrder.UpdateOneID(order.ID).SetProviderInstanceID(strconv.FormatInt(instance.ID, 10)).SaveX(fixture.ctx)
			config := &PaymentConfigService{entClient: fixture.client}
			err := config.DeleteProviderInstance(fixture.ctx, instance.ID)
			require.Equal(t, "PENDING_ORDERS", infraerrors.Reason(err))
			_, err = fixture.client.PaymentProviderInstance.Get(fixture.ctx, instance.ID)
			require.NoError(t, err)
		})
	}
}

func TestCustomPaymentProviderDeletionFailsClosedOnLedgerError(t *testing.T) {
	fixture := newCustomReleaseFixture(t)
	instance := customProviderGuardInstance(t, fixture.ctx, fixture.client)
	_, err := fixture.client.ExecContext(fixture.ctx, "DROP TABLE custom_payment_collection_ledger")
	require.NoError(t, err)
	config := &PaymentConfigService{entClient: fixture.client}
	require.Error(t, config.DeleteProviderInstance(fixture.ctx, instance.ID))
	_, err = fixture.client.PaymentProviderInstance.Get(fixture.ctx, instance.ID)
	require.NoError(t, err)
}

func TestCustomCollectionMissingProviderRequiresManualRelease(t *testing.T) {
	fixture := newCustomReleaseFixture(t)
	instance := customProviderGuardInstance(t, fixture.ctx, fixture.client)
	order := fixture.orderWithSnapshot(payment.TypeAlipay, OrderStatusExpired, time.Now().Add(-25*time.Hour), false, "",
		map[string]any{"schema_version": 2, "provider_key": payment.TypeAlipay, "provider_instance_id": strconv.FormatInt(instance.ID, 10)})
	require.NoError(t, fixture.client.PaymentProviderInstance.DeleteOneID(instance.ID).Exec(fixture.ctx))
	service := &PaymentService{entClient: fixture.client, registry: payment.NewRegistry(), providersLoaded: true}

	released, err := service.CustomReleaseUnpayableCollectionHolds(fixture.ctx)
	require.NoError(t, err)
	require.Zero(t, released)
	require.False(t, fixture.released(order.ID))
	require.NoError(t, service.CustomForceReleaseCollection(fixture.ctx, order.ID, 9, "原支付通道已删除，已核实商户账单未收款", true))
	require.True(t, fixture.released(order.ID))
	require.Equal(t, 1, fixture.audits(order.ID, "COLLECTION_HOLD_FORCE_RELEASED"))
	released, err = service.CustomReleaseUnpayableCollectionHolds(fixture.ctx)
	require.NoError(t, err)
	require.Zero(t, released)
}
