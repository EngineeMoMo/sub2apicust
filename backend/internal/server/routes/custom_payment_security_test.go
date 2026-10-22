package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCustomPaymentPublicRoutesShareQueryBudget(t *testing.T) {
	router := newPaymentRoutesTestRouter(nil)
	for requestIndex := 0; requestIndex <= 20; requestIndex++ {
		endpoint := "verify"
		if requestIndex%2 == 0 {
			endpoint = "resolve"
		}
		request := httptest.NewRequest(http.MethodPost, "/api/v1/payment/public/orders/"+endpoint, strings.NewReader(`{}`))
		request.Header.Set("Content-Type", "application/json")
		request.RemoteAddr = "198.51.100.20:1000"
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		expected := http.StatusBadRequest
		if requestIndex == 20 {
			expected = http.StatusTooManyRequests
		}
		require.Equal(t, expected, response.Code, "request %d", requestIndex+1)
	}
}

func TestCustomAllAdminPaymentWritesUseStepUp(t *testing.T) {
	router := gin.New()
	noop := func(request *gin.Context) { request.Next() }
	reject := func(request *gin.Context) {
		request.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "STEP_UP_REQUIRED"})
	}
	RegisterPaymentRoutes(router.Group("/api/v1"), &handler.PaymentHandler{}, &handler.PaymentWebhookHandler{}, &admin.PaymentHandler{}, middleware.JWTAuthMiddleware(noop), middleware.AdminAuthMiddleware(noop), middleware.AuditLogMiddleware(noop), middleware.StepUpAuthMiddleware(reject), nil, nil, nil)
	checked := 0
	for _, route := range router.Routes() {
		if !strings.HasPrefix(route.Path, "/api/v1/admin/payment/") || route.Method == http.MethodGet {
			continue
		}
		t.Run(route.Method+route.Path, func(t *testing.T) {
			path := strings.ReplaceAll(route.Path, ":id", "1")
			request := httptest.NewRequest(route.Method, path, strings.NewReader(`{}`))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.Equal(t, http.StatusForbidden, response.Code)
			require.Contains(t, response.Body.String(), "STEP_UP_REQUIRED")
		})
		checked++
	}
	require.GreaterOrEqual(t, checked, 15)
}
