//go:build unit

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentauditlog"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

var errCustomTradeNotExist = errors.New(`alipay TradeClose: {"sub_code":"ACQ.TRADE_NOT_EXIST"}`)

// 支付宝式测试渠道：关单可失败，CustomCollectionUnpaidFinal 返回预设的终态证据。
type customReleaseAlipay struct {
	*paymentOrderLifecycleQueryProvider
	closeErr  error
	final     bool
	finalErr  error
	finalRefs []string
}

func (p *customReleaseAlipay) CancelPayment(context.Context, string) error { return p.closeErr }
func (p *customReleaseAlipay) CustomCollectionUnpaidFinal(_ context.Context, ref string) (bool, error) {
	p.finalRefs = append(p.finalRefs, ref)
	return p.final, p.finalErr
}

// 易支付式测试渠道：不能关单、没有终态复查，只有查询。
type customReleaseQueryOnly struct {
	key   string
	resp  *payment.QueryOrderResponse
	err   error
	calls int
}

func (p *customReleaseQueryOnly) Name() string        { return "query-only" }
func (p *customReleaseQueryOnly) ProviderKey() string { return p.key }
func (p *customReleaseQueryOnly) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{p.key}
}
func (p *customReleaseQueryOnly) CreatePayment(context.Context, payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	panic("unexpected call")
}
func (p *customReleaseQueryOnly) QueryOrder(context.Context, string) (*payment.QueryOrderResponse, error) {
	p.calls++
	return p.resp, p.err
}
func (p *customReleaseQueryOnly) VerifyNotification(context.Context, string, map[string]string) (*payment.PaymentNotification, error) {
	panic("unexpected call")
}
func (p *customReleaseQueryOnly) Refund(context.Context, payment.RefundRequest) (*payment.RefundResponse, error) {
	panic("unexpected call")
}

type customReleaseFixture struct {
	t      *testing.T
	ctx    context.Context
	client *dbent.Client
	user   *dbent.User
	seq    int
}

func newCustomReleaseFixture(t *testing.T) *customReleaseFixture {
	t.Helper()
	ctx := context.Background()
	client := newPaymentOrderLifecycleTestClient(t)
	_, err := client.ExecContext(ctx, `CREATE TABLE custom_payment_collection_ledger(order_id BIGINT, amount NUMERIC, paid_at TIMESTAMP, released BOOLEAN DEFAULT false, release_checked_at TIMESTAMP)`)
	require.NoError(t, err)
	user := client.User.Create().SetEmail("collection-release@test.invalid").SetPasswordHash("test").SaveX(ctx)
	return &customReleaseFixture{t: t, ctx: ctx, client: client, user: user}
}

// order 建一笔 ¥10 订单并写入未释放的收款预占。wap=true 时保存与订单一致、带 time_expire 的官方 WAP 签名地址。
func (f *customReleaseFixture) order(paymentType, status string, expires time.Time, wap bool, payURL string) *dbent.PaymentOrder {
	return f.orderWithSnapshot(paymentType, status, expires, wap, payURL, nil)
}

// snapshot 非空时写入不含实例编号的渠道快照：生产解析渠道会失败，复查只能使用注入的渠道（相当于实例已删除）。
func (f *customReleaseFixture) orderWithSnapshot(paymentType, status string, expires time.Time, wap bool, payURL string, snapshot map[string]any) *dbent.PaymentOrder {
	f.t.Helper()
	f.seq++
	tradeNo := fmt.Sprintf("release-%d", f.seq)
	expires = expires.Truncate(time.Minute)
	b := f.client.PaymentOrder.Create().SetUserID(f.user.ID).SetUserEmail(f.user.Email).SetUserName("test").
		SetAmount(10).SetPayAmount(10).SetRechargeCode(fmt.Sprintf("REL%d", f.seq)).SetOutTradeNo(tradeNo).
		SetPaymentType(paymentType).SetPaymentTradeNo("").SetOrderType(payment.OrderTypeBalance).SetStatus(status).
		SetExpiresAt(expires).SetClientIP("127.0.0.1").SetSrcHost("test.invalid")
	if wap {
		biz, err := json.Marshal(map[string]string{"out_trade_no": tradeNo, "total_amount": "10.00", "product_code": "QUICK_WAP_WAY",
			"time_expire": expires.In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04")})
		require.NoError(f.t, err)
		payURL = "https://openapi.alipay.com/gateway.do?" + url.Values{"method": {"alipay.trade.wap.pay"}, "app_id": {"test-app"}, "sign": {"fixture"}, "biz_content": {string(biz)}}.Encode()
		b.SetProviderSnapshot(map[string]any{"schema_version": 2, "provider_key": payment.TypeAlipay, "merchant_app_id": "test-app"})
	} else if snapshot != nil {
		b.SetProviderSnapshot(snapshot)
	}
	if payURL != "" {
		b.SetPayURL(payURL)
	}
	o := b.SaveX(f.ctx)
	_, err := f.client.ExecContext(f.ctx, `INSERT INTO custom_payment_collection_ledger(order_id, amount) VALUES($1, 10)`, o.ID)
	require.NoError(f.t, err)
	return o
}

func (f *customReleaseFixture) released(id int64) bool {
	f.t.Helper()
	ok, err := customCollectionHoldReleased(f.ctx, f.client, id)
	require.NoError(f.t, err)
	return ok
}

func (f *customReleaseFixture) audits(id int64, action string) int {
	f.t.Helper()
	return f.client.PaymentAuditLog.Query().Where(paymentauditlog.OrderIDEQ(strconv.FormatInt(id, 10)), paymentauditlog.ActionEQ(action)).CountX(f.ctx)
}

func TestCustomCollectionPolicyHoldDefaults(t *testing.T) {
	old := customParseCollection(`{"enabled":true,"single_max":20,"daily_max":1000,"timezone":"Asia/Shanghai","quick_amounts":[10]}`)
	require.Equal(t, 60.0, old.UserHoldMax, "旧配置缺字段：单笔上限的3倍")
	require.Equal(t, 24, old.UnprovenReleaseHours)
	tight := customParseCollection(`{"enabled":true,"single_max":50,"daily_max":100,"timezone":"Asia/Shanghai","quick_amounts":[10]}`)
	require.Equal(t, 100.0, tight.UserHoldMax, "不超过每日上限")

	p := customDefaultCollection()
	require.Equal(t, 150.0, p.UserHoldMax)
	require.NoError(t, customValidateCollection(p))
	p.UserHoldMax = 40 // 低于单笔上限：用户连一笔最大额订单都下不了
	require.Error(t, customValidateCollection(p))
	p.UserHoldMax = 1001
	require.Error(t, customValidateCollection(p))
	p.UserHoldMax = 0 // 旧前端不传：保存前按当前单笔／每日上限补齐
	require.NoError(t, customValidateCollection(p))
	require.Equal(t, 150.0, p.UserHoldMax)
	for _, hours := range []int{-1, 361} {
		p.UnprovenReleaseHours = hours
		require.Error(t, customValidateCollection(p))
	}
	p.UnprovenReleaseHours = 0
	require.NoError(t, customValidateCollection(p))
	require.Equal(t, 24, p.UnprovenReleaseHours)
	require.ErrorContains(t, customCollectionError(errors.New("pq: CUSTOM_COLLECTION_USER_HOLD_LIMIT")), "未完成付款")
}

func TestCustomCollectionMobileWapExpiryGate(t *testing.T) {
	enabled := &PaymentConfig{Collection: &CustomCollectionPolicy{Enabled: true}}
	alipay := &payment.InstanceSelection{ProviderKey: payment.TypeAlipay}
	require.True(t, customShouldUseAlipayMobileWapExpiry(CreateOrderRequest{IsMobile: true}, enabled, alipay))
	require.False(t, customShouldUseAlipayMobileWapExpiry(CreateOrderRequest{IsMobile: false}, enabled, alipay), "电脑端走原有逻辑")
	require.False(t, customShouldUseAlipayMobileWapExpiry(CreateOrderRequest{IsMobile: true}, &PaymentConfig{Collection: &CustomCollectionPolicy{}}, alipay), "未启用限额不改手机参数")
	require.False(t, customShouldUseAlipayMobileWapExpiry(CreateOrderRequest{IsMobile: true}, enabled, &payment.InstanceSelection{ProviderKey: payment.TypeEasyPay}))
	precreate := &PaymentConfig{Collection: &CustomCollectionPolicy{Enabled: true}, AlipayMobilePrecreateDeepLink: true}
	require.False(t, customShouldUseAlipayMobileWapExpiry(CreateOrderRequest{IsMobile: true}, precreate, alipay), "当面付深链可关单，不走 WAP")
}

func TestCustomCollectionHoldDeadline(t *testing.T) {
	f := newCustomReleaseFixture(t)
	exp := time.Now().Add(-time.Hour)
	wait := 24 * time.Hour
	wap := f.order(payment.TypeAlipay, OrderStatusCancelled, exp, true, "")
	deadline, provable := customCollectionHoldDeadline(wap, wait)
	require.True(t, provable)
	require.True(t, deadline.Equal(wap.ExpiresAt))
	page := f.order(payment.TypeAlipay, OrderStatusExpired, exp, false, "https://openapi.alipay.com/gateway.do?method=alipay.trade.page.pay&sign=x")
	deadline, provable = customCollectionHoldDeadline(page, wait)
	require.False(t, provable, "电脑网站支付没有可校验的 time_expire")
	require.True(t, deadline.Equal(page.ExpiresAt.Add(wait)))
	failed := f.order(payment.TypeEasyPay, OrderStatusFailed, exp, false, "")
	_, provable = customCollectionHoldDeadline(failed, wait)
	require.True(t, provable, "建单失败且从未交出付款入口")
	failedWithURL := f.order(payment.TypeEasyPay, OrderStatusFailed, exp, false, "https://pay.example/submit")
	_, provable = customCollectionHoldDeadline(failedWithURL, wait)
	require.False(t, provable)
	require.False(t, customCollectionTerminalUnpaid(f.order(payment.TypeAlipay, OrderStatusPending, exp, false, "")))
}

func TestCustomCollectionReleaseBatch(t *testing.T) {
	f := newCustomReleaseFixture(t)
	now := time.Now()
	easySnapshot := map[string]any{"schema_version": 2, "provider_key": payment.TypeEasyPay}

	// 无快照的支付宝旧单走注册表：查询仍未付、关单返回交易不存在，终态复查确认未付。
	alipayReg := &customReleaseAlipay{paymentOrderLifecycleQueryProvider: &paymentOrderLifecycleQueryProvider{resp: &payment.QueryOrderResponse{Status: payment.ProviderStatusPending}}, closeErr: errCustomTradeNotExist, final: true}
	registry := payment.NewRegistry()
	registry.Register(alipayReg)
	svc := &PaymentService{entClient: f.client, registry: registry, providersLoaded: true}

	provableDue := f.order(payment.TypeAlipay, OrderStatusCancelled, now.Add(-15*time.Minute), true, "")
	provableNotDue := f.order(payment.TypeAlipay, OrderStatusCancelled, now.Add(-5*time.Minute), true, "")
	unprovableRecent := f.order(payment.TypeAlipay, OrderStatusExpired, now.Add(-2*time.Hour), false, "https://openapi.alipay.com/gateway.do?method=alipay.trade.wap.pay&sign=x")
	unprovableDue := f.order(payment.TypeAlipay, OrderStatusExpired, now.Add(-25*time.Hour), false, "https://openapi.alipay.com/gateway.do?method=alipay.trade.wap.pay&sign=x")
	unprovableOpen := f.order(payment.TypeAlipay, OrderStatusCancelled, now.Add(-26*time.Hour), false, "https://openapi.alipay.com/gateway.do?method=alipay.trade.page.pay&sign=x")
	easyDue := f.orderWithSnapshot(payment.TypeEasyPay, OrderStatusFailed, now.Add(-25*time.Hour), false, "https://pay.example/submit", easySnapshot)
	easyFailing := f.orderWithSnapshot(payment.TypeEasyPay, OrderStatusExpired, now.Add(-25*time.Hour), false, "https://pay.example/submit", easySnapshot)
	failedNoEntry := f.orderWithSnapshot(payment.TypeEasyPay, OrderStatusFailed, now.Add(-20*time.Minute), false, "", easySnapshot)
	pendingOld := f.order(payment.TypeAlipay, OrderStatusPending, now.Add(-48*time.Hour), false, "")

	openAlipay := &customReleaseAlipay{paymentOrderLifecycleQueryProvider: &paymentOrderLifecycleQueryProvider{}, final: false} // 交易仍在等待付款
	easyOK := &customReleaseQueryOnly{key: payment.TypeEasyPay, resp: &payment.QueryOrderResponse{Status: payment.ProviderStatusPending}}
	easyDown := &customReleaseQueryOnly{key: payment.TypeEasyPay, err: errors.New("upstream timeout")}
	resolve := func(_ context.Context, o *dbent.PaymentOrder) (payment.Provider, error) {
		switch o.ID {
		case unprovableOpen.ID:
			return openAlipay, nil
		case easyDue.ID, failedNoEntry.ID:
			return easyOK, nil
		case easyFailing.ID:
			return easyDown, nil
		}
		return alipayReg, nil
	}

	n, err := svc.customReleaseUnpayableCollectionHolds(f.ctx, now, resolve)
	require.NoError(t, err)
	require.Equal(t, 4, n)
	for _, o := range []*dbent.PaymentOrder{provableDue, unprovableDue, easyDue, failedNoEntry} {
		require.True(t, f.released(o.ID), "应释放订单 %d", o.ID)
		require.Equal(t, 1, f.audits(o.ID, "COLLECTION_HOLD_RELEASED"))
	}
	for _, o := range []*dbent.PaymentOrder{provableNotDue, unprovableRecent, unprovableOpen, easyFailing, pendingOld} {
		require.False(t, f.released(o.ID), "应保留订单 %d", o.ID)
	}
	require.Len(t, openAlipay.finalRefs, 1, "等待付款的交易复查后继续占用")
	require.Equal(t, 1, easyDown.calls)

	// 复查失败的订单进入退避，同一时刻再跑不会重复请求渠道。
	n, err = svc.customReleaseUnpayableCollectionHolds(f.ctx, now, resolve)
	require.NoError(t, err)
	require.Zero(t, n)
	require.Len(t, openAlipay.finalRefs, 1)
	require.Equal(t, 1, easyDown.calls)

	// 退避期过后重试；此时可证明期限的订单也已到期。
	later := now.Add(customCollectionRecheckBackoff + time.Minute)
	n, err = svc.customReleaseUnpayableCollectionHolds(f.ctx, later, resolve)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.True(t, f.released(provableNotDue.ID))
	require.False(t, f.released(unprovableRecent.ID), "无法证明期限的订单要等满等待期")
	require.Len(t, openAlipay.finalRefs, 2)
	require.Equal(t, 2, easyDown.calls)
}

func TestCustomCollectionForceRelease(t *testing.T) {
	f := newCustomReleaseFixture(t)
	svc := &PaymentService{entClient: f.client, registry: payment.NewRegistry(), providersLoaded: true}
	easySnapshot := map[string]any{"schema_version": 2, "provider_key": payment.TypeEasyPay}
	stuck := f.orderWithSnapshot(payment.TypeEasyPay, OrderStatusCancelled, time.Now().Add(-25*time.Hour), false, "https://pay.example/submit", easySnapshot)

	require.ErrorContains(t, svc.CustomForceReleaseCollection(f.ctx, stuck.ID, 9, "短", true), "原因")
	require.ErrorContains(t, svc.CustomForceReleaseCollection(f.ctx, stuck.ID, 9, "商户账单核对无此笔收款", false), "确认")
	require.False(t, f.released(stuck.ID))

	recent := f.orderWithSnapshot(payment.TypeEasyPay, OrderStatusCancelled, time.Now().Add(-2*time.Hour), false, "https://pay.example/submit", easySnapshot)
	require.ErrorContains(t, svc.CustomForceReleaseCollection(f.ctx, recent.ID, 9, "商户账单核对无此笔收款", true), "之后才能人工释放")
	pending := f.order(payment.TypeEasyPay, OrderStatusPending, time.Now().Add(-48*time.Hour), false, "")
	require.ErrorContains(t, svc.CustomForceReleaseCollection(f.ctx, pending.ID, 9, "商户账单核对无此笔收款", true), "已取消、已过期")

	// 渠道实例已删除、无法复查：管理员确认后释放，并记录操作者与原因。
	require.NoError(t, svc.CustomForceReleaseCollection(f.ctx, stuck.ID, 9, "渠道已删除，商户账单核对无此笔收款", true))
	require.True(t, f.released(stuck.ID))
	require.Equal(t, 1, f.audits(stuck.ID, "COLLECTION_HOLD_FORCE_RELEASED"))
	log := f.client.PaymentAuditLog.Query().Where(paymentauditlog.ActionEQ("COLLECTION_HOLD_FORCE_RELEASED")).OnlyX(f.ctx)
	require.Equal(t, "admin:9", log.Operator)
	require.Contains(t, log.Detail, "渠道已删除")
	require.ErrorContains(t, svc.CustomForceReleaseCollection(f.ctx, stuck.ID, 9, "重复提交同一笔订单", true), "没有可释放")
}

func TestCustomCollectionLatePaymentNeverReleased(t *testing.T) {
	f := newCustomReleaseFixture(t)
	o := f.order(payment.TypeAlipay, OrderStatusCancelled, time.Now().Add(-25*time.Hour), false, "")
	// 付款回调晚到：247/248 触发器把账本写为已收款。
	_, err := f.client.ExecContext(f.ctx, `UPDATE custom_payment_collection_ledger SET paid_at=$1 WHERE order_id=$2`, time.Now(), o.ID)
	require.NoError(t, err)
	ok, err := customMarkCollectionReleased(f.ctx, f.client, o.ID)
	require.NoError(t, err)
	require.False(t, ok)
	now := time.Now()
	ids, err := customDueCollectionHolds(f.ctx, f.client, now, 24*time.Hour, 20, now.Add(-customCollectionRecheckBackoff))
	require.NoError(t, err)
	require.NotContains(t, ids, o.ID)
}

func TestCustomCollectionPaidChannelNeverReleased(t *testing.T) {
	f := newCustomReleaseFixture(t)
	expired := time.Now().Add(-25 * time.Hour)
	auto := f.order(payment.TypeAlipay, OrderStatusCancelled, expired, false, "")
	forced := f.order(payment.TypeAlipay, OrderStatusExpired, expired, false, "")
	userRepo := &mockUserRepo{getByIDUser: &User{ID: f.user.ID, Email: f.user.Email}}
	userRepo.updateBalanceFn = func(_ context.Context, _ int64, amount float64) error {
		userRepo.getByIDUser.Balance += amount
		return nil
	}
	redeemRepo := &paymentOrderLifecycleRedeemRepo{codesByCode: map[string]*RedeemCode{
		auto.RechargeCode:   {ID: 1, Code: auto.RechargeCode, Type: RedeemTypeBalance, Value: 10, Status: StatusUnused},
		forced.RechargeCode: {ID: 2, Code: forced.RechargeCode, Type: RedeemTypeBalance, Value: 10, Status: StatusUnused},
	}}
	registry := payment.NewRegistry()
	prov := &paymentOrderLifecycleQueryProvider{resp: &payment.QueryOrderResponse{TradeNo: "late-paid", Status: payment.ProviderStatusPaid, Amount: 10}}
	registry.Register(prov)
	svc := &PaymentService{entClient: f.client, registry: registry, providersLoaded: true, userRepo: userRepo,
		redeemService: NewRedeemService(redeemRepo, userRepo, nil, nil, nil, f.client, nil, nil)}

	outcome, _ := svc.customRecheckCollectionHold(f.ctx, auto, prov, time.Now(), 24*time.Hour, "system", false)
	require.Equal(t, customHoldPaid, outcome)
	require.False(t, f.released(auto.ID))
	require.NotEqual(t, OrderStatusCancelled, f.client.PaymentOrder.GetX(f.ctx, auto.ID).Status, "晚到付款已同步，不能留在已取消")

	require.ErrorContains(t, svc.CustomForceReleaseCollection(f.ctx, forced.ID, 9, "商户账单核对无此笔收款", true), "已付款")
	require.False(t, f.released(forced.ID))
	require.Zero(t, f.audits(forced.ID, "COLLECTION_HOLD_FORCE_RELEASED"))
}
