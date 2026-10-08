//go:build unit

package service

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

type customCollectionProvider struct {
	*paymentOrderLifecycleQueryProvider
	closeErr error
}

func (p *customCollectionProvider) CancelPayment(context.Context, string) error { return p.closeErr }

func TestCustomCollectionCloseConfirmation(t *testing.T) {
	for _, fail := range []bool{true, false} {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			ctx := context.Background()
			client := newPaymentOrderLifecycleTestClient(t)
			_, err := client.ExecContext(ctx, `CREATE TABLE custom_payment_collection_ledger(order_id BIGINT,paid_at TIMESTAMP,released BOOLEAN DEFAULT false)`)
			require.NoError(t, err)
			user := client.User.Create().SetEmail(fmt.Sprintf("collection-%t@test.invalid", fail)).SetPasswordHash("test").SaveX(ctx)
			order := client.PaymentOrder.Create().SetUserID(user.ID).SetUserEmail(user.Email).SetUserName("test").SetAmount(50).SetPayAmount(50).SetRechargeCode("COLLECTION").SetOutTradeNo("collection-close").SetPaymentType(payment.TypeAlipay).SetPaymentTradeNo("").SetOrderType(payment.OrderTypeBalance).SetStatus(OrderStatusPending).SetExpiresAt(time.Now().Add(time.Hour)).SetClientIP("127.0.0.1").SetSrcHost("test.invalid").SaveX(ctx)
			_, err = client.ExecContext(ctx, `INSERT INTO custom_payment_collection_ledger(order_id) VALUES($1)`, order.ID)
			require.NoError(t, err)
			prov := &customCollectionProvider{paymentOrderLifecycleQueryProvider: &paymentOrderLifecycleQueryProvider{resp: &payment.QueryOrderResponse{Status: payment.ProviderStatusPending}}}
			if fail {
				prov.closeErr = fmt.Errorf("upstream unavailable")
			}
			registry := payment.NewRegistry()
			registry.Register(prov)
			svc := &PaymentService{entClient: client, registry: registry, providersLoaded: true}
			err = svc.CustomCloseCollection(ctx, order.ID)
			if fail {
				require.ErrorContains(t, err, "继续保留")
			} else {
				require.NoError(t, err)
			}
			rows, e := client.QueryContext(ctx, `SELECT released FROM custom_payment_collection_ledger`)
			require.NoError(t, e)
			require.True(t, rows.Next())
			var released bool
			require.NoError(t, rows.Scan(&released))
			_ = rows.Close()
			require.Equal(t, !fail, released)
			require.Equal(t, OrderStatusCancelled, client.PaymentOrder.GetX(ctx, order.ID).Status)
		})
	}
}

func TestCustomCollectionValidation(t *testing.T) {
	require.NoError(t, customValidateCollection(nil))
	require.NoError(t, customValidateCollection(customDefaultCollection()))
	for _, v := range []float64{-1, 0, 0.001, math.NaN(), math.Inf(1), 100000001} {
		p := customDefaultCollection()
		p.SingleMax = v
		require.Error(t, customValidateCollection(p))
	}
	p := customDefaultCollection()
	p.SingleMax = 49.99
	require.NoError(t, customValidateCollection(p))
	p.DailyMax = 40
	require.Error(t, customValidateCollection(p))
	p.DailyMax = 1000
	p.Timezone = ""
	require.Error(t, customValidateCollection(p))
	p.Timezone = "Local"
	require.Error(t, customValidateCollection(p))
	p.Timezone = "bad/zone"
	require.Error(t, customValidateCollection(p))
	p.Timezone = "Asia/Shanghai"
	p.QuickAmounts = []float64{10, 10}
	require.Error(t, customValidateCollection(p))
	require.False(t, customParseCollection("").Enabled, "升级不自动打开限额")
	require.Equal(t, 50.0, customParseCollection(`{"enabled":true,"single_max":50}`).SingleMax)
}

func TestCustomCollectionPlanAndInput(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	cfg := &PaymentConfigService{entClient: client}
	contact := "contact_admin"
	plan, err := cfg.CreatePlan(ctx, CreatePlanRequest{GroupID: 1, Name: "人工开通", Price: 60, ValidityDays: 30, ValidityUnit: "days", ForSale: true, SalesMode: &contact})
	require.NoError(t, err)
	require.Equal(t, contact, plan.SalesMode)
	name := "改名"
	plan, err = cfg.UpdatePlan(ctx, plan.ID, UpdatePlanRequest{Name: &name})
	require.NoError(t, err)
	require.Equal(t, contact, plan.SalesMode)
	svc := &PaymentService{configService: cfg}
	_, err = svc.validateSubOrder(ctx, CreateOrderRequest{PlanID: plan.ID})
	require.ErrorContains(t, err, "管理员")
	bad := "invalid"
	_, err = cfg.UpdatePlan(ctx, plan.ID, UpdatePlanRequest{SalesMode: &bad})
	require.Error(t, err)
	p := customDefaultCollection()
	p.AllowCustomAmount = false
	_, err = svc.validateOrderInput(ctx, CreateOrderRequest{OrderType: "balance", Amount: 11}, &PaymentConfig{Collection: p})
	require.ErrorContains(t, err, "快捷")
	_, err = svc.validateOrderInput(ctx, CreateOrderRequest{OrderType: "balance", Amount: 10}, &PaymentConfig{Collection: p})
	require.NoError(t, err)
}
