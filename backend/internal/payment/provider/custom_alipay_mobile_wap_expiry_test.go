//go:build unit

package provider

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/smartwalle/alipay/v3"
	"github.com/stretchr/testify/require"
)

// 启用收款限额时手机 WAP 也写入东八区分钟期限；未启用时保持原手机参数。
func TestCustomAlipayMobileWapExpiry(t *testing.T) {
	oldWap := alipayTradeWapPay
	t.Cleanup(func() { alipayTradeWapPay = oldWap })
	var received alipay.TradeWapPay
	calls := 0
	alipayTradeWapPay = func(_ *alipay.Client, p alipay.TradeWapPay) (*url.URL, error) {
		received = p
		calls++
		return url.Parse("https://openapi.alipay.com/gateway.do?method=alipay.trade.wap.pay")
	}
	a := &Alipay{client: &alipay.Client{}, config: map[string]string{}}
	req := payment.CreatePaymentRequest{OrderID: "mobile-wap", Amount: "10.00", Subject: "测试", IsMobile: true,
		AlipayMobileWapExpiry: true, ExpiresAt: time.Date(2099, 1, 1, 2, 30, 0, 0, time.UTC)}
	_, err := a.CreatePayment(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "2099-01-01 10:30", received.TimeExpire)
	require.Equal(t, "QUICK_WAP_WAY", received.ProductCode)

	req.AlipayMobileWapExpiry = false
	_, err = a.CreatePayment(context.Background(), req)
	require.NoError(t, err)
	require.Empty(t, received.TimeExpire, "未启用限额不改变手机参数")

	req.AlipayMobileWapExpiry, req.ExpiresAt = true, time.Time{}
	_, err = a.CreatePayment(context.Background(), req)
	require.ErrorContains(t, err, "missing expiry", "要求期限却没有期限时拒绝建单，不发出无期限的付款入口")
	require.Equal(t, 2, calls)
}
