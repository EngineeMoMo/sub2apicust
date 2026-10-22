package repository

import (
	"context"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/predicate"
	"github.com/Wei-Shaw/sub2api/ent/redeemcode"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (r *redeemCodeRepository) GetStockPlanID(ctx context.Context, codeID int64) (int64, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, "SELECT plan_id FROM custom_redeem_stock_sources WHERE redeem_code_id=$1", codeID)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()
	var planID int64
	if rows.Next() {
		if err := rows.Scan(&planID); err != nil {
			return 0, err
		}
	}
	return planID, rows.Err()
}

func (r *redeemCodeRepository) CreateWithStockPlan(ctx context.Context, code *service.RedeemCode, planID int64) error {
	if dbent.TxFromContext(ctx) != nil {
		return r.customCreateWithStockPlan(ctx, code, planID)
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := r.customCreateWithStockPlan(dbent.NewTxContext(ctx, tx), code, planID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *redeemCodeRepository) customCreateWithStockPlan(ctx context.Context, code *service.RedeemCode, planID int64) error {
	if code.Type != service.RedeemTypeSubscription || code.GroupID == nil || code.ValidityDays < 0 || planID <= 0 {
		return infraerrors.BadRequest("SUBSCRIPTION_STOCK_PLAN_INVALID", "库存套餐仅适用于增加订阅权益的兑换码")
	}
	if _, err := service.ResolveCustomRedeemStockPlan(ctx, clientFromContext(ctx, r.client), *code.GroupID, planID, true); err != nil {
		return err
	}
	if err := r.Create(ctx, code); err != nil {
		return err
	}
	return r.customSaveStockPlan(ctx, code.ID, planID)
}

func (r *redeemCodeRepository) customSaveStockPlan(ctx context.Context, codeID, planID int64) error {
	_, err := clientFromContext(ctx, r.client).ExecContext(ctx, "INSERT INTO custom_redeem_stock_sources(redeem_code_id,plan_id) VALUES($1,$2) ON CONFLICT(redeem_code_id) DO UPDATE SET plan_id=excluded.plan_id", codeID, planID)
	return err
}

func (r *redeemCodeRepository) customValidateBatchStock(ctx context.Context, client *dbent.Client, codes []*dbent.RedeemCode, fields service.RedeemCodeBatchUpdateFields) (map[int64]int64, error) {
	bindings := make(map[int64]int64)
	if !fields.GroupID.Set && !fields.PlanID.Set {
		return bindings, nil
	}
	for _, code := range codes {
		if code.Type != service.RedeemTypeSubscription {
			continue
		}
		groupID := code.GroupID
		if fields.GroupID.Set {
			groupID = fields.GroupID.Value
		}
		if groupID == nil || *groupID <= 0 {
			return nil, infraerrors.BadRequest("REDEEM_CODE_GROUP_ID_REQUIRED", "订阅兑换码不能清空分组")
		}
		planID := int64(0)
		if fields.PlanID.Set {
			if fields.PlanID.Value == nil || *fields.PlanID.Value <= 0 {
				return nil, infraerrors.BadRequest("SUBSCRIPTION_STOCK_PLAN_REQUIRED", "订阅兑换码必须选择库存套餐")
			}
			planID = *fields.PlanID.Value
		} else if code.GroupID != nil && *code.GroupID == *groupID && code.ValidityDays >= 0 {
			var err error
			planID, err = r.GetStockPlanID(ctx, code.ID)
			if err != nil {
				return nil, err
			}
		}
		resolved, err := service.ResolveCustomRedeemStockPlan(ctx, client, *groupID, planID, code.ValidityDays >= 0)
		if err != nil {
			return nil, err
		}
		if resolved > 0 {
			bindings[code.ID] = resolved
		}
	}
	if fields.PlanID.Set && len(bindings) == 0 {
		return nil, infraerrors.BadRequest("SUBSCRIPTION_STOCK_PLAN_INVALID", "库存来源仅适用于增加订阅权益的兑换码")
	}
	return bindings, nil
}

func customRedeemBatchPredicates(codes []*dbent.RedeemCode) predicate.RedeemCode {
	conditions := make([]predicate.RedeemCode, 0, len(codes))
	for _, code := range codes {
		groupCondition := redeemcode.GroupIDIsNil()
		if code.GroupID != nil {
			groupCondition = redeemcode.GroupIDEQ(*code.GroupID)
		}
		conditions = append(conditions, redeemcode.And(redeemcode.IDEQ(code.ID), groupCondition, redeemcode.TypeEQ(code.Type)))
	}
	return redeemcode.Or(conditions...)
}

var _ service.CustomRedeemStockRepository = (*redeemCodeRepository)(nil)
