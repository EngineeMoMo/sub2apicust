package provider

import (
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var customAlipayAmountPattern = regexp.MustCompile(`^[0-9]+(?:\.[0-9]{1,2})?$`)

// 支付金额只能来自签名中的订单总额，不能用实收、优惠或买家付款额补齐。
func customAlipayPaidAmount(raw string) (float64, error) {
	if len(raw) > 32 || !customAlipayAmountPattern.MatchString(raw) {
		return 0, fmt.Errorf("invalid alipay total_amount")
	}
	amount, err := strconv.ParseFloat(raw, 64)
	if err != nil || amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0, fmt.Errorf("invalid alipay total_amount")
	}
	return amount, nil
}

// 先拒绝歧义与缺项，再交给 SDK 使用配置的支付宝公钥验签。
func customAlipayNotificationValues(rawBody, expectedAppID string) (url.Values, error) {
	values, err := url.ParseQuery(rawBody)
	if err != nil {
		return nil, fmt.Errorf("invalid alipay notification encoding")
	}
	for _, vals := range values {
		if len(vals) != 1 {
			return nil, fmt.Errorf("duplicate alipay notification parameter")
		}
	}
	for _, key := range []string{"app_id", "out_trade_no", "trade_no", "trade_status", "total_amount", "sign", "sign_type"} {
		if value := values.Get(key); value == "" || strings.TrimSpace(value) != value {
			return nil, fmt.Errorf("invalid alipay notification field: %s", key)
		}
	}
	if expectedAppID == "" || values.Get("app_id") != expectedAppID || values.Get("sign_type") != "RSA2" {
		return nil, fmt.Errorf("alipay notification identity or sign type mismatch")
	}
	if _, err := customAlipayPaidAmount(values.Get("total_amount")); err != nil {
		return nil, err
	}
	// 当前接入仅支持公钥模式。此字段不参与 SDK 验签，不能让未验签请求
	// 通过任意证书序列号触发 SDK 的远程证书下载或切换已配置的信任公钥。
	values.Del("alipay_cert_sn")
	return values, nil
}
