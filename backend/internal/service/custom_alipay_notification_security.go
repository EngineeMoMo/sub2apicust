package service

import "github.com/shopspring/decimal"

// 支付宝 total_amount 必须等于发送给网关的两位金额；不容忍一分钱差额。
func customAlipayAmountsMatch(paid, expected float64) bool {
	if !isValidProviderAmount(paid) || !isValidProviderAmount(expected) {
		return false
	}
	return decimal.NewFromFloat(paid).Equal(decimal.NewFromFloat(expected).Round(2))
}
