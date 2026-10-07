package service

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const customAlipayDesktopWapSnapshotKey = "custom_alipay_desktop_wap_qrcode"
const customAlipayDesktopWapPath = "/api/v1/payment/public/alipay/wap/"

func shouldUseAlipayDesktopWapQRCode(req CreateOrderRequest, cfg *PaymentConfig, sel *payment.InstanceSelection) bool {
	return cfg != nil && cfg.AlipayDesktopWapQRCode && !req.IsMobile && sel != nil &&
		strings.EqualFold(strings.TrimSpace(sel.ProviderKey), payment.TypeAlipay)
}

// 新扫码模式不信任客户端 Referer；配置了前端 URL 时只认管理员指定站点。
// 未配置时复用请求 Host，部署反代须拒绝非本站 Host，不从转发头扩展白名单。
func (s *PaymentService) canonicalAlipayDesktopWapReturnURL(ctx context.Context, req CreateOrderRequest) (string, error) {
	trustedHost := req.SrcHost
	if s.configService != nil && s.configService.settingRepo != nil {
		frontend, err := s.configService.settingRepo.GetValue(ctx, SettingKeyFrontendURL)
		if err != nil && !errors.Is(err, ErrSettingNotFound) {
			return "", infraerrors.ServiceUnavailable("PAYMENT_SITE_UNAVAILABLE", "无法校验本站付款地址")
		}
		if strings.TrimSpace(frontend) != "" {
			u, err := url.Parse(strings.TrimSpace(frontend))
			if err != nil || u.Host == "" || u.User != nil || !isSecureAlipayDesktopWapOrigin(u) {
				return "", infraerrors.ServiceUnavailable("PAYMENT_SITE_UNAVAILABLE", "请为本站配置有效的 HTTPS 前端地址")
			}
			trustedHost = u.Host
		}
	}
	canonical, err := CanonicalizeReturnURL(req.ReturnURL, trustedHost, "")
	if err != nil {
		return "", err
	}
	u, err := url.Parse(canonical)
	if err != nil || u.Host == "" || u.User != nil || !isSecureAlipayDesktopWapOrigin(u) {
		return "", infraerrors.BadRequest("INVALID_RETURN_URL", "电脑端支付宝扫码需要 HTTPS 本站回跳地址（本机回环允许 HTTP）")
	}
	return canonical, nil
}

// 新模式的签名和二维码随订单保存，扫码不创建新订单、不重新读取开关。
func prepareAlipayDesktopWapQRCode(order *dbent.PaymentOrder, canonicalReturnURL string, pr *payment.CreatePaymentResponse) error {
	if pr == nil || validateAlipayDesktopWapURL(order, pr.PayURL) != nil {
		return infraerrors.ServiceUnavailable("INVALID_ALIPAY_WAP_URL", "支付宝手机网站支付地址无效")
	}
	u, err := url.Parse(canonicalReturnURL)
	if err != nil || u == nil || u.Host == "" || u.User != nil || !isSecureAlipayDesktopWapOrigin(u) {
		return infraerrors.BadRequest("INVALID_RETURN_URL", "电脑端支付宝扫码需要 HTTPS 本站回跳地址（本机回环允许 HTTP）")
	}
	// 直接扫描官方签名WAP地址，不再经过本站外链确认和302入口。
	// 前端M级纠错的二维码字节容量为2331，保留余量，禁止截断付款签名。
	if len(pr.PayURL) > 2300 {
		return infraerrors.BadRequest("ALIPAY_WAP_QR_TOO_LONG", "支付宝付款地址过长，无法生成可扫描二维码，请缩短订单标题或回跳地址")
	}
	pr.QRCode = pr.PayURL
	return nil
}

// 真实付款令牌必须通过 HTTPS 传输；仅本机开发入口允许回环 HTTP。
func isSecureAlipayDesktopWapOrigin(u *url.URL) bool {
	if u.Scheme == "https" {
		return true
	}
	if u.Scheme != "http" {
		return false
	}
	host := u.Hostname()
	ip := net.ParseIP(host)
	return strings.EqualFold(host, "localhost") || (ip != nil && ip.IsLoopback())
}

// 只允许已经生成且与本订单商户、金额、订单号匹配的官方 WAP 收银台。
func validateAlipayDesktopWapURL(order *dbent.PaymentOrder, raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u == nil || u.Scheme != "https" || u.Host != "openapi.alipay.com" ||
		u.User != nil || u.Fragment != "" || u.Path != "/gateway.do" || order == nil {
		return fmt.Errorf("invalid alipay wap gateway")
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil || q.Get("method") != "alipay.trade.wap.pay" || q.Get("sign") == "" {
		return fmt.Errorf("invalid alipay wap method")
	}
	for _, vals := range q {
		if len(vals) != 1 {
			return fmt.Errorf("duplicate alipay wap parameter")
		}
	}
	snapshot := psOrderProviderSnapshot(order)
	if snapshot == nil || snapshot.ProviderKey != payment.TypeAlipay || snapshot.MerchantAppID == "" || q.Get("app_id") != snapshot.MerchantAppID {
		return fmt.Errorf("invalid alipay wap merchant")
	}
	var biz struct {
		OutTradeNo  string `json:"out_trade_no"`
		TotalAmount string `json:"total_amount"`
		ProductCode string `json:"product_code"`
		TimeExpire  string `json:"time_expire"`
	}
	if !isUnambiguousAlipayWapContent(q.Get("biz_content")) || json.Unmarshal([]byte(q.Get("biz_content")), &biz) != nil {
		return fmt.Errorf("invalid alipay wap content")
	}
	expires := order.ExpiresAt.In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04")
	if biz.OutTradeNo != order.OutTradeNo || biz.TotalAmount != payment.FormatAmountForCurrency(order.PayAmount, payment.DefaultPaymentCurrency) ||
		biz.ProductCode != "QUICK_WAP_WAY" || biz.TimeExpire != expires {
		return fmt.Errorf("alipay wap order mismatch")
	}
	return nil
}

// JSON map/struct 默认静默覆盖重复键，网关可能采用另一种解释，须先拒绝。
func isUnambiguousAlipayWapContent(raw string) bool {
	dec := json.NewDecoder(strings.NewReader(raw))
	start, err := dec.Token()
	if err != nil || start != json.Delim('{') {
		return false
	}
	seen := map[string]bool{}
	for dec.More() {
		token, err := dec.Token()
		key, ok := token.(string)
		key = strings.ToLower(key) // 与 Go struct 解码的大小写兼容保持同一解释。
		if err != nil || !ok || seen[key] {
			return false
		}
		seen[key] = true
		var value json.RawMessage
		if dec.Decode(&value) != nil {
			return false
		}
	}
	end, err := dec.Token()
	if err != nil || end != json.Delim('}') {
		return false
	}
	_, err = dec.Token()
	return err == io.EOF
}

func invalidAlipayDesktopWapToken() error {
	return infraerrors.NotFound("INVALID_ALIPAY_WAP_TOKEN", "付款入口无效，请返回电脑端重新发起支付")
}

// ResolveAlipayDesktopWapURL 只用持有二维码的随机令牌授权，不依赖手机 Cookie/JWT。
// 此入口只读取订单和返回已签名地址，不查询支付宝、不入账，也不暴露订单资料。
func (s *PaymentService) ResolveAlipayDesktopWapURL(ctx context.Context, orderID int64, token string) (string, error) {
	if len(token) != 43 || orderID <= 0 {
		return "", invalidAlipayDesktopWapToken()
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(token)
	if err != nil || len(decoded) != 32 {
		return "", invalidAlipayDesktopWapToken()
	}
	order, err := s.entClient.PaymentOrder.Get(ctx, orderID)
	if err != nil {
		if dbent.IsNotFound(err) {
			return "", invalidAlipayDesktopWapToken()
		}
		return "", fmt.Errorf("load alipay wap order: %w", err)
	}
	qr, err := url.Parse(psStringValue(order.QrCode))
	if err != nil || qr == nil || qr.Path != customAlipayDesktopWapPath+strconv.FormatInt(orderID, 10) ||
		order.ProviderSnapshot[customAlipayDesktopWapSnapshotKey] != true || psStringValue(order.ProviderKey) != payment.TypeAlipay ||
		subtle.ConstantTimeCompare([]byte(qr.Query().Get("token")), []byte(token)) != 1 {
		return "", invalidAlipayDesktopWapToken()
	}
	if order.Status != OrderStatusPending {
		return "", infraerrors.New(http.StatusGone, "ALIPAY_WAP_ORDER_UNAVAILABLE", "订单已支付、取消或失效，请返回电脑端查看订单")
	}
	if !order.ExpiresAt.After(time.Now()) {
		return "", infraerrors.New(http.StatusGone, "ALIPAY_WAP_ORDER_EXPIRED", "二维码已过期，请返回电脑端重新发起支付")
	}
	if err := validateAlipayDesktopWapURL(order, psStringValue(order.PayURL)); err != nil {
		return "", infraerrors.ServiceUnavailable("INVALID_ALIPAY_WAP_URL", "付款入口暂不可用，请返回电脑端查看订单")
	}
	return psStringValue(order.PayURL), nil
}
