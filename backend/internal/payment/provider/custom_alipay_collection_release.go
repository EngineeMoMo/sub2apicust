package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/smartwalle/alipay/v3"
)

// CustomCollectionUnpaidFinal 在释放收款预占前向支付宝复查一次。只读查询，不关单、不入账。
// [CUSTOM] 只有“交易不存在”，或交易已关闭且从未打款给卖家，才算这笔订单不会再收款；
// 等待付款、已付款、查询失败、验签失败或订单号不符都返回 false，额度继续保留。
func (a *Alipay) CustomCollectionUnpaidFinal(ctx context.Context, outTradeNo string) (bool, error) {
	outTradeNo = strings.TrimSpace(outTradeNo)
	if outTradeNo == "" {
		return false, fmt.Errorf("alipay collection recheck: missing out_trade_no")
	}
	client, err := a.getClient()
	if err != nil {
		return false, err
	}
	result, err := client.TradeQuery(ctx, alipay.TradeQuery{OutTradeNo: outTradeNo})
	if err != nil {
		// 无签名的业务错误由 SDK 以 *alipay.Error 返回；只认交易不存在，网络、验签等错误都不确定。
		var gatewayErr *alipay.Error
		if errors.As(err, &gatewayErr) && gatewayErr != nil && gatewayErr.SubCode == alipayErrTradeNotExist {
			return true, nil
		}
		return false, fmt.Errorf("alipay collection recheck: %w", err)
	}
	if result == nil {
		return false, fmt.Errorf("alipay collection recheck: empty response")
	}
	if !result.IsSuccess() {
		// 已验签的业务失败同样只接受交易不存在。
		if result.SubCode == alipayErrTradeNotExist {
			return true, nil
		}
		return false, fmt.Errorf("alipay collection recheck: %s", result.Error.Error())
	}
	if result.OutTradeNo != outTradeNo {
		return false, fmt.Errorf("alipay collection recheck: order mismatch")
	}
	return result.TradeStatus == alipay.TradeStatusClosed && strings.TrimSpace(result.SendPayDate) == "", nil
}
