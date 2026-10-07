//go:build unit

package routes

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestCustomAlipayDesktopWapPublicRouteNeedsNoLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := sql.Open("sqlite", "file:custom_wap_route?mode=memory&cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(entsql.OpenDB(dialect.SQLite, db))))
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	user, err := client.User.Create().SetEmail("route-wap@example.test").SetUsername("wap").SetPasswordHash("hash").Save(ctx)
	require.NoError(t, err)
	exp := time.Now().Add(30 * time.Minute).Truncate(time.Minute)
	order, err := client.PaymentOrder.Create().SetUserID(user.ID).SetUserEmail(user.Email).SetUserName(user.Username).
		SetAmount(10).SetPayAmount(10).SetFeeRate(0).SetRechargeCode("CUSTOM-WAP-ROUTE").SetOutTradeNo("custom-route-wap").
		SetPaymentType(payment.TypeAlipay).SetPaymentTradeNo("").SetClientIP("127.0.0.1").SetSrcHost("merchant.example").
		SetOrderType(payment.OrderTypeBalance).SetStatus(service.OrderStatusPending).
		SetExpiresAt(exp).SetProviderKey(payment.TypeAlipay).SetProviderSnapshot(map[string]any{"provider_key": payment.TypeAlipay,
		"merchant_app_id": "test-app", "custom_alipay_desktop_wap_qrcode": true}).Save(ctx)
	require.NoError(t, err)
	biz, err := json.Marshal(map[string]string{"out_trade_no": order.OutTradeNo, "total_amount": "10.00", "product_code": "QUICK_WAP_WAY",
		"time_expire": exp.In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04")})
	require.NoError(t, err)
	payURL := "https://openapi.alipay.com/gateway.do?" + url.Values{"app_id": {"test-app"}, "method": {"alipay.trade.wap.pay"}, "sign": {"fixture"}, "biz_content": {string(biz)}}.Encode()
	token := base64.RawURLEncoding.EncodeToString(make([]byte, 32))
	path := "/api/v1/payment/public/alipay/wap/" + strconv.FormatInt(order.ID, 10)
	_, err = client.PaymentOrder.UpdateOneID(order.ID).SetPayURL(payURL).SetQrCode("https://merchant.example" + path + "?token=" + token).Save(ctx)
	require.NoError(t, err)
	svc := service.NewPaymentService(client, payment.NewRegistry(), nil, nil, nil, nil, nil, nil, nil)
	router := gin.New()
	rejectAuth := func(c *gin.Context) { c.AbortWithStatus(http.StatusUnauthorized) }
	RegisterPaymentRoutes(router.Group("/api/v1"), handler.NewPaymentHandler(svc, nil), &handler.PaymentWebhookHandler{}, &admin.PaymentHandler{},
		middleware.JWTAuthMiddleware(rejectAuth), middleware.AdminAuthMiddleware(rejectAuth), middleware.AuditLogMiddleware(func(c *gin.Context) { c.Next() }), nil, nil, nil)
	get := func(target string) *httptest.ResponseRecorder {
		r := httptest.NewRecorder()
		router.ServeHTTP(r, httptest.NewRequest(http.MethodGet, target, nil))
		return r
	}
	r := get(path + "?token=" + token)
	require.Equal(t, http.StatusFound, r.Code)
	require.Equal(t, payURL, r.Header().Get("Location"))
	require.Contains(t, r.Header().Get("Cache-Control"), "no-store")
	require.Equal(t, "no-referrer", r.Header().Get("Referrer-Policy"))
	require.NotContains(t, r.Header().Get("Location"), "/login")
	require.Empty(t, r.Body.String())
	// 真机支付宝的外链确认页会追加流程参数，不能将合法订单误判为404。
	r = get(path + "?token=" + token + "&flowT=1791344936544&flowSign=9dfb5850c7&flow=flow")
	require.Equal(t, http.StatusFound, r.Code)
	require.Equal(t, payURL, r.Header().Get("Location"))
	require.NotContains(t, r.Header().Get("Location"), "flowSign")
	require.Equal(t, http.StatusUnauthorized, get("/api/v1/payment/orders/"+strconv.FormatInt(order.ID, 10)).Code)
	r = get(path + "?token=invalid")
	require.Equal(t, http.StatusNotFound, r.Code)
	require.Empty(t, r.Header().Get("Location"))
	require.NotContains(t, r.Body.String(), payURL)
	for _, target := range []string{
		path + "?token=" + token + "&token=other",
		path + "?token=" + token + "&other=value",
		path + "?token=" + token + "&flow=flow&flow=other",
		path + "?token=" + token + "&flowSign=" + strings.Repeat("A", 129),
		path + "?flow=flow&flowT=1791344936544&flowSign=test",
		path + "?token=" + token + "&bad=%ZZ",
		path + "?token=" + strings.Repeat("A", 10000),
		path + "?token=" + token + "%0A",
		strings.Replace(path, "/"+strconv.FormatInt(order.ID, 10), "/0"+strconv.FormatInt(order.ID, 10), 1) + "?token=" + token,
	} {
		r = get(target)
		require.Equal(t, http.StatusNotFound, r.Code)
		require.Empty(t, r.Header().Get("Location"))
		require.Contains(t, r.Header().Get("Cache-Control"), "no-store")
	}
	_, err = client.PaymentOrder.UpdateOneID(order.ID).SetStatus(service.OrderStatusCancelled).Save(ctx)
	require.NoError(t, err)
	r = get(path + "?token=" + token)
	require.Equal(t, http.StatusGone, r.Code)
	require.Contains(t, r.Body.String(), "取消")
	require.Empty(t, r.Header().Get("Location"))
}

func TestCustomAlipayDesktopWapGuardProtectsRateLimitedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &handler.PaymentHandler{}
	r := gin.New()
	r.GET("/wap/:id", h.AlipayDesktopWapGuard, func(c *gin.Context) { c.AbortWithStatus(http.StatusTooManyRequests) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/wap/1?token="+base64.RawURLEncoding.EncodeToString(make([]byte, 32)), nil))
	require.Equal(t, http.StatusTooManyRequests, w.Code)
	require.Contains(t, w.Header().Get("Cache-Control"), "no-store")
	require.Equal(t, "no-referrer", w.Header().Get("Referrer-Policy"))
	require.Empty(t, w.Header().Get("Location"))
}
