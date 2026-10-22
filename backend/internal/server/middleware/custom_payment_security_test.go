package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func customQueryTestRouter(t *testing.T, redisClient *redis.Client) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	require.NoError(t, router.SetTrustedProxies(nil))
	router.Use(CustomPaymentPublicQueryLimiter(redisClient))
	router.POST("/query", func(request *gin.Context) { request.Status(http.StatusNoContent) })
	return router
}

func customQueryRequest(router *gin.Engine, forwarded string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/query", nil)
	request.RemoteAddr = "198.51.100.20:2000"
	request.Header.Set("X-Forwarded-For", forwarded)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestCustomPaymentQuerySharedRedisBudget(t *testing.T) {
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })
	routers := []*gin.Engine{customQueryTestRouter(t, redisClient), customQueryTestRouter(t, redisClient)}
	for requestIndex := 0; requestIndex < customPaymentQueryIPLimit; requestIndex++ {
		require.Equal(t, http.StatusNoContent, customQueryRequest(routers[requestIndex%2], "").Code)
	}
	rejected := customQueryRequest(routers[0], "")
	require.Equal(t, http.StatusTooManyRequests, rejected.Code)
	require.NotEmpty(t, rejected.Header().Get("Retry-After"))
	require.Equal(t, "no-store", rejected.Header().Get("Cache-Control"))
	redisServer.FastForward(time.Minute + time.Second)
	require.Equal(t, http.StatusNoContent, customQueryRequest(customQueryTestRouter(t, redisClient), "").Code)
}

func TestCustomPaymentQueryFallbackRejectsForgedForwardedIPs(t *testing.T) {
	router := customQueryTestRouter(t, nil)
	for requestIndex := 0; requestIndex < customPaymentQueryIPLimit; requestIndex++ {
		require.Equal(t, http.StatusNoContent, customQueryRequest(router, "203.0.113."+strconv.Itoa(requestIndex+1)).Code)
	}
	require.Equal(t, http.StatusTooManyRequests, customQueryRequest(router, "192.0.2.99").Code)
}

func TestCustomPaymentQueryGlobalBudgetAndReset(t *testing.T) {
	budget := &customPaymentQueryBudget{}
	now := time.Now()
	for clientIndex := 0; clientIndex < customPaymentQueryGlobalLimit; clientIndex++ {
		allowed, _ := budget.allow(strconv.Itoa(clientIndex), now)
		require.True(t, allowed)
	}
	allowed, retryAfter := budget.allow("another-client", now)
	require.False(t, allowed)
	require.Positive(t, retryAfter)
	require.Len(t, budget.clients, customPaymentQueryGlobalLimit)
	allowed, _ = budget.allow("another-client", now.Add(time.Minute))
	require.True(t, allowed)
	require.Len(t, budget.clients, 1)
}

func TestCustomPaymentAdminStepUp(t *testing.T) {
	for _, scenario := range []struct {
		name    string
		method  string
		enabled bool
		bound   bool
		grant   bool
		apiKey  bool
		status  int
	}{
		{"read_without_grant", http.MethodGet, true, true, false, false, http.StatusNoContent},
		{"write_without_grant", http.MethodPost, true, true, false, false, http.StatusForbidden},
		{"write_with_grant", http.MethodPut, true, true, true, false, http.StatusNoContent},
		{"unbound_admin", http.MethodDelete, true, false, true, false, http.StatusForbidden},
		{"machine_credential", http.MethodPost, true, true, true, true, http.StatusForbidden},
		{"feature_off", http.MethodPost, false, false, false, false, http.StatusNoContent},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(request *gin.Context) {
				request.Set(string(ContextKeyUser), AuthSubject{UserID: 1})
				if scenario.apiKey {
					request.Set("auth_method", service.AuditAuthMethodAdminAPIKey)
				}
			})
			stepUp := stepUpAuth(stubStepUpGrantChecker{granted: scenario.grant}, stubStepUpUserReader{user: &service.User{ID: 1, TotpEnabled: scenario.bound}}, stubStepUpSettingReader{enabled: scenario.enabled})
			router.Use(CustomPaymentAdminStepUp(StepUpAuthMiddleware(stepUp)))
			router.Handle(scenario.method, "/sensitive", func(request *gin.Context) { request.Status(http.StatusNoContent) })
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(scenario.method, "/sensitive", nil))
			require.Equal(t, scenario.status, response.Code)
		})
	}
}

func TestCustomPaymentAdminStepUpMissingMiddleware(t *testing.T) {
	router := gin.New()
	router.Use(CustomPaymentAdminStepUp(nil))
	router.POST("/sensitive", func(request *gin.Context) { t.Fatal("write reached handler") })
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/sensitive", nil))
	require.Equal(t, http.StatusServiceUnavailable, response.Code)
}

func TestCustomPaymentQueryGlobalRedisBudget(t *testing.T) {
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })
	require.NoError(t, redisClient.Set(context.Background(), "rate_limit:custom-payment-public-query:global", customPaymentQueryGlobalLimit, time.Minute).Err())
	require.Equal(t, http.StatusTooManyRequests, customQueryRequest(customQueryTestRouter(t, redisClient), "").Code)
}
