//go:build unit

package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestCustomAlipayDesktopWapSecurityAmbiguousURL(t *testing.T) {
	order := customWapFixture(t)
	raw := customWapFixtureURL(t, order)
	for _, key := range []string{"method", "app_id", "sign", "biz_content"} {
		t.Run(key, func(t *testing.T) {
			u, err := url.Parse(raw)
			require.NoError(t, err)
			q := u.Query()
			q.Add(key, "other-value")
			u.RawQuery = q.Encode()
			require.Error(t, validateAlipayDesktopWapURL(order, u.String()))
		})
	}
	u, err := url.Parse(raw)
	require.NoError(t, err)
	q := u.Query()
	q.Set("biz_content", strings.Replace(q.Get("biz_content"), "{", `{"out_trade_no":"other-order",`, 1))
	u.RawQuery = q.Encode()
	require.Error(t, validateAlipayDesktopWapURL(order, u.String()), "重复 JSON 键不能按最后一个值绕过")
}

func TestCustomAlipayDesktopWapSecurityConcurrentLease(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	order := createPaymentFulfillmentSubscriptionOrder(t, ctx, client, OrderStatusPaid, time.Now())
	svc := &PaymentService{entClient: client}
	start := make(chan struct{})
	winners := make(chan *paymentFulfillmentLease, 16)
	errors := make(chan error, 16)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			lease, err := svc.acquirePaymentFulfillmentLease(ctx, order)
			if err != nil {
				errors <- err
			} else if lease != nil {
				winners <- lease
			}
		}()
	}
	close(start)
	wg.Wait()
	close(winners)
	close(errors)
	require.Len(t, winners, 1, "并发处理同一订单只允许一个入账租约")
	for err := range errors {
		require.True(t, strings.Contains(err.Error(), "CONFLICT") || strings.Contains(err.Error(), "being processed") ||
			strings.Contains(err.Error(), "SQLITE_LOCKED") || strings.Contains(err.Error(), "SQLITE_BUSY"), fmt.Sprint(err))
	}
	lease := <-winners
	require.NoError(t, svc.markCompleted(ctx, order, lease, "SUBSCRIPTION_SUCCESS"))
	reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusCompleted, reloaded.Status)
}

func FuzzCustomAlipayDesktopWapSecurityHost(f *testing.F) {
	for _, host := range []string{"openapi.alipay.com", "evil.example", "openapi.alipay.com.evil.example", "openapi.alipay.com:443", "openapi.alipay.com@evil.example", "openapi.alipay.com\\@evil.example"} {
		f.Add(host)
	}
	// 模糊测试不签名、不发网络请求，只验证攻击者控制 host 时不能绕过白名单。
	f.Fuzz(func(t *testing.T, host string) {
		if len(host) > 4096 {
			return
		}
		order := customWapFixture(t)
		u, err := url.Parse(customWapFixtureURL(t, order))
		require.NoError(t, err)
		u.Host = host
		if validateAlipayDesktopWapURL(order, u.String()) == nil {
			require.Equal(t, "openapi.alipay.com", host)
		}
	})
}

func TestCustomAlipayDesktopWapSecurityRequiresSecureOrigin(t *testing.T) {
	order := customWapFixture(t)
	pr := &payment.CreatePaymentResponse{PayURL: customWapFixtureURL(t, order)}
	require.Error(t, prepareAlipayDesktopWapQRCode(order, "http://merchant.example/payment/result", pr))
	for _, origin := range []string{"https://merchant.example", "http://localhost:8080", "http://127.0.0.1:8080", "http://[::1]:8080"} {
		require.NoError(t, prepareAlipayDesktopWapQRCode(order, origin+"/payment/result", pr))
	}
}

func TestCustomAlipayDesktopWapSecurityReturnURLTrust(t *testing.T) {
	ctx := context.Background()
	svc := &PaymentService{}
	req := CreateOrderRequest{SrcHost: "merchant.example", SrcURL: "https://evil.example/page", ReturnURL: "https://evil.example/payment/result"}
	_, err := svc.canonicalAlipayDesktopWapReturnURL(ctx, req)
	require.Error(t, err, "伪造 Referer 不能放行站外二维码")
	req.ReturnURL = "https://merchant.example/payment/result"
	_, err = svc.canonicalAlipayDesktopWapReturnURL(ctx, req)
	require.NoError(t, err)
	req.ReturnURL = "https://user@merchant.example/payment/result"
	_, err = svc.canonicalAlipayDesktopWapReturnURL(ctx, req)
	require.Error(t, err)
	repo := &paymentConfigSettingRepoStub{values: map[string]string{SettingKeyFrontendURL: "https://front.example"}}
	svc.configService = &PaymentConfigService{settingRepo: repo}
	req.ReturnURL = "https://front.example/payment/result"
	_, err = svc.canonicalAlipayDesktopWapReturnURL(ctx, req)
	require.NoError(t, err, "允许管理员明确配置的前端域名")
	req.SrcHost, req.SrcURL, req.ReturnURL = "evil.example", "https://evil.example", "https://evil.example/payment/result"
	_, err = svc.canonicalAlipayDesktopWapReturnURL(ctx, req)
	require.Error(t, err, "配置了前端 URL 后伪造 Host 也不能扩展白名单")
}
