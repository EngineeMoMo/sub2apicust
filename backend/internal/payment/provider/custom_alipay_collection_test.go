//go:build unit

package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// 使用本地签名响应验证真实 SDK 的关单结果，不连接商户服务。
func TestCustomAlipayCollectionCloseConfirmation(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		wantErr    bool
	}{
		{"not-created", `{"code":"40004","msg":"Business Failed","sub_code":"ACQ.TRADE_NOT_EXIST","sub_msg":"交易不存在"}`, true},
		{"closed", `{"code":"10000","msg":"Success","out_trade_no":"collection-order"}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, _ := customAlipaySignedNotification(t)
			sign, err := a.client.SignBytes([]byte(tc.body))
			require.NoError(t, err)
			response := fmt.Sprintf(`{"alipay_trade_close_response":%s,"sign":"%s"}`, tc.body, base64.StdEncoding.EncodeToString(sign))
			a.client.Client = &http.Client{Transport: customAlipaySecurityTransport(func(r *http.Request) (*http.Response, error) {
				require.NoError(t, r.ParseForm())
				require.Equal(t, "alipay.trade.close", r.Form.Get("method"))
				return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(response))}, nil
			})}
			err = a.CancelPayment(context.Background(), "collection-order")
			if tc.wantErr {
				require.ErrorContains(t, err, "交易不存在")
			} else {
				require.NoError(t, err)
			}
		})
	}
}
