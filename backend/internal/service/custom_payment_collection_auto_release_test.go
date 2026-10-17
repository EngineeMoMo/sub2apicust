//go:build unit

package service

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentauditlog"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func customAutoCollectionService(fixture *customReleaseFixture, enabled bool) *PaymentService {
	policy := customDefaultCollection()
	policy.MissingProviderAutoRelease = enabled
	raw, err := json.Marshal(policy)
	require.NoError(fixture.t, err)
	settings := &paymentConfigSettingRepoStub{values: map[string]string{customCollectionKey: string(raw)}}
	return &PaymentService{entClient: fixture.client, registry: payment.NewRegistry(), providersLoaded: true,
		loadBalancer:  newWebhookProviderTestLoadBalancer(fixture.client),
		configService: NewPaymentConfigService(fixture.client, settings, nil)}
}

func customAutoCollectionOrder(fixture *customReleaseFixture, status string, expires time.Time, binding string, keepProvider bool) *dbent.PaymentOrder {
	instance := customProviderGuardInstance(fixture.t, fixture.ctx, fixture.client)
	instanceID := strconv.FormatInt(instance.ID, 10)
	snapshot := map[string]any{"schema_version": 2, "provider_key": payment.TypeAlipay}
	if binding != "column" {
		value := instanceID
		if binding == "snapshot_whitespace" {
			value = " " + value + " "
		}
		snapshot["provider_instance_id"] = value
	}
	order := fixture.orderWithSnapshot(payment.TypeAlipay, status, expires, false, "", snapshot)
	if binding == "column" || binding == "both" {
		order = fixture.client.PaymentOrder.UpdateOneID(order.ID).SetProviderInstanceID(instanceID).SaveX(fixture.ctx)
	}
	if !keepProvider {
		require.NoError(fixture.t, fixture.client.PaymentProviderInstance.DeleteOneID(instance.ID).Exec(fixture.ctx))
	}
	return order
}

func TestCustomCollectionAutoMissingProviderPolicy(t *testing.T) {
	require.False(t, customDefaultCollection().MissingProviderAutoRelease)
	require.False(t, customParseCollection(`{"single_max":50,"daily_max":1000}`).MissingProviderAutoRelease)
	policy := customParseCollection(`{"single_max":50,"daily_max":1000,"unproven_release_hours":48,"missing_provider_auto_release":true}`)
	require.True(t, policy.MissingProviderAutoRelease)
	require.Equal(t, 48*time.Hour, customCollectionUnprovenWindowOf(policy))
}

func TestCustomCollectionAutoMissingProviderRelease(t *testing.T) {
	for _, status := range []string{OrderStatusCancelled, OrderStatusExpired, OrderStatusFailed} {
		for _, binding := range []string{"column", "snapshot", "both", "snapshot_whitespace"} {
			t.Run(status+"/"+binding, func(t *testing.T) {
				fixture := newCustomReleaseFixture(t)
				order := customAutoCollectionOrder(fixture, status, time.Now().Add(-25*time.Hour), binding, false)
				service := customAutoCollectionService(fixture, true)
				released, err := service.CustomReleaseUnpayableCollectionHolds(fixture.ctx)
				require.NoError(t, err)
				require.Equal(t, 1, released)
				require.True(t, fixture.released(order.ID))
				require.Equal(t, 1, fixture.audits(order.ID, "COLLECTION_HOLD_AUTO_RELEASED"))
				current := fixture.client.PaymentOrder.GetX(fixture.ctx, order.ID)
				require.Equal(t, status, current.Status)
				require.Nil(t, current.PaidAt)
				require.Zero(t, fixture.client.User.GetX(fixture.ctx, fixture.user.ID).Balance)
				released, err = service.CustomReleaseUnpayableCollectionHolds(fixture.ctx)
				require.NoError(t, err)
				require.Zero(t, released)
				require.Equal(t, 1, fixture.audits(order.ID, "COLLECTION_HOLD_AUTO_RELEASED"))
			})
		}
	}
}

func TestCustomCollectionAutoMissingProviderKeepsIneligibleOrders(t *testing.T) {
	for _, name := range []string{"disabled", "wait_not_over", "failed_entry_wait", "pending", "paid", "ledger_paid", "provider_exists", "provider_disabled", "missing_binding", "invalid_binding", "mismatch"} {
		t.Run(name, func(t *testing.T) {
			fixture := newCustomReleaseFixture(t)
			status, expires := OrderStatusCancelled, time.Now().Add(-25*time.Hour)
			if name == "wait_not_over" {
				expires = time.Now().Add(-23 * time.Hour)
			}
			if name == "failed_entry_wait" {
				status, expires = OrderStatusFailed, time.Now().Add(-20*time.Minute)
			}
			if name == "pending" {
				status = OrderStatusPending
			}
			order := customAutoCollectionOrder(fixture, status, expires, "snapshot", name == "provider_exists" || name == "provider_disabled")
			switch name {
			case "paid":
				fixture.client.PaymentOrder.UpdateOneID(order.ID).SetPaidAt(time.Now()).SaveX(fixture.ctx)
			case "ledger_paid":
				_, err := fixture.client.ExecContext(fixture.ctx, "UPDATE custom_payment_collection_ledger SET paid_at=$1 WHERE order_id=$2", time.Now(), order.ID)
				require.NoError(t, err)
			case "provider_disabled":
				instanceID, _, valid := customCollectionBoundInstance(order)
				require.True(t, valid)
				fixture.client.PaymentProviderInstance.UpdateOneID(instanceID).SetEnabled(false).SaveX(fixture.ctx)
			case "missing_binding":
				fixture.client.PaymentOrder.UpdateOneID(order.ID).SetProviderSnapshot(map[string]any{"schema_version": 2, "provider_key": payment.TypeAlipay}).SaveX(fixture.ctx)
			case "invalid_binding":
				fixture.client.PaymentOrder.UpdateOneID(order.ID).SetProviderSnapshot(map[string]any{"schema_version": 2, "provider_instance_id": "invalid"}).SaveX(fixture.ctx)
			case "mismatch":
				fixture.client.PaymentOrder.UpdateOneID(order.ID).SetProviderInstanceID("999999").SaveX(fixture.ctx)
			}
			service := customAutoCollectionService(fixture, name != "disabled")
			released, err := service.CustomReleaseUnpayableCollectionHolds(fixture.ctx)
			require.NoError(t, err)
			require.Zero(t, released)
			if name != "ledger_paid" {
				require.False(t, fixture.released(order.ID))
			}
			require.Zero(t, fixture.audits(order.ID, "COLLECTION_HOLD_AUTO_RELEASED"))
		})
	}
}

func TestCustomCollectionAutoMissingProviderRechecksCurrentBinding(t *testing.T) {
	for _, binding := range []string{"column", "snapshot"} {
		t.Run(binding, func(t *testing.T) {
			fixture := newCustomReleaseFixture(t)
			order := customAutoCollectionOrder(fixture, OrderStatusCancelled, time.Now().Add(-25*time.Hour), "both", false)
			update := fixture.client.PaymentOrder.UpdateOneID(order.ID)
			if binding == "column" {
				update.SetProviderInstanceID("999999")
			} else {
				update.SetProviderSnapshot(map[string]any{"schema_version": 2, "provider_instance_id": "999999"})
			}
			update.SaveX(fixture.ctx)
			released, err := customAutoCollectionService(fixture, true).customAutoReleaseMissingProviderHold(fixture.ctx, order, time.Now(), 24*time.Hour, "system")
			require.NoError(t, err)
			require.False(t, released)
			require.False(t, fixture.released(order.ID))
			require.Zero(t, fixture.audits(order.ID, "COLLECTION_HOLD_AUTO_RELEASED"))
		})
	}
}

func TestCustomCollectionAutoMissingProviderAuditFailureRollsBack(t *testing.T) {
	fixture := newCustomReleaseFixture(t)
	order := customAutoCollectionOrder(fixture, OrderStatusCancelled, time.Now().Add(-25*time.Hour), "both", false)
	_, err := fixture.client.ExecContext(fixture.ctx, "DROP TABLE payment_audit_logs")
	require.NoError(t, err)
	released, err := customAutoCollectionService(fixture, true).customAutoReleaseMissingProviderHold(fixture.ctx, order, time.Now(), 24*time.Hour, "system")
	require.Error(t, err)
	require.False(t, released)
	require.False(t, fixture.released(order.ID))
}

func TestCustomCollectionRecheckBypassesRetryBackoffOnly(t *testing.T) {
	fixture := newCustomReleaseFixture(t)
	old := customAutoCollectionOrder(fixture, OrderStatusCancelled, time.Now().Add(-25*time.Hour), "both", false)
	recent := customAutoCollectionOrder(fixture, OrderStatusFailed, time.Now().Add(-20*time.Minute), "both", false)
	require.NoError(t, customMarkCollectionChecked(fixture.ctx, fixture.client, old.ID, time.Now()))
	service := customAutoCollectionService(fixture, true)
	released, err := service.CustomReleaseUnpayableCollectionHolds(fixture.ctx)
	require.NoError(t, err)
	require.Zero(t, released)
	released, err = service.CustomRecheckCollectionHolds(fixture.ctx, 9)
	require.NoError(t, err)
	require.Equal(t, 1, released)
	require.True(t, fixture.released(old.ID))
	require.False(t, fixture.released(recent.ID))
	log := fixture.client.PaymentAuditLog.Query().Where(paymentauditlog.ActionEQ("COLLECTION_HOLD_AUTO_RELEASED")).OnlyX(fixture.ctx)
	require.Equal(t, "admin:9", log.Operator)
}

func TestCustomCollectionAutoMissingProviderBatchProgress(t *testing.T) {
	fixture := newCustomReleaseFixture(t)
	for index := 0; index < 45; index++ {
		customAutoCollectionOrder(fixture, OrderStatusCancelled, time.Now().Add(-25*time.Hour), "both", false)
	}
	service := customAutoCollectionService(fixture, true)
	for _, expected := range []int{20, 20, 5, 0} {
		released, err := service.CustomReleaseUnpayableCollectionHolds(fixture.ctx)
		require.NoError(t, err)
		require.Equal(t, expected, released)
	}
	require.Equal(t, 45, fixture.client.PaymentAuditLog.Query().Where(paymentauditlog.ActionEQ("COLLECTION_HOLD_AUTO_RELEASED")).CountX(fixture.ctx))
}
