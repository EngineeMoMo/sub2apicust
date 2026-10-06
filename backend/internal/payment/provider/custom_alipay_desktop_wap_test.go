//go:build unit

package provider

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/smartwalle/alipay/v3"
	"github.com/stretchr/testify/require"
)

func TestCustomAlipayDesktopWapRouting(t *testing.T) {
	oldWap, oldPre, oldPage := alipayTradeWapPay, alipayTradePreCreate, alipayTradePagePay
	t.Cleanup(func() { alipayTradeWapPay, alipayTradePreCreate, alipayTradePagePay = oldWap, oldPre, oldPage })
	var received alipay.TradeWapPay
	alipayTradeWapPay = func(_ *alipay.Client, p alipay.TradeWapPay) (*url.URL, error) {
		received = p
		return url.Parse("https://openapi.alipay.com/gateway.do?method=alipay.trade.wap.pay")
	}
	pre, page := 0, 0
	alipayTradePreCreate = func(context.Context, *alipay.Client, alipay.TradePreCreate) (*alipay.TradePreCreateRsp, error) {
		pre++
		return &alipay.TradePreCreateRsp{Error: alipay.Error{Code: alipay.CodeSuccess}, QRCode: "precreate-qr"}, nil
	}
	alipayTradePagePay = func(*alipay.Client, alipay.TradePagePay) (*url.URL, error) {
		page++
		return url.Parse("https://openapi.alipay.com/gateway.do?page-pay")
	}
	a := &Alipay{client: &alipay.Client{}, config: map[string]string{"paymentMode": "redirect"}}
	req := payment.CreatePaymentRequest{OrderID: "custom-wap", Amount: "10.00", Subject: "测试", AlipayDesktopWapQRCode: true,
		ExpiresAt: time.Date(2099, 1, 1, 2, 30, 0, 0, time.UTC), NotifyURL: "https://merchant.example/notify", ReturnURL: "https://merchant.example/payment/result"}
	r, err := a.CreatePayment(context.Background(), req)
	require.NoError(t, err)
	require.NotEmpty(t, r.PayURL)
	require.Empty(t, r.QRCode) // 本站短二维码由订单服务生成。
	require.Zero(t, pre)
	require.Zero(t, page)
	require.Equal(t, "2099-01-01 10:30", received.TimeExpire)
	require.Equal(t, "QUICK_WAP_WAY", received.ProductCode)
	require.Equal(t, req.NotifyURL, received.NotifyURL)
	require.Equal(t, req.ReturnURL, received.ReturnURL)
	require.Equal(t, req.Amount, received.TotalAmount)

	// 即使内部请求错误地携带桌面开关，手机仍优先使用原 WAP/当面付设置。
	req.IsMobile = true
	_, err = a.CreatePayment(context.Background(), req)
	require.NoError(t, err)
	require.Empty(t, received.TimeExpire)
	req.AlipayMobilePrecreate = true
	r, err = a.CreatePayment(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "precreate-qr", r.QRCode)
	require.Equal(t, 1, pre)

	req.IsMobile, req.AlipayDesktopWapQRCode, req.AlipayMobilePrecreate = false, false, false
	_, err = a.CreatePayment(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 1, page)

	req.AlipayDesktopWapQRCode = true
	alipayTradeWapPay = func(*alipay.Client, alipay.TradeWapPay) (*url.URL, error) { return nil, errors.New("wap failed") }
	_, err = a.CreatePayment(context.Background(), req)
	require.ErrorContains(t, err, "wap failed")
	require.Equal(t, 1, pre)
	require.Equal(t, 1, page) // WAP 失败也不能回退到未开通的接口。
	req.ExpiresAt = time.Time{}
	_, err = a.CreatePayment(context.Background(), req)
	require.ErrorContains(t, err, "missing expiry")
}
