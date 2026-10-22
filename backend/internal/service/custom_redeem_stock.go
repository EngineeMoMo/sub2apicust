package service

import (
	"context"
	"fmt"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/ent/subscriptionplan"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

type CustomRedeemStockRepository interface {
	CreateWithStockPlan(context.Context, *RedeemCode, int64) error
	GetStockPlanID(context.Context, int64) (int64, error)
}

func ResolveCustomRedeemStockPlan(ctx context.Context, client *dbent.Client, groupID, planID int64, consumesStock bool) (int64, error) {
	if client == nil {
		return 0, fmt.Errorf("redeem stock validation requires a database client")
	}
	if groupID <= 0 || planID < 0 {
		return 0, infraerrors.BadRequest("REDEEM_CODE_GROUP_ID_INVALID", "请选择有效的订阅分组和库存套餐")
	}
	group, err := client.Group.Query().Where(group.IDEQ(groupID), group.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return 0, err
	}
	if group.SubscriptionType != SubscriptionTypeSubscription {
		return 0, infraerrors.BadRequest("GROUP_NOT_SUBSCRIPTION_TYPE", "分组必须为订阅类型")
	}
	if !consumesStock {
		return 0, nil
	}
	if planID == 0 {
		plans, err := client.SubscriptionPlan.Query().Where(subscriptionplan.GroupIDEQ(groupID)).Limit(2).All(ctx)
		if err != nil {
			return 0, err
		}
		if len(plans) != 1 {
			return 0, infraerrors.BadRequest("SUBSCRIPTION_STOCK_PLAN_REQUIRED", "请选择此订阅分组的库存套餐；无法唯一确定套餐")
		}
		planID = plans[0].ID
	}
	plan, err := client.SubscriptionPlan.Get(ctx, planID)
	if dbent.IsNotFound(err) {
		return 0, infraerrors.BadRequest("SUBSCRIPTION_STOCK_PLAN_INVALID", "库存套餐不存在或不属于此订阅分组")
	}
	if err != nil {
		return 0, err
	}
	if plan.GroupID != groupID {
		return 0, infraerrors.BadRequest("SUBSCRIPTION_STOCK_PLAN_INVALID", "库存套餐不存在或不属于此订阅分组")
	}
	if plan.StockLimit >= 0 && plan.StockUsed >= plan.StockLimit {
		return 0, customPlanOutOfStock()
	}
	return planID, nil
}

func createCustomRedeemCodeWithStock(ctx context.Context, repo RedeemCodeRepository, code *RedeemCode, planID int64) error {
	if planID == 0 {
		return repo.Create(ctx, code)
	}
	stockRepo, ok := repo.(CustomRedeemStockRepository)
	if !ok {
		return fmt.Errorf("redeem repository does not support stock plan bindings")
	}
	return stockRepo.CreateWithStockPlan(ctx, code, planID)
}

func (s *RedeemService) customRedeemStockPlanID(ctx context.Context, codeID int64) (int64, error) {
	if stockRepo, ok := s.redeemRepo.(CustomRedeemStockRepository); ok {
		return stockRepo.GetStockPlanID(ctx, codeID)
	}
	return 0, nil
}
