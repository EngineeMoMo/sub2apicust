//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestCustomRedeemStockConcurrentRenewals(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client := testEntClient(t)
	marker := fmt.Sprintf("redeem-stock-%d", time.Now().UnixNano())
	user, err := client.User.Create().SetEmail(marker + "@example.invalid").SetPasswordHash("test").Save(ctx)
	require.NoError(t, err)
	group, err := client.Group.Create().SetName(marker).SetSubscriptionType("subscription").Save(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		for _, statement := range []struct {
			query string
			id    int64
		}{
			{"DELETE FROM redeem_codes WHERE group_id=$1", group.ID},
			{"DELETE FROM custom_subscription_stock_allocations WHERE group_id=$1", group.ID},
			{"DELETE FROM user_subscriptions WHERE group_id=$1", group.ID},
			{"DELETE FROM subscription_plans WHERE group_id=$1", group.ID},
			{"DELETE FROM groups WHERE id=$1", group.ID},
			{"DELETE FROM users WHERE id=$1", user.ID},
		} {
			_, err := integrationDB.ExecContext(context.Background(), statement.query, statement.id)
			require.NoError(t, err)
		}
	})
	previous, err := client.SubscriptionPlan.Create().SetName(marker + "-old").SetGroupID(group.ID).SetPrice(1).SetStockLimit(-1).SetStockUsed(1).Save(ctx)
	require.NoError(t, err)
	selected, err := client.SubscriptionPlan.Create().SetName(marker + "-selected").SetGroupID(group.ID).SetPrice(1).SetStockLimit(1).Save(ctx)
	require.NoError(t, err)
	now := time.Now().UTC().Truncate(time.Microsecond)
	subscription, err := client.UserSubscription.Create().SetUserID(user.ID).SetGroupID(group.ID).SetStartsAt(now.Add(-time.Hour)).SetExpiresAt(now.AddDate(0, 0, 7)).SetStatus(service.SubscriptionStatusActive).Save(ctx)
	require.NoError(t, err)
	_, err = integrationDB.ExecContext(ctx, "INSERT INTO custom_subscription_stock_allocations(subscription_id,plan_id,user_id,group_id,source) VALUES($1,$2,$3,$4,'manual')", subscription.ID, previous.ID, user.ID, group.ID)
	require.NoError(t, err)
	codes := NewRedeemCodeRepository(client)
	sourceRepo := codes.(service.CustomRedeemStockRepository)
	subscriptions := service.NewSubscriptionService(NewGroupRepository(client, integrationDB), NewUserSubscriptionRepository(client), nil, client, nil)
	t.Cleanup(subscriptions.Stop)
	redeem := service.NewRedeemService(codes, NewUserRepository(client, integrationDB), subscriptions, nil, nil, client, nil, nil)
	codeValues := []string{fmt.Sprintf("crs-%d-1", now.UnixNano()), fmt.Sprintf("crs-%d-2", now.UnixNano())}
	for _, value := range codeValues {
		code := &service.RedeemCode{Code: value, Type: service.RedeemTypeSubscription, Status: service.StatusUnused, GroupID: &group.ID, ValidityDays: 7}
		require.NoError(t, sourceRepo.CreateWithStockPlan(ctx, code, selected.ID))
	}
	results := make(chan error, len(codeValues))
	start := make(chan struct{})
	for _, value := range codeValues {
		go func(code string) { <-start; _, err := redeem.Redeem(ctx, user.ID, code); results <- err }(value)
	}
	close(start)
	successes, stockFailures := 0, 0
	for range codeValues {
		select {
		case err := <-results:
			if err == nil {
				successes++
			} else {
				require.Equal(t, "PLAN_OUT_OF_STOCK", infraerrors.Reason(err))
				stockFailures++
			}
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, stockFailures)
	updated, err := client.UserSubscription.Get(ctx, subscription.ID)
	require.NoError(t, err)
	require.Equal(t, subscription.ExpiresAt.AddDate(0, 0, 7), updated.ExpiresAt)
	chosen, err := client.SubscriptionPlan.Get(ctx, selected.ID)
	require.NoError(t, err)
	require.Equal(t, 1, chosen.StockUsed)
	original, err := client.SubscriptionPlan.Get(ctx, previous.ID)
	require.NoError(t, err)
	require.Equal(t, 1, original.StockUsed)
	for _, value := range codeValues {
		code, err := codes.GetByCode(ctx, value)
		require.NoError(t, err)
		if code.Status == service.StatusUnused {
			require.Nil(t, code.UsedBy)
			require.NoError(t, codes.Delete(ctx, code.ID))
			planID, err := sourceRepo.GetStockPlanID(ctx, code.ID)
			require.NoError(t, err)
			require.Zero(t, planID, "deleting an unused code must cascade its stock binding")
		}
	}
}
