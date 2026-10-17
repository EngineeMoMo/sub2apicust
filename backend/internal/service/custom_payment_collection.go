package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// [CUSTOM] 同一收款账户的人民币总限额，覆盖所有通道、充值与订阅。
const customCollectionKey = "CUSTOM_PAYMENT_COLLECTION"

type CustomCollectionPolicy struct {
	Enabled           bool      `json:"enabled"`
	SingleMax         float64   `json:"single_max"`
	DailyMax          float64   `json:"daily_max"`
	Timezone          string    `json:"timezone"`
	QuickAmounts      []float64 `json:"quick_amounts"`
	AllowCustomAmount bool      `json:"allow_custom_amount"`
	ContactText       string    `json:"contact_text"`
	// 同一用户未付且未释放的预占合计上限；缺省为单笔上限的3倍且不超过每日上限（248 触发器同一口径）。
	UserHoldMax float64 `json:"user_hold_max"`
	// 无法由付款入口证明期限的订单（旧手机 WAP、电脑网站支付、易支付等）在订单期限后再等多少小时才复查释放。
	UnprovenReleaseHours       int  `json:"unproven_release_hours"`
	MissingProviderAutoRelease bool `json:"missing_provider_auto_release"`
}

const customCollectionDefaultUnprovenHours = 24

func customDefaultCollection() *CustomCollectionPolicy {
	return customNormalizeCollection(customBaseCollection())
}

// 新字段留空，由 customNormalizeCollection 按实际单笔／每日上限推导，与触发器缺字段时一致。
func customBaseCollection() *CustomCollectionPolicy {
	return &CustomCollectionPolicy{SingleMax: 50, DailyMax: 1000, Timezone: "Asia/Shanghai", QuickAmounts: []float64{10, 20, 30, 40}, AllowCustomAmount: true}
}

// 旧配置和旧前端不带新字段时补默认值，保存前与读取后口径一致。
func customNormalizeCollection(p *CustomCollectionPolicy) *CustomCollectionPolicy {
	if p == nil {
		return nil
	}
	// 只把缺字段（0）视为未设置；负数保留给校验拒绝，不悄悄改成默认值。
	if p.UserHoldMax == 0 {
		p.UserHoldMax = math.Min(p.SingleMax*3, p.DailyMax)
	}
	if p.UnprovenReleaseHours == 0 {
		p.UnprovenReleaseHours = customCollectionDefaultUnprovenHours
	}
	return p
}

func customParseCollection(raw string) *CustomCollectionPolicy {
	p := customBaseCollection()
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), p); err != nil {
			return customDefaultCollection()
		}
	}
	return customNormalizeCollection(p)
}

func customMoneyValid(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v > 0 && v <= 100000000 && math.Abs(v*100-math.Round(v*100)) < 0.000001
}

func customValidateCollection(p *CustomCollectionPolicy) error {
	if p == nil {
		return nil
	}
	customNormalizeCollection(p)
	if !customMoneyValid(p.SingleMax) || !customMoneyValid(p.DailyMax) || p.SingleMax > p.DailyMax {
		return infraerrors.BadRequest("COLLECTION_CONFIG_INVALID", "单笔和每日收款上限必须为正数、最多两位小数，且单笔不能超过每日上限")
	}
	if !customMoneyValid(p.UserHoldMax) || p.UserHoldMax < p.SingleMax || p.UserHoldMax > p.DailyMax {
		return infraerrors.BadRequest("COLLECTION_CONFIG_INVALID", "每个用户的待付占用上限必须为正数、最多两位小数，且不低于单笔上限、不超过每日上限")
	}
	if p.UnprovenReleaseHours < 1 || p.UnprovenReleaseHours > 360 {
		return infraerrors.BadRequest("COLLECTION_CONFIG_INVALID", "无法证明期限的订单复查等待时间须为1至360小时")
	}
	if _, err := time.LoadLocation(p.Timezone); err != nil || p.Timezone == "" || p.Timezone == "Local" {
		return infraerrors.BadRequest("COLLECTION_CONFIG_INVALID", "收款时区无效")
	}
	if len(p.ContactText) > 2000 || len(p.QuickAmounts) == 0 || len(p.QuickAmounts) > 20 {
		return infraerrors.BadRequest("COLLECTION_CONFIG_INVALID", "请配置1至20个快捷金额，联系说明不超过2000字节")
	}
	seen := map[float64]bool{}
	for _, v := range p.QuickAmounts {
		if !customMoneyValid(v) || seen[v] {
			return infraerrors.BadRequest("COLLECTION_CONFIG_INVALID", "快捷金额必须为正数、最多两位小数，不能重复")
		}
		seen[v] = true
	}
	return nil
}

func customValidateSalesMode(mode *string) error {
	if mode != nil && *mode != "online" && *mode != "contact_admin" {
		return infraerrors.BadRequest("PLAN_SALES_MODE_INVALID", "开通方式只能为在线购买或联系管理员")
	}
	return nil
}

func customCollectionError(err error) error {
	if err == nil {
		return nil
	}
	for _, rule := range []struct{ code, msg string }{
		{"CUSTOM_COLLECTION_SINGLE_LIMIT", "实付金额超过单笔收款上限，请调整金额或联系管理员"},
		{"CUSTOM_COLLECTION_DAILY_LIMIT", "今日在线收款额度不足，请稍后再试或联系管理员"},
		{"CUSTOM_COLLECTION_CURRENCY", "已启用人民币收款限额，当前仅支持人民币在线支付"},
		{"CUSTOM_COLLECTION_USER_HOLD_LIMIT", "你有未完成付款的订单正在占用收款额度，请先完成付款，或等订单过期、系统确认未付款后再下单"},
		{"CUSTOM_PLAN_CONTACT_ADMIN", "此套餐需联系管理员开通，不能在线购买"},
	} {
		if strings.Contains(err.Error(), rule.code) {
			return infraerrors.Conflict(rule.code, rule.msg)
		}
	}
	return err
}

type CustomCollectionUsage struct {
	Paid      float64                `json:"paid"`
	Held      float64                `json:"held"`
	Remaining float64                `json:"remaining"`
	Pending   []CustomCollectionHold `json:"pending"`
}

type CustomCollectionHold struct {
	OrderID   int64     `json:"order_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
	// 管理员人工释放的最早时间：订单期限＋无法证明期限的等待时间＋余量。
	ManualReleaseAfter time.Time `json:"manual_release_after"`
}

func (s *PaymentConfigService) CustomCollectionUsage(ctx context.Context, p *CustomCollectionPolicy) (*CustomCollectionUsage, error) {
	rows, err := s.entClient.QueryContext(ctx, `SELECT COALESCE(sum(amount) FILTER(WHERE paid_at IS NOT NULL AND (paid_at AT TIME ZONE $1)::date=(now() AT TIME ZONE $1)::date),0), COALESCE(sum(amount) FILTER(WHERE paid_at IS NULL AND NOT released),0) FROM custom_payment_collection_ledger`, p.Timezone)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	u := &CustomCollectionUsage{}
	if rows.Next() {
		if err = rows.Scan(&u.Paid, &u.Held); err != nil {
			return nil, err
		}
	}
	u.Remaining = math.Max(0, p.DailyMax-u.Paid-u.Held)
	if err = rows.Err(); err != nil {
		return nil, err
	}
	_ = rows.Close()
	u.Pending = []CustomCollectionHold{}
	// 第二次查询使用独立变量：第一个 defer 捕获的是 rows，查询失败时不能把它覆盖成 nil。
	pending, err := s.entClient.QueryContext(ctx, `SELECT l.order_id,l.amount,o.status,o.expires_at FROM custom_payment_collection_ledger l JOIN payment_orders o ON o.id=l.order_id WHERE l.paid_at IS NULL AND NOT l.released ORDER BY l.order_id DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = pending.Close() }()
	wait := customCollectionUnprovenWindowOf(p)
	for pending.Next() {
		var h CustomCollectionHold
		if err = pending.Scan(&h.OrderID, &h.Amount, &h.Status, &h.ExpiresAt); err != nil {
			return nil, err
		}
		h.ManualReleaseAfter = h.ExpiresAt.Add(wait + customCollectionReleaseMargin)
		u.Pending = append(u.Pending, h)
	}
	return u, pending.Err()
}

// 管理员“查询并关闭”：先按原流程查询并关单；未能关单时按可证明的期限复查渠道，确认未付才释放。
// 没有渠道证据的释放只能走 CustomForceReleaseCollection（需原因、确认且过等待期），见 custom_payment_collection_release.go。
func (s *PaymentService) CustomCloseCollection(ctx context.Context, id int64) error {
	order, err := s.entClient.PaymentOrder.Get(ctx, id)
	if err != nil {
		return infraerrors.NotFound("ORDER_NOT_FOUND", "订单不存在")
	}
	if order.PaidAt != nil {
		return infraerrors.Conflict("ORDER_PAID", "订单已付款，不释放收款额度")
	}
	if order.Status == OrderStatusPending {
		result, cancelErr := s.cancelCore(ctx, order, OrderStatusCancelled, "admin", "collection hold reconciliation")
		if cancelErr != nil {
			return cancelErr
		}
		if result == checkPaidResultAlreadyPaid {
			return infraerrors.Conflict("ORDER_PAID", "渠道确认订单已付款，已尝试同步付款状态")
		}
	}
	wait, err := s.customCollectionUnprovenWindow(ctx)
	if err != nil {
		return err
	}
	current, err := s.entClient.PaymentOrder.Get(ctx, id)
	if err != nil {
		return infraerrors.NotFound("ORDER_NOT_FOUND", "订单不存在")
	}
	prov, err := s.getOrderProvider(ctx, current)
	if err != nil {
		return infraerrors.Conflict("COLLECTION_CLOSE_UNCONFIRMED", "无法连接该订单的支付渠道，额度继续保留；渠道已删除或长期不可用时，可在等待期过后填写原因人工释放")
	}
	// 待付订单刚由 cancelCore 查询并关单，不再重复请求渠道。
	switch outcome, deadline := s.customRecheckCollectionHold(ctx, current, prov, time.Now(), wait, "admin", order.Status != OrderStatusPending); outcome {
	case customHoldReleased:
		return nil
	case customHoldPaid:
		return infraerrors.Conflict("ORDER_PAID", "渠道确认订单已付款，已尝试同步付款状态")
	case customHoldNotDue:
		return infraerrors.Conflict("COLLECTION_HOLD_NOT_DUE", fmt.Sprintf("渠道未确认关单，额度继续保留；系统会在%s后自动复查渠道，确认未付款再释放", deadline.Add(customCollectionReleaseMargin).Format("2006-01-02 15:04 MST")))
	}
	return infraerrors.Conflict("COLLECTION_CLOSE_UNCONFIRMED", "未能确认渠道关单，额度继续保留；请检查支付渠道或联系客服核实")
}

// 渠道确认关闭后才释放。回调先到时 paid_at 非空，不能释放已收款。
func (s *PaymentService) customReleaseCollection(ctx context.Context, id int64) error {
	result, err := s.entClient.ExecContext(ctx, `UPDATE custom_payment_collection_ledger SET released=true WHERE order_id=$1 AND paid_at IS NULL AND NOT released`, id)
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n > 0 {
		s.writeAuditLog(ctx, id, "COLLECTION_HOLD_RELEASED", "system", map[string]any{"reason": "渠道确认关单成功"})
	}
	return nil
}

// 人工收款必须登记，参考号唯一使重复提交幂等；不会自动发放订阅。
func (s *PaymentConfigService) CustomRecordCollection(ctx context.Context, amount float64, reference string) error {
	reference = strings.TrimSpace(reference)
	if !customMoneyValid(amount) || reference == "" || len(reference) > 100 {
		return infraerrors.BadRequest("COLLECTION_RECEIPT_INVALID", "请输入有效人民币收款金额与唯一收款参考号（最多100字节）")
	}
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	// 与订单触发器同一行锁，人工入账不拒绝实际已发生的收款，超额会阻止后续下单。
	rows, err := tx.Client().QueryContext(ctx, `SELECT value FROM settings WHERE key=$1 FOR UPDATE`, customCollectionKey)
	if err != nil {
		return err
	}
	_ = rows.Close()
	_, err = tx.Client().ExecContext(ctx, `INSERT INTO custom_payment_collection_ledger(reference,amount,paid_at) VALUES($1,$2,now()) ON CONFLICT(reference) DO NOTHING`, reference, amount)
	if err != nil {
		return fmt.Errorf("record collection: %w", err)
	}
	rows, err = tx.Client().QueryContext(ctx, `SELECT amount FROM custom_payment_collection_ledger WHERE reference=$1`, reference)
	if err != nil {
		return err
	}
	var existing float64
	if rows.Next() {
		err = rows.Scan(&existing)
	}
	_ = rows.Close()
	if err != nil {
		return err
	}
	if existing != amount {
		return infraerrors.Conflict("COLLECTION_REFERENCE_EXISTS", "该参考号已登记且金额不同，请核对收款记录")
	}
	return tx.Commit()
}
