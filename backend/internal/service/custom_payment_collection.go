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
}

func customDefaultCollection() *CustomCollectionPolicy {
	return &CustomCollectionPolicy{SingleMax: 50, DailyMax: 1000, Timezone: "Asia/Shanghai", QuickAmounts: []float64{10, 20, 30, 40}, AllowCustomAmount: true}
}

func customParseCollection(raw string) *CustomCollectionPolicy {
	p := customDefaultCollection()
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), p); err != nil {
			return customDefaultCollection()
		}
	}
	return p
}

func customMoneyValid(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v > 0 && v <= 100000000 && math.Abs(v*100-math.Round(v*100)) < 0.000001
}

func customValidateCollection(p *CustomCollectionPolicy) error {
	if p == nil {
		return nil
	}
	if !customMoneyValid(p.SingleMax) || !customMoneyValid(p.DailyMax) || p.SingleMax > p.DailyMax {
		return infraerrors.BadRequest("COLLECTION_CONFIG_INVALID", "单笔和每日收款上限必须为正数、最多两位小数，且单笔不能超过每日上限")
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
	OrderID int64   `json:"order_id"`
	Amount  float64 `json:"amount"`
	Status  string  `json:"status"`
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
	rows, err = s.entClient.QueryContext(ctx, `SELECT l.order_id,l.amount,o.status FROM custom_payment_collection_ledger l JOIN payment_orders o ON o.id=l.order_id WHERE l.paid_at IS NULL AND NOT l.released ORDER BY l.order_id DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var h CustomCollectionHold
		if err = rows.Scan(&h.OrderID, &h.Amount, &h.Status); err != nil {
			return nil, err
		}
		u.Pending = append(u.Pending, h)
	}
	return u, rows.Err()
}

// 管理员可重新查询并关闭之前未能确认关闭的订单；不提供无证据释放按钮。
func (s *PaymentService) CustomCloseCollection(ctx context.Context, id int64) error {
	order, err := s.entClient.PaymentOrder.Get(ctx, id)
	if err != nil {
		return infraerrors.NotFound("ORDER_NOT_FOUND", "订单不存在")
	}
	if order.PaidAt != nil {
		return infraerrors.Conflict("ORDER_PAID", "订单已付款，不释放收款额度")
	}
	var result string
	if order.Status == OrderStatusPending {
		result, err = s.cancelCore(ctx, order, OrderStatusCancelled, "admin", "collection hold reconciliation")
		if err != nil {
			return err
		}
	} else {
		result = s.checkPaid(ctx, order)
	}
	if result == checkPaidResultAlreadyPaid {
		return infraerrors.Conflict("ORDER_PAID", "渠道确认订单已付款，已尝试同步付款状态")
	}
	rows, err := s.entClient.QueryContext(ctx, `SELECT released FROM custom_payment_collection_ledger WHERE order_id=$1 AND paid_at IS NULL`, id)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	released := false
	if rows.Next() {
		if err = rows.Scan(&released); err != nil {
			return err
		}
	}
	if !released {
		return infraerrors.Conflict("COLLECTION_CLOSE_UNCONFIRMED", "未能确认渠道关单，额度继续保留；请检查支付渠道或联系客服核实")
	}
	return nil
}

// 渠道确认关闭后才释放。回调先到时 paid_at 非空，不能释放已收款。
func (s *PaymentService) customReleaseCollection(ctx context.Context, id int64) error {
	_, err := s.entClient.ExecContext(ctx, `UPDATE custom_payment_collection_ledger SET released=true WHERE order_id=$1 AND paid_at IS NULL`, id)
	return err
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
