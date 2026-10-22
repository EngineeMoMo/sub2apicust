//go:build unit

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type customAssignmentReadFailureRepo struct {
	service.UserSubscriptionRepository
}

type customAssignmentGroupRepo struct {
	service.GroupRepository
}

func (repo customAssignmentGroupRepo) GetByID(ctx context.Context, id int64) (*service.Group, error) {
	return repo.GetByIDLite(ctx, id)
}

func (customAssignmentReadFailureRepo) GetByID(context.Context, int64) (*service.UserSubscription, error) {
	return nil, errors.New("assignment snapshot unavailable")
}

func TestCustomAdminAssignmentStock(t *testing.T) {
	for _, scenario := range []string{"new", "active", "expired", "suspended", "sold_out", "wrong_plan", "snapshot_failure"} {
		t.Run(scenario, func(t *testing.T) {
			fixture := newCustomRedeemFixture(t, 1, 10)
			ctx := context.Background()
			plan := fixture.client.SubscriptionPlan.Query().OnlyX(ctx)
			input := &service.AssignSubscriptionInput{UserID: fixture.userID, GroupID: fixture.groupID, PlanID: plan.ID, ValidityDays: 30}
			var before *service.UserSubscription
			if scenario != "new" {
				var err error
				before, err = fixture.subscriptions.AssignSubscription(ctx, input)
				require.NoError(t, err)
				fixture.client.UserSubscription.UpdateOneID(before.ID).SetDailyUsageUsd(5).ExecX(ctx)
			}
			input.ValidityDays = 3
			expectedReason := ""
			svc := fixture.subscriptions
			switch scenario {
			case "expired":
				fixture.client.UserSubscription.UpdateOneID(before.ID).SetStatus(service.SubscriptionStatusExpired).SetExpiresAt(time.Now().Add(-time.Hour)).ExecX(ctx)
			case "suspended":
				fixture.client.UserSubscription.UpdateOneID(before.ID).SetStatus(service.SubscriptionStatusSuspended).ExecX(ctx)
				expectedReason = "SUBSCRIPTION_SUSPENDED"
			case "sold_out":
				fixture.client.SubscriptionPlan.UpdateOneID(plan.ID).SetStockLimit(1).ExecX(ctx)
				expectedReason = "PLAN_OUT_OF_STOCK"
			case "wrong_plan":
				otherGroup := fixture.client.Group.Create().SetName("another subscription group").SetSubscriptionType("subscription").SaveX(ctx)
				otherPlan := fixture.client.SubscriptionPlan.Create().SetName("another plan").SetGroupID(otherGroup.ID).SetPrice(10).SaveX(ctx)
				input.PlanID = otherPlan.ID
				expectedReason = "SUBSCRIPTION_STOCK_PLAN_INVALID"
			case "snapshot_failure":
				svc = service.NewSubscriptionService(customAssignmentGroupRepo{NewGroupRepository(fixture.client, nil)}, customAssignmentReadFailureRepo{customSequentialSubscriptionRepo{NewUserSubscriptionRepository(fixture.client)}}, nil, fixture.client, nil)
				t.Cleanup(svc.Stop)
			}
			result, err := svc.AssignAdminSubscription(ctx, input)
			used := fixture.client.SubscriptionPlan.GetX(ctx, plan.ID).StockUsed
			if expectedReason != "" || scenario == "snapshot_failure" {
				require.Error(t, err)
				if expectedReason != "" {
					require.Equal(t, expectedReason, infraerrors.Reason(err))
				} else {
					require.ErrorContains(t, err, "assignment snapshot unavailable")
				}
				require.Nil(t, result)
				require.Equal(t, 1, used)
				stored := fixture.client.UserSubscription.GetX(ctx, before.ID)
				require.True(t, before.ExpiresAt.Equal(stored.ExpiresAt))
				require.Equal(t, float64(5), stored.DailyUsageUsd)
				if scenario == "suspended" {
					require.Equal(t, service.SubscriptionStatusSuspended, stored.Status)
				}
				return
			}
			require.NoError(t, err)
			require.Equal(t, service.SubscriptionStatusActive, result.Status)
			if scenario == "new" || scenario == "expired" {
				require.WithinDuration(t, time.Now(), result.StartsAt, 3*time.Second)
				require.True(t, result.StartsAt.AddDate(0, 0, 3).Equal(result.ExpiresAt))
				require.Zero(t, result.DailyUsageUSD)
			} else {
				require.Equal(t, before.ID, result.ID)
				require.True(t, before.ExpiresAt.AddDate(0, 0, 3).Equal(result.ExpiresAt))
				require.Equal(t, before.StartsAt, result.StartsAt)
				require.Equal(t, float64(5), result.DailyUsageUSD)
			}
			if scenario == "new" {
				require.Equal(t, 1, used)
			} else {
				require.Equal(t, 2, used)
			}
		})
	}
}
