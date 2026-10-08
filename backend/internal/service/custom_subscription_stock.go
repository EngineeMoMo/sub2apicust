package service

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/subscriptionplan"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

type customSubscriptionStockContextKey struct{}
type customSubscriptionStockSelection struct{ planID, paidOrderID int64 }

// WithSubscriptionStockPlan 仅选择要扣的套餐，不能授权跳过扣库存。
func WithSubscriptionStockPlan(ctx context.Context, planID int64) context.Context {
	if planID <= 0 {
		return ctx
	}
	return context.WithValue(ctx, customSubscriptionStockContextKey{}, customSubscriptionStockSelection{planID: planID})
}

// 仅由已验证付款的内部履约流程调用，付款订单已预占，因此只记录套餐归属。
func customWithPaidSubscriptionStock(ctx context.Context, order *dbent.PaymentOrder) context.Context {
	selection := customSubscriptionStockSelection{paidOrderID: order.ID}
	if order.PlanID != nil {
		selection.planID = *order.PlanID
	}
	return context.WithValue(ctx, customSubscriptionStockContextKey{}, selection)
}

// 必须在订阅实际写入的同一事务调用。幂等复用不走这里，写入失败／缺库存一起回滚。
func (s *SubscriptionService) customConsumeSubscriptionStock(ctx context.Context, subID, userID, groupID int64) error {
	if s.entClient == nil {
		return nil
	} // 无数据库的纯领域测试替身；生产构造器注入 entClient。
	tx := dbent.TxFromContext(ctx)
	if tx == nil {
		return fmt.Errorf("subscription stock requires a transaction")
	}
	client := tx.Client()
	selection, _ := ctx.Value(customSubscriptionStockContextKey{}).(customSubscriptionStockSelection)
	planID := selection.planID
	if selection.paidOrderID > 0 && planID == 0 {
		return nil
	} // 历史已删除套餐的付款快照保持原履约。
	if selection.paidOrderID > 0 {
		// 套餐删除／调组后仍按已付款订单快照履约，不重扣、不依赖现售配置。
		_, err := client.ExecContext(ctx, `INSERT INTO custom_subscription_stock_allocations(subscription_id,plan_id,user_id,group_id,source,source_key) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(source_key) DO NOTHING`, subID, planID, userID, groupID, "payment", fmt.Sprintf("order:%d", selection.paidOrderID))
		return err
	}
	if planID == 0 {
		rows, err := client.QueryContext(ctx, "SELECT plan_id FROM custom_subscription_stock_allocations WHERE subscription_id=$1 ORDER BY id DESC LIMIT 1", subID)
		if err != nil {
			return fmt.Errorf("read subscription stock source: %w", err)
		}
		if rows.Next() {
			if err = rows.Scan(&planID); err != nil {
				_ = rows.Close()
				return err
			}
		}
		err = rows.Err()
		_ = rows.Close()
		if err != nil {
			return err
		}
	}
	if planID == 0 {
		plans, err := client.SubscriptionPlan.Query().Where(subscriptionplan.GroupIDEQ(groupID)).Limit(2).All(ctx)
		if err != nil {
			return err
		}
		if len(plans) != 1 {
			return infraerrors.BadRequest("SUBSCRIPTION_STOCK_PLAN_REQUIRED", "请先创建或选择该分组要扣库存的套餐；无法唯一确定套餐")
		}
		planID = plans[0].ID
	}
	plan, err := client.SubscriptionPlan.Get(ctx, planID)
	if err != nil || plan.GroupID != groupID {
		return infraerrors.BadRequest("SUBSCRIPTION_STOCK_PLAN_INVALID", "库存套餐不存在或不属于此订阅分组")
	}
	count, err := client.SubscriptionPlan.Update().Where(subscriptionplan.IDEQ(planID), subscriptionplan.GroupIDEQ(groupID), func(selector *sql.Selector) {
		selector.Where(sql.Or(sql.EQ(subscriptionplan.FieldStockLimit, -1), sql.ColumnsLT(subscriptionplan.FieldStockUsed, subscriptionplan.FieldStockLimit)))
	}).AddStockUsed(1).Save(ctx)
	if err != nil {
		return err
	}
	if count == 0 {
		return customPlanOutOfStock()
	}
	_, err = client.ExecContext(ctx, `INSERT INTO custom_subscription_stock_allocations(subscription_id,plan_id,user_id,group_id,source,source_key) VALUES($1,$2,$3,$4,$5,$6)`, subID, planID, userID, groupID, "manual", nil)
	return err
}

func (s *SubscriptionService) customCreateSubscriptionWithStock(ctx context.Context, input *AssignSubscriptionInput) (*UserSubscription, error) {
	var sub *UserSubscription
	err := s.withSubscriptionUpdateTx(ctx, func(txCtx context.Context) error {
		var err error
		sub, err = s.createSubscriptionWithoutStock(txCtx, input)
		if err != nil {
			return err
		}
		return s.customConsumeSubscriptionStock(txCtx, sub.ID, sub.UserID, sub.GroupID)
	})
	return sub, err
}
