package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// [CUSTOM] 收款预占的到期复查释放。
// 原先只有渠道关单成功才释放；“交易不存在”或渠道不支持关单时永远关不掉，额度会跨日累计，
// 普通用户反复下单再取消即可占满全站额度。这里给每笔已结束、未付款的订单确定“最晚可付款时间”：
//   - 能由付款入口证明期限：支付宝 WAP 签名地址带 time_expire 且与订单期限一致，或建单失败、从未交出付款入口；
//   - 无法证明：订单期限再加管理员配置的等待时间（默认24小时）。
//
// 过了最晚可付款时间再加余量，向渠道复查，确认未付款才释放；渠道确认已付款则同步付款，绝不释放。
const (
	customCollectionReleaseMargin  = 10 * time.Minute // 吸收本站与渠道的时钟差和通知延迟
	customCollectionRecheckBackoff = 30 * time.Minute // 同一笔复查失败后的最短重试间隔，防止批量被少数订单占满
	customCollectionReleaseBatch   = 20               // 每轮最多复查的订单数，控制渠道查询量
)

// customCollectionUnpaidChecker 由渠道实现：只读查询，确认交易不存在或已关闭且从未付款。
type customCollectionUnpaidChecker interface {
	CustomCollectionUnpaidFinal(ctx context.Context, outTradeNo string) (bool, error)
}

type customCollectionSQL interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type customHoldOutcome int

const (
	customHoldKept     customHoldOutcome = iota // 不符合条件或渠道未确认，继续占用
	customHoldNotDue                            // 尚未超过最晚可付款时间加余量
	customHoldPaid                              // 渠道确认已付款，已尝试同步付款状态
	customHoldReleased                          // 已释放（本次或之前已由渠道确认关单）
)

// 启用收款限额时，手机官方支付宝走 WAP（未开当面付深链）也写入绝对期限，到期后可证明付款入口失效。
// 只在限额启用时改变手机网关参数，未启用限额的站点保持既有手机流程。
func customShouldUseAlipayMobileWapExpiry(req CreateOrderRequest, cfg *PaymentConfig, sel *payment.InstanceSelection) bool {
	return cfg != nil && cfg.Collection != nil && cfg.Collection.Enabled && req.IsMobile && sel != nil &&
		strings.EqualFold(strings.TrimSpace(sel.ProviderKey), payment.TypeAlipay) &&
		!shouldUseAlipayMobilePrecreate(req, cfg, sel)
}

func customCollectionTerminalUnpaid(o *dbent.PaymentOrder) bool {
	if o == nil || o.PaidAt != nil {
		return false
	}
	switch o.Status {
	case OrderStatusExpired, OrderStatusCancelled, OrderStatusFailed:
		return true
	}
	return false
}

// customCollectionHoldDeadline 返回最晚可付款时间，以及它是否由付款入口证明。
func customCollectionHoldDeadline(o *dbent.PaymentOrder, unproven time.Duration) (time.Time, bool) {
	if o.Status == OrderStatusFailed && strings.TrimSpace(psStringValue(o.PayURL)) == "" && strings.TrimSpace(psStringValue(o.QrCode)) == "" {
		return o.ExpiresAt, true // 建单失败且从未保存付款地址或二维码：付款入口没有交给用户。
	}
	// 只认与本订单商户、金额、订单号、绝对期限都一致的官方签名 WAP 地址（电脑扫码与手机 WAP 共用校验）。
	if validateAlipayDesktopWapURL(o, psStringValue(o.PayURL)) == nil {
		return o.ExpiresAt, true
	}
	return o.ExpiresAt.Add(unproven), false
}

// customCollectionChannelUnpaid 向渠道确认这笔订单不会再收款。
// 支付宝只认“交易不存在”或“已关闭且从未打款”；其他渠道没有终态证据，只在等待期之后接受查询成功且未付款。
func customCollectionChannelUnpaid(ctx context.Context, o *dbent.PaymentOrder, prov payment.Provider) (bool, error) {
	ref := paymentOrderQueryReference(o, prov)
	if ref == "" {
		return false, fmt.Errorf("missing query reference")
	}
	if checker, ok := prov.(customCollectionUnpaidChecker); ok {
		return checker.CustomCollectionUnpaidFinal(ctx, ref)
	}
	resp, err := prov.QueryOrder(ctx, ref)
	if err != nil {
		return false, err
	}
	if resp == nil {
		return false, fmt.Errorf("empty query response")
	}
	return resp.Status != payment.ProviderStatusPaid && resp.Status != payment.ProviderStatusSuccess, nil
}

// customRecheckCollectionHold 复查一笔已结束、未付款订单的预占。
// channelChecked 为 false 时先按原流程查询渠道并尝试关单：查到已付款就同步付款，关单成功即释放。
// 未能关单时，只有超过最晚可付款时间加余量、且渠道确认未付款才释放并写审计。
func (s *PaymentService) customRecheckCollectionHold(ctx context.Context, o *dbent.PaymentOrder, prov payment.Provider, now time.Time, unproven time.Duration, operator string, channelChecked bool) (customHoldOutcome, time.Time) {
	if !customCollectionTerminalUnpaid(o) || prov == nil {
		return customHoldKept, time.Time{}
	}
	deadline, provable := customCollectionHoldDeadline(o, unproven)
	if !channelChecked && s.checkPaid(ctx, o) == checkPaidResultAlreadyPaid {
		return customHoldPaid, deadline
	}
	released, err := customCollectionHoldReleased(ctx, s.entClient, o.ID)
	if err != nil {
		slog.Warn("collection hold read failed", "orderID", o.ID, "error", err)
		return customHoldKept, deadline
	}
	if released {
		return customHoldReleased, deadline
	}
	if now.Before(deadline.Add(customCollectionReleaseMargin)) {
		return customHoldNotDue, deadline
	}
	unpaid, err := customCollectionChannelUnpaid(ctx, o, prov)
	if err != nil {
		slog.Warn("collection hold recheck failed", "orderID", o.ID, "error", err)
		return customHoldKept, deadline
	}
	if !unpaid {
		return customHoldKept, deadline
	}
	ok, err := customMarkCollectionReleased(ctx, s.entClient, o.ID)
	if err != nil || !ok {
		if err != nil {
			slog.Error("release collection hold failed", "orderID", o.ID, "error", err)
		}
		return customHoldKept, deadline
	}
	reason := "最晚可付款时间已过，渠道复查确认未付款"
	if !provable {
		reason = "无法由付款入口证明期限，已超过等待期且渠道复查确认未付款"
	}
	s.writeAuditLog(ctx, o.ID, "COLLECTION_HOLD_RELEASED", operator, map[string]any{
		"reason":   reason,
		"deadline": deadline.UTC().Format(time.RFC3339),
		"provable": provable,
	})
	return customHoldReleased, deadline
}

// CustomForceReleaseCollection 是管理员的人工释放：渠道已删除、长期查询失败等无法复查的情况才使用。
// 必须写明原因并确认；只针对已结束、未付款的订单，并且要过订单期限＋等待期＋余量。
// 渠道仍能查到已付款时拒绝并同步付款。每次释放都会写审计，记录管理员和原因。
func (s *PaymentService) CustomForceReleaseCollection(ctx context.Context, id, adminID int64, reason string, confirmed bool) error {
	reason = strings.TrimSpace(reason)
	if !confirmed || utf8.RuneCountInString(reason) < 3 || utf8.RuneCountInString(reason) > 500 {
		return infraerrors.BadRequest("COLLECTION_RELEASE_REASON", "请填写3至500字的释放原因，并确认已核对商户账单该订单未收款")
	}
	order, err := s.entClient.PaymentOrder.Get(ctx, id)
	if err != nil {
		return infraerrors.NotFound("ORDER_NOT_FOUND", "订单不存在")
	}
	if !customCollectionTerminalUnpaid(order) {
		return infraerrors.Conflict("COLLECTION_RELEASE_NOT_ALLOWED", "只有已取消、已过期或建单失败且未付款的订单可以人工释放")
	}
	if released, readErr := customCollectionHoldReleased(ctx, s.entClient, id); readErr != nil {
		return readErr
	} else if released {
		return infraerrors.Conflict("COLLECTION_RELEASE_NOT_ALLOWED", "该订单没有可释放的收款占用，可能已释放")
	}
	wait, err := s.customCollectionUnprovenWindow(ctx)
	if err != nil {
		return err
	}
	if after := order.ExpiresAt.Add(wait + customCollectionReleaseMargin); time.Now().Before(after) {
		return infraerrors.Conflict("COLLECTION_RELEASE_TOO_EARLY", fmt.Sprintf("付款入口可能仍有效，%s 之后才能人工释放", after.Format("2006-01-02 15:04 MST")))
	}
	// 渠道还能查询时先查一次：已付款就同步，不释放。查询失败或无法连接渠道，由管理员的确认负责。
	if s.checkPaid(ctx, order) == checkPaidResultAlreadyPaid {
		return infraerrors.Conflict("ORDER_PAID", "渠道确认订单已付款，已尝试同步付款状态，不能释放")
	}
	ok, err := customMarkCollectionReleased(ctx, s.entClient, id)
	if err != nil {
		return err
	}
	if !ok {
		if released, readErr := customCollectionHoldReleased(ctx, s.entClient, id); readErr == nil && released {
			return nil // 刚才的查询已由渠道确认关单并释放。
		}
		return infraerrors.Conflict("COLLECTION_RELEASE_NOT_ALLOWED", "该订单没有可释放的收款占用，或已付款")
	}
	s.writeAuditLog(ctx, id, "COLLECTION_HOLD_FORCE_RELEASED", fmt.Sprintf("admin:%d", adminID), map[string]any{
		"reason":   reason,
		"deadline": order.ExpiresAt.Add(wait).UTC().Format(time.RFC3339),
	})
	return nil
}

func customCollectionUnprovenWindowOf(p *CustomCollectionPolicy) time.Duration {
	hours := customCollectionDefaultUnprovenHours
	if p != nil && p.UnprovenReleaseHours > 0 {
		hours = p.UnprovenReleaseHours
	}
	return time.Duration(hours) * time.Hour
}

// 读不到配置时报错而不是退回默认值，避免把管理员调长的等待期悄悄缩短。
func (s *PaymentService) customCollectionUnprovenWindow(ctx context.Context) (time.Duration, error) {
	wait, _, err := s.customCollectionReleasePolicy(ctx)
	return wait, err
}

func (s *PaymentService) customCollectionReleasePolicy(ctx context.Context) (time.Duration, bool, error) {
	if s.configService == nil {
		return customCollectionUnprovenWindowOf(nil), false, nil
	}
	cfg, err := s.configService.GetPaymentConfig(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("load collection policy: %w", err)
	}
	return customCollectionUnprovenWindowOf(cfg.Collection), cfg.Collection != nil && cfg.Collection.MissingProviderAutoRelease, nil
}

// CustomReleaseUnpayableCollectionHolds 由订单过期任务定期调用，复查已结束订单的收款预占。
func (s *PaymentService) CustomReleaseUnpayableCollectionHolds(ctx context.Context) (int, error) {
	return s.customReleaseUnpayableCollectionHolds(ctx, time.Now(), s.getOrderProvider)
}

func (s *PaymentService) customReleaseUnpayableCollectionHolds(ctx context.Context, now time.Time, resolve func(context.Context, *dbent.PaymentOrder) (payment.Provider, error)) (int, error) {
	return s.customSweepCollectionHolds(ctx, now, resolve, now.Add(-customCollectionRecheckBackoff), "system")
}

func (s *PaymentService) CustomRecheckCollectionHolds(ctx context.Context, adminID int64) (int, error) {
	now := time.Now()
	return s.customSweepCollectionHolds(ctx, now, s.getOrderProvider, now, fmt.Sprintf("admin:%d", adminID))
}

func (s *PaymentService) customSweepCollectionHolds(ctx context.Context, now time.Time, resolve func(context.Context, *dbent.PaymentOrder) (payment.Provider, error), checkedBefore time.Time, operator string) (int, error) {
	if s == nil || s.entClient == nil {
		return 0, nil
	}
	wait, autoReleaseMissing, err := s.customCollectionReleasePolicy(ctx)
	if err != nil {
		return 0, err
	}
	ids, err := customDueCollectionHolds(ctx, s.entClient, now, wait, customCollectionReleaseBatch, checkedBefore)
	if err != nil {
		return 0, fmt.Errorf("list collection holds: %w", err)
	}
	released := 0
	for _, id := range ids {
		if ctx.Err() != nil {
			break
		}
		// 先记复查时间，失败的订单按退避间隔重试，不会每轮都占满批次。
		if err := customMarkCollectionChecked(ctx, s.entClient, id, now); err != nil {
			slog.Warn("mark collection hold checked failed", "orderID", id, "error", err)
		}
		o, err := s.entClient.PaymentOrder.Get(ctx, id)
		if err != nil {
			slog.Warn("load collection hold order failed", "orderID", id, "error", err)
			continue
		}
		prov, err := resolve(ctx, o)
		if err != nil {
			if autoReleaseMissing {
				ok, releaseErr := s.customAutoReleaseMissingProviderHold(ctx, o, now, wait, operator)
				if releaseErr != nil {
					slog.Warn("auto release missing provider collection hold failed", "orderID", id, "error", releaseErr)
				}
				if ok {
					released++
					continue
				}
			}
			slog.Warn("resolve collection hold provider failed", "orderID", id, "error", err)
			continue
		}
		if outcome, _ := s.customRecheckCollectionHold(ctx, o, prov, now, wait, operator, false); outcome == customHoldReleased {
			released++
		}
	}
	return released, nil
}

// customDueCollectionHolds 粗筛可能到期的预占：已结束未付、未释放、未在退避期内复查过。
// 可证明期限的订单（建单失败无入口，或地址带 time_expire）在期限加余量后即可复查；
// 其他订单等到期限＋等待期＋余量。最终资格仍由 customCollectionHoldDeadline 判定。
func customDueCollectionHolds(ctx context.Context, db customCollectionSQL, now time.Time, unproven time.Duration, limit int, checkedBefore time.Time) ([]int64, error) {
	rows, err := db.QueryContext(ctx, `SELECT o.id FROM payment_orders o
 JOIN custom_payment_collection_ledger l ON l.order_id=o.id
 WHERE l.paid_at IS NULL AND NOT l.released AND o.paid_at IS NULL
 AND o.status IN ($1,$2,$3)
 AND (l.release_checked_at IS NULL OR l.release_checked_at <= $4)
 AND ((o.expires_at <= $5 AND ((o.status=$6 AND COALESCE(o.pay_url,'')='' AND COALESCE(o.qr_code,'')='') OR COALESCE(o.pay_url,'') LIKE $7))
  OR o.expires_at <= $8)
 ORDER BY CASE WHEN l.release_checked_at IS NULL THEN 0 ELSE 1 END, o.expires_at, o.id LIMIT $9`,
		OrderStatusExpired, OrderStatusCancelled, OrderStatusFailed,
		checkedBefore,
		now.Add(-customCollectionReleaseMargin), OrderStatusFailed, "%time_expire%",
		now.Add(-unproven-customCollectionReleaseMargin), limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	ids := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func customMarkCollectionChecked(ctx context.Context, db customCollectionSQL, orderID int64, now time.Time) error {
	_, err := db.ExecContext(ctx, `UPDATE custom_payment_collection_ledger SET release_checked_at=$1 WHERE order_id=$2`, now, orderID)
	return err
}

// 订单仍未付款且处于终态时才置 released；付款回调已写入 paid_at 时不会被覆盖。
func customMarkCollectionReleased(ctx context.Context, db customCollectionSQL, orderID int64) (bool, error) {
	result, err := db.ExecContext(ctx, `UPDATE custom_payment_collection_ledger SET released=true
 WHERE order_id=$1 AND paid_at IS NULL AND NOT released
 AND EXISTS (SELECT 1 FROM payment_orders o WHERE o.id=$2 AND o.paid_at IS NULL AND o.status IN ($3,$4,$5))`,
		orderID, orderID, OrderStatusExpired, OrderStatusCancelled, OrderStatusFailed)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n > 0, err
}

func customCollectionHoldReleased(ctx context.Context, db customCollectionSQL, orderID int64) (bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT released FROM custom_payment_collection_ledger WHERE order_id=$1 AND paid_at IS NULL`, orderID)
	if err != nil {
		return false, err
	}
	defer func() { _ = rows.Close() }()
	released := false
	if rows.Next() {
		if err := rows.Scan(&released); err != nil {
			return false, err
		}
	}
	return released, rows.Err()
}
