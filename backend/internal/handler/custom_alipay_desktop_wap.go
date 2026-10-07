package handler

import (
	"net/http"
	"net/url"
	"strconv"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/gin-gonic/gin"
)

// AlipayDesktopWapGuard 在限流中间件前执行，错误和 429 响应同样禁止缓存。
func (h *PaymentHandler) AlipayDesktopWapGuard(c *gin.Context) {
	if _, _, ok := parseAlipayDesktopWapRequest(c); !ok {
		c.Abort()
		return
	}
	c.Next()
}

func parseAlipayDesktopWapRequest(c *gin.Context) (int64, string, bool) {
	c.Header("Cache-Control", "no-store, private")
	c.Header("Pragma", "no-cache")
	c.Header("Referrer-Policy", "no-referrer")
	c.Header("X-Content-Type-Options", "nosniff")
	if len(c.Request.URL.RawQuery) > 512 {
		c.String(http.StatusNotFound, "付款入口无效，请返回电脑端重新发起支付")
		return 0, "", false
	}
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	query, queryErr := url.ParseQuery(c.Request.URL.RawQuery)
	if err != nil || orderID <= 0 || strconv.FormatInt(orderID, 10) != c.Param("id") ||
		queryErr != nil || !validAlipayWapQuery(query) {
		c.String(http.StatusNotFound, "付款入口无效，请返回电脑端重新发起支付")
		return 0, "", false
	}
	return orderID, query.Get("token"), true
}

// 支付宝外链确认页会追加这些流程参数；只忽略已知、单值且有界的参数。
// 它们不参与授权，也绝不拼入最终付款地址；订单仍仅以原随机 token 校验。
func validAlipayWapQuery(query url.Values) bool {
	if len(query["token"]) != 1 || len(query.Get("token")) != 43 {
		return false
	}
	for key, values := range query {
		limit := 0
		switch key {
		case "token":
			continue
		case "flowT", "flow":
			limit = 32
		case "flowSign":
			limit = 128
		default:
			return false
		}
		if len(values) != 1 || len(values[0]) > limit {
			return false
		}
	}
	return true
}

// AlipayDesktopWapRedirect 无需手机登录，订单校验后直接进入支付宝收银台。
func (h *PaymentHandler) AlipayDesktopWapRedirect(c *gin.Context) {
	orderID, token, ok := parseAlipayDesktopWapRequest(c)
	if !ok {
		return
	}
	payURL, err := h.paymentService.ResolveAlipayDesktopWapURL(c.Request.Context(), orderID, token)
	if err != nil {
		message := "付款入口暂不可用，请稍后重试"
		if infraerrors.Code(err) < http.StatusInternalServerError {
			message = infraerrors.Message(err)
		}
		c.String(infraerrors.Code(err), "%s", message)
		return
	}
	// 付款长地址只出现在跳转所需的 Location，不再复制进 HTML 响应正文。
	c.Header("Location", payURL)
	c.Status(http.StatusFound)
}
