//go:build unit

package provider

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/smartwalle/alipay/v3"
	"github.com/smartwalle/nsign"
	"github.com/stretchr/testify/require"
)

// 使用测试 RSA 密钥生成真实签名，不接触商户密钥或支付宝网络。
func customAlipaySignedNotification(t *testing.T) (*Alipay, func(url.Values) string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	client, err := alipay.New("test-app", base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PrivateKey(key)), true)
	require.NoError(t, err)
	public, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	require.NoError(t, client.LoadAliPayPublicKey(base64.StdEncoding.EncodeToString(public)))
	client.Client = &http.Client{Transport: customAlipaySecurityTransport(func(*http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("security test disables merchant network")
	})}
	return &Alipay{client: client, config: map[string]string{"appId": "test-app"}}, func(v url.Values) string {
		sign, err := client.SignValues(v, nsign.WithIgnore("sign", "sign_type", "alipay_cert_sn"))
		require.NoError(t, err)
		v.Set("sign", base64.StdEncoding.EncodeToString(sign))
		return v.Encode()
	}
}

type customAlipaySecurityTransport func(*http.Request) (*http.Response, error)

func (f customAlipaySecurityTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func customAlipayNotificationFields() url.Values {
	return url.Values{"app_id": {"test-app"}, "out_trade_no": {"sub2_security_order"}, "trade_no": {"test-trade"},
		"total_amount": {"10.00"}, "trade_status": {"TRADE_SUCCESS"}, "sign_type": {"RSA2"}}
}

func TestCustomAlipaySecuritySignedNotification(t *testing.T) {
	a, sign := customAlipaySignedNotification(t)
	valid := sign(customAlipayNotificationFields())
	n, err := a.VerifyNotification(context.Background(), valid, nil)
	require.NoError(t, err)
	require.Equal(t, "test-app", n.Metadata["app_id"])
	require.Equal(t, 10.0, n.Amount)
	for _, tc := range []struct {
		name string
		edit func(url.Values)
	}{
		{"missing-app", func(v url.Values) { v.Del("app_id") }},
		{"wrong-app", func(v url.Values) { v.Set("app_id", "other-app") }},
		{"missing-total", func(v url.Values) { v.Del("total_amount"); v.Set("receipt_amount", "10.00") }},
		{"invalid-total-fallback", func(v url.Values) { v.Set("total_amount", "invalid"); v.Set("receipt_amount", "10.00") }},
		{"nan", func(v url.Values) { v.Set("total_amount", "NaN") }},
		{"infinite", func(v url.Values) { v.Set("total_amount", "Inf") }},
		{"negative", func(v url.Values) { v.Set("total_amount", "-10.00") }},
		{"zero", func(v url.Values) { v.Set("total_amount", "0.00") }},
		{"sub-cent", func(v url.Values) { v.Set("total_amount", "10.001") }},
		{"exponent", func(v url.Values) { v.Set("total_amount", "1e1") }},
		{"missing-trade", func(v url.Values) { v.Del("trade_no") }},
		{"missing-order", func(v url.Values) { v.Del("out_trade_no") }},
		{"duplicate-app", func(v url.Values) { v.Add("app_id", "other-app") }},
		{"duplicate-amount", func(v url.Values) { v.Add("total_amount", "0.01") }},
		{"wrong-sign-type", func(v url.Values) { v.Set("sign_type", "RSA") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := customAlipayNotificationFields()
			tc.edit(v)
			n, err := a.VerifyNotification(context.Background(), sign(v), nil)
			require.Error(t, err)
			require.Nil(t, n)
		})
	}
	for _, change := range []func(url.Values){
		func(v url.Values) { v.Set("total_amount", "99.00") },
		func(v url.Values) { v.Set("out_trade_no", "other-order") },
		func(v url.Values) { v.Del("sign") },
	} {
		v, err := url.ParseQuery(valid)
		require.NoError(t, err)
		change(v)
		_, err = a.VerifyNotification(context.Background(), v.Encode(), nil)
		require.Error(t, err, "未重新签名的篡改必须拒绝")
	}
	// 未参与签名的证书序列号不能触发网络请求或替换公钥模式的信任锚。
	networkCalls := 0
	a.client.Client = &http.Client{Transport: customAlipaySecurityTransport(func(*http.Request) (*http.Response, error) {
		networkCalls++
		return nil, fmt.Errorf("unexpected certificate download")
	})}
	v := customAlipayNotificationFields()
	v.Set("alipay_cert_sn", "attacker-controlled-serial")
	_, err = a.VerifyNotification(context.Background(), sign(v), nil)
	require.NoError(t, err)
	require.Zero(t, networkCalls)
}

func TestCustomAlipaySecurityPaidQuery(t *testing.T) {
	a, _ := customAlipaySignedNotification(t)
	for _, tc := range []struct {
		name   string
		edit   func(map[string]any)
		accept bool
	}{
		{"valid", func(map[string]any) {}, true},
		{"other-order", func(v map[string]any) { v["out_trade_no"] = "other-order" }, false},
		{"missing-order", func(v map[string]any) { delete(v, "out_trade_no") }, false},
		{"missing-trade", func(v map[string]any) { delete(v, "trade_no") }, false},
		{"missing-total-with-receipt", func(v map[string]any) { delete(v, "total_amount"); v["receipt_amount"] = "10.00" }, false},
		{"sub-cent", func(v map[string]any) { v["total_amount"] = "10.001" }, false},
		{"nan", func(v map[string]any) { v["total_amount"] = "NaN" }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			biz := map[string]any{"code": "10000", "msg": "Success", "out_trade_no": "sub2_security_order", "trade_no": "test-trade",
				"trade_status": "TRADE_SUCCESS", "total_amount": "10.00"}
			tc.edit(biz)
			body, err := json.Marshal(biz)
			require.NoError(t, err)
			sign, err := a.client.SignBytes(body)
			require.NoError(t, err)
			response := fmt.Sprintf(`{"alipay_trade_query_response":%s,"sign":"%s"}`, body, base64.StdEncoding.EncodeToString(sign))
			a.client.Client = &http.Client{Transport: customAlipaySecurityTransport(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(response))}, nil
			})}
			r, err := a.QueryOrder(context.Background(), "sub2_security_order")
			if tc.accept {
				require.NoError(t, err)
				require.Equal(t, 10.0, r.Amount)
			} else {
				require.Error(t, err)
				require.Nil(t, r)
			}
		})
	}
}
