//go:build unit

package service

import (
	"context"
	"fmt"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCustomPlanStockValidation(t *testing.T) {
	require.NoError(t, customValidateStockLimit(nil))
	for _, n := range []int{-1, 0, 1, 2147483647} {
		require.NoError(t, customValidateStockLimit(&n))
	}
	invalid := -2
	require.Error(t, customValidateStockLimit(&invalid))
	for _, test := range []struct{ limit, used, want int }{{-1, 100, -1}, {0, 0, 0}, {3, 2, 1}, {3, 9, 0}} {
		require.Equal(t, test.want, CustomPlanStockRemaining(&dbent.SubscriptionPlan{StockLimit: test.limit, StockUsed: test.used}))
	}
	err := fmt.Errorf("wrapped: CUSTOM_PLAN_OUT_OF_STOCK")
	require.ErrorContains(t, customPlanStockError(err), "售罄")
	other := fmt.Errorf("connection failed")
	require.Same(t, other, customPlanStockError(other))
	require.NoError(t, customPlanStockError(nil))
}

func TestCustomPlanStockCRUDCompatibility(t *testing.T) {
	client := newPaymentConfigServiceTestClient(t)
	svc := &PaymentConfigService{entClient: client}
	ctx := context.Background()
	plan, err := svc.CreatePlan(ctx, CreatePlanRequest{GroupID: 1, Name: "套餐", Price: 60, ValidityDays: 30, ValidityUnit: "days", ForSale: true})
	require.NoError(t, err)
	require.Equal(t, -1, plan.StockLimit)
	require.Equal(t, 0, plan.StockUsed)
	limit := 10
	plan, err = svc.UpdatePlan(ctx, plan.ID, UpdatePlanRequest{StockLimit: &limit})
	require.NoError(t, err)
	_, err = client.SubscriptionPlan.UpdateOneID(plan.ID).SetStockUsed(3).Save(ctx)
	require.NoError(t, err)
	name := "改名不改库存"
	plan, err = svc.UpdatePlan(ctx, plan.ID, UpdatePlanRequest{Name: &name})
	require.NoError(t, err)
	require.Equal(t, 10, plan.StockLimit)
	require.Equal(t, 3, plan.StockUsed)
	limit = 0
	plan, err = svc.UpdatePlan(ctx, plan.ID, UpdatePlanRequest{StockLimit: &limit})
	require.NoError(t, err)
	require.Equal(t, 0, CustomPlanStockRemaining(plan))
	require.Equal(t, 3, plan.StockUsed)
	limit = -2
	_, err = svc.UpdatePlan(ctx, plan.ID, UpdatePlanRequest{StockLimit: &limit})
	require.Error(t, err)
}
