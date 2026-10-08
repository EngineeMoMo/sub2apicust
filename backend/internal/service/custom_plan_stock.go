package service

import (
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// CustomPlanStockRemaining 返回可售份数；-1 表示不限量。计数由迁移中的事务触发器维护。
func CustomPlanStockRemaining(plan *dbent.SubscriptionPlan) int {
	if plan.StockLimit < 0 {
		return -1
	}
	return max(0, plan.StockLimit-plan.StockUsed)
}

func customValidateStockLimit(limit *int) error {
	if limit != nil && (*limit < -1 || *limit > 2147483647) {
		return infraerrors.BadRequest("PLAN_STOCK_INVALID", "库存总限额必须为 -1（不限量）或非负整数")
	}
	return nil
}

// 数据库条件更新负责并发裁决，服务层只将专用错误转成可读的 API 错误。
func customPlanStockError(err error) error {
	if err != nil && strings.Contains(err.Error(), "CUSTOM_PLAN_OUT_OF_STOCK") {
		return customPlanOutOfStock()
	}
	return err
}

func customPlanOutOfStock() error {
	return infraerrors.Conflict("PLAN_OUT_OF_STOCK", "套餐已售罄，请选择其他套餐或联系管理员")
}
