//go:build unit

package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func customWapFixture(t *testing.T) *dbent.PaymentOrder {
	t.Helper()
	exp := time.Now().Add(30 * time.Minute).Truncate(time.Minute)
	return &dbent.PaymentOrder{ID: 123, OutTradeNo: "sub2_custom_wap", PayAmount: 10, Status: OrderStatusPending,
		ExpiresAt: exp, ProviderKey: paymentConfigStrPtr(payment.TypeAlipay), ProviderSnapshot: map[string]any{
			"provider_key": payment.TypeAlipay, "merchant_app_id": "test-app", customAlipayDesktopWapSnapshotKey: true,
		}}
}

func customWapFixtureURL(t *testing.T, order *dbent.PaymentOrder) string {
	t.Helper()
	biz, err := json.Marshal(map[string]string{"out_trade_no": order.OutTradeNo, "total_amount": "10.00", "product_code": "QUICK_WAP_WAY",
		"time_expire": order.ExpiresAt.In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04")})
	require.NoError(t, err)
	return "https://openapi.alipay.com/gateway.do?" + url.Values{"method": {"alipay.trade.wap.pay"}, "app_id": {"test-app"}, "sign": {"fixture"}, "biz_content": {string(biz)}}.Encode()
}

func TestCustomAlipayDesktopWapConfig(t *testing.T) {
	repo := &paymentConfigSettingRepoStub{values: map[string]string{SettingAlipayForceQRCode: "false", SettingAlipayMobilePrecreateDeepLink: "true"}}
	svc := &PaymentConfigService{settingRepo: repo, entClient: newPaymentConfigServiceTestClient(t)}
	cfg, err := svc.GetPaymentConfig(context.Background())
	require.NoError(t, err)
	require.False(t, cfg.AlipayDesktopWapQRCode)
	on := true
	require.NoError(t, svc.UpdatePaymentConfig(context.Background(), UpdatePaymentConfigRequest{AlipayDesktopWapQRCode: &on}))
	cfg, err = svc.GetPaymentConfig(context.Background())
	require.NoError(t, err)
	require.True(t, cfg.AlipayDesktopWapQRCode)
	require.NoError(t, svc.UpdatePaymentConfig(context.Background(), UpdatePaymentConfigRequest{}))
	require.Equal(t, "true", repo.values[SettingAlipayDesktopWapQRCode])
	require.Equal(t, "true", repo.values[SettingAlipayMobilePrecreateDeepLink])
	off := false
	require.NoError(t, svc.UpdatePaymentConfig(context.Background(), UpdatePaymentConfigRequest{AlipayDesktopWapQRCode: &off}))
	require.False(t, svc.parsePaymentConfig(repo.values).AlipayDesktopWapQRCode)
	for _, tc := range []struct {
		mobile   bool
		provider string
		want     bool
	}{
		{false, payment.TypeAlipay, true}, {true, payment.TypeAlipay, false}, {false, payment.TypeEasyPay, false},
		{false, payment.TypeStripe, false}, {false, payment.TypeWxpay, false},
	} {
		require.Equal(t, tc.want, shouldUseAlipayDesktopWapQRCode(CreateOrderRequest{IsMobile: tc.mobile}, &PaymentConfig{AlipayDesktopWapQRCode: true}, &payment.InstanceSelection{ProviderKey: tc.provider}))
	}
	require.False(t, shouldUseAlipayDesktopWapQRCode(CreateOrderRequest{}, &PaymentConfig{}, &payment.InstanceSelection{ProviderKey: payment.TypeAlipay}))
}

func TestCustomAlipayDesktopWapRejectsUnsafeTargets(t *testing.T) {
	order := customWapFixture(t)
	raw := customWapFixtureURL(t, order)
	require.NoError(t, validateAlipayDesktopWapURL(order, raw))
	for _, change := range []func(*url.URL){
		func(u *url.URL) { u.Host = "evil.example" },
		func(u *url.URL) { u.Host = "openapi.alipay.com.evil.example" },
		func(u *url.URL) { u.Host = "openapi.alipay.com:444" },
		func(u *url.URL) { u.User = url.User("evil") },
		func(u *url.URL) { u.Scheme = "http" },
		func(u *url.URL) { u.Path = "/other" },
		func(u *url.URL) { u.Fragment = "other" },
		func(u *url.URL) { q := u.Query(); q.Set("method", "alipay.trade.page.pay"); u.RawQuery = q.Encode() },
		func(u *url.URL) { q := u.Query(); q.Set("app_id", "other-app"); u.RawQuery = q.Encode() },
		func(u *url.URL) { q := u.Query(); q.Del("sign"); u.RawQuery = q.Encode() },
		func(u *url.URL) {
			q := u.Query()
			q.Set("biz_content", strings.ReplaceAll(q.Get("biz_content"), "10.00", "99.00"))
			u.RawQuery = q.Encode()
		},
		func(u *url.URL) {
			q := u.Query()
			q.Set("biz_content", strings.ReplaceAll(q.Get("biz_content"), order.OutTradeNo, "other-order"))
			u.RawQuery = q.Encode()
		},
		func(u *url.URL) { q := u.Query(); q.Set("biz_content", "{}"); u.RawQuery = q.Encode() },
	} {
		u, err := url.Parse(raw)
		require.NoError(t, err)
		change(u)
		require.Error(t, validateAlipayDesktopWapURL(order, u.String()))
	}
	pr := &payment.CreatePaymentResponse{PayURL: raw}
	require.Error(t, prepareAlipayDesktopWapQRCode(order, "", pr))
	require.NoError(t, prepareAlipayDesktopWapQRCode(order, "https://merchant.example/payment/result?old=query#fragment", pr))
	u, err := url.Parse(pr.QRCode)
	require.NoError(t, err)
	require.Equal(t, "merchant.example", u.Host)
	require.Equal(t, customAlipayDesktopWapPath+"123", u.Path)
	require.Empty(t, u.Fragment)
	require.Len(t, u.Query(), 1)
	token := u.Query().Get("token")
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(token)
	require.NoError(t, err)
	require.Len(t, decoded, 32)
	require.NoError(t, prepareAlipayDesktopWapQRCode(order, "https://merchant.example/payment/result", pr))
	require.NotContains(t, pr.QRCode, token)
}

// 真实 SDK 本地生成签名，不向支付宝发网络请求，不创建商户交易。
func TestCustomAlipayDesktopWapSignedOrderAndScan(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	public, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	for _, scenario := range []struct {
		name, orderType       string
		credited, paid, bonus float64
	}{
		{"balance", payment.OrderTypeBalance, 10, 10, 0},
		{"subscription", payment.OrderTypeSubscription, 10, 10, 0},
		{"balance_bonus", payment.OrderTypeBalance, 12, 10, 2},
		{"balance_discount", payment.OrderTypeBalance, 10, 8, 0},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			orderType := scenario.orderType
			paidText := strconv.FormatFloat(scenario.paid, 'f', 2, 64)
			ctx := context.Background()
			client := newPaymentConfigServiceTestClient(t)
			user, err := client.User.Create().SetEmail("wap@example.test").SetPasswordHash("hash").SetUsername("wap").Save(ctx)
			require.NoError(t, err)
			sel := &payment.InstanceSelection{ProviderKey: payment.TypeAlipay, InstanceID: "1", PaymentMode: "redirect", Config: map[string]string{
				"appId": "test-app", "privateKey": base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PrivateKey(key)), "publicKey": base64.StdEncoding.EncodeToString(public),
				"notifyUrl": "https://merchant.example/api/v1/payment/webhook/alipay",
			}}
			cfg := &PaymentConfig{AlipayDesktopWapQRCode: true, OrderTimeoutMin: 30, MaxPendingOrders: 3}
			svc := &PaymentService{entClient: client, resumeService: NewPaymentResumeService([]byte("test-resume-signing-key"))}
			req := CreateOrderRequest{UserID: user.ID, PaymentType: payment.TypeAlipay, OrderType: orderType, SrcHost: "merchant.example", ReturnURL: "https://merchant.example/payment/result"}
			order, err := svc.createOrderInTx(ctx, req, &User{ID: user.ID, Email: user.Email, Username: user.Username}, nil, cfg, scenario.credited, scenario.paid, 0, scenario.paid, scenario.bonus, sel)
			require.NoError(t, err)
			require.Equal(t, scenario.credited, order.Amount)
			require.Equal(t, scenario.bonus, order.BonusAmount)
			require.Zero(t, order.ExpiresAt.Second())
			require.Equal(t, true, order.ProviderSnapshot[customAlipayDesktopWapSnapshotKey])
			require.Equal(t, "qrcode", order.ProviderSnapshot["payment_mode"])
			resp, err := svc.invokeProvider(ctx, order, req, cfg, scenario.paid, paidText, scenario.paid, nil, sel)
			require.NoError(t, err)
			require.Equal(t, "qrcode", resp.PaymentMode)
			require.Equal(t, "redirect", sel.PaymentMode)
			require.NotEmpty(t, resp.ResumeToken)
			u, err := url.Parse(resp.QRCode)
			require.NoError(t, err)
			require.Equal(t, customAlipayDesktopWapPath+strconv.FormatInt(order.ID, 10), u.Path)
			token := u.Query().Get("token")
			cfg.AlipayDesktopWapQRCode = false // 已保存的订单不读取当前开关。
			payURL, err := svc.ResolveAlipayDesktopWapURL(ctx, order.ID, token)
			require.NoError(t, err)
			require.Equal(t, resp.PayURL, payURL)
			gatewayURL, err := url.Parse(payURL)
			require.NoError(t, err)
			var signedBiz map[string]any
			require.NoError(t, json.Unmarshal([]byte(gatewayURL.Query().Get("biz_content")), &signedBiz))
			require.Equal(t, paidText, signedBiz["total_amount"], "签名和扫码校验必须使用实付金额，不使用含赠金的到账额")
			for _, invalid := range []string{"", token[:42], strings.Repeat("A", 43)} {
				_, err = svc.ResolveAlipayDesktopWapURL(ctx, order.ID, invalid)
				require.Equal(t, "INVALID_ALIPAY_WAP_TOKEN", infraerrors.Reason(err))
			}
			_, err = svc.ResolveAlipayDesktopWapURL(ctx, order.ID+1, token)
			require.Equal(t, "INVALID_ALIPAY_WAP_TOKEN", infraerrors.Reason(err))
			cfg.AlipayDesktopWapQRCode = true
			other, err := svc.createOrderInTx(ctx, req, &User{ID: user.ID, Email: user.Email, Username: user.Username}, nil, cfg, 10, 10, 0, 10, 0, sel)
			require.NoError(t, err)
			otherResp, err := svc.invokeProvider(ctx, other, req, cfg, 10, "10.00", 10, nil, sel)
			require.NoError(t, err)
			cfg.AlipayDesktopWapQRCode = false
			_, err = svc.ResolveAlipayDesktopWapURL(ctx, other.ID, token)
			require.Equal(t, "INVALID_ALIPAY_WAP_TOKEN", infraerrors.Reason(err))
			otherURL, err := url.Parse(otherResp.QRCode)
			require.NoError(t, err)
			_, err = svc.ResolveAlipayDesktopWapURL(ctx, order.ID, otherURL.Query().Get("token"))
			require.Equal(t, "INVALID_ALIPAY_WAP_TOKEN", infraerrors.Reason(err))
			_, err = client.PaymentOrder.UpdateOneID(order.ID).SetPayURL(strings.Replace(payURL, "openapi.alipay.com", "evil.example", 1)).Save(ctx)
			require.NoError(t, err)
			_, err = svc.ResolveAlipayDesktopWapURL(ctx, order.ID, token)
			require.Equal(t, "INVALID_ALIPAY_WAP_URL", infraerrors.Reason(err))
			_, err = client.PaymentOrder.UpdateOneID(order.ID).SetPayURL(payURL).Save(ctx)
			require.NoError(t, err)
			for _, state := range []string{OrderStatusPaid, OrderStatusCompleted, OrderStatusCancelled, OrderStatusExpired, OrderStatusFailed} {
				_, err = client.PaymentOrder.UpdateOneID(order.ID).SetStatus(state).Save(ctx)
				require.NoError(t, err)
				_, err = svc.ResolveAlipayDesktopWapURL(ctx, order.ID, token)
				require.Equal(t, "ALIPAY_WAP_ORDER_UNAVAILABLE", infraerrors.Reason(err))
			}
			_, err = client.PaymentOrder.UpdateOneID(order.ID).SetStatus(OrderStatusPending).SetExpiresAt(time.Now().Add(-time.Second)).Save(ctx)
			require.NoError(t, err)
			_, err = svc.ResolveAlipayDesktopWapURL(ctx, order.ID, token)
			require.Equal(t, "ALIPAY_WAP_ORDER_EXPIRED", infraerrors.Reason(err))
		})
	}
}
