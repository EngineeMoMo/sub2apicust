package middleware

import (
	"context"
	"math"
	"net/http"
	"net/netip"
	"strconv"
	"sync"
	"time"

	ratelimit "github.com/Wei-Shaw/sub2api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const customPaymentQueryIPLimit = 20
const customPaymentQueryGlobalLimit = 600

type customPaymentQueryWindow struct {
	count int
	until time.Time
}

type customPaymentQueryBudget struct {
	mutex   sync.Mutex
	clients map[string]customPaymentQueryWindow
	global  customPaymentQueryWindow
}

func (budget *customPaymentQueryBudget) allow(clientIP string, now time.Time) (bool, time.Duration) {
	budget.mutex.Lock()
	defer budget.mutex.Unlock()
	if !now.Before(budget.global.until) {
		budget.global = customPaymentQueryWindow{until: now.Add(time.Minute)}
		budget.clients = make(map[string]customPaymentQueryWindow)
	}
	client := budget.clients[clientIP]
	if client.count >= customPaymentQueryIPLimit || budget.global.count >= customPaymentQueryGlobalLimit {
		return false, budget.global.until.Sub(now)
	}
	client.count++
	budget.clients[clientIP] = client
	budget.global.count++
	return true, 0
}

func CustomPaymentPublicQueryLimiter(redisClient *redis.Client) gin.HandlerFunc {
	budget := &customPaymentQueryBudget{}
	limiter := ratelimit.NewRateLimiter(redisClient)
	return func(request *gin.Context) {
		request.Header("Cache-Control", "no-store")
		clientIP, err := netip.ParseAddr(request.ClientIP())
		if err != nil || clientIP.IsUnspecified() {
			AbortWithError(request, http.StatusServiceUnavailable, "CLIENT_IP_UNAVAILABLE", "Unable to verify client address")
			return
		}
		address := clientIP.Unmap().String()
		if allowed, retryAfter := budget.allow(address, time.Now()); !allowed {
			customRejectPaymentQuery(request, retryAfter)
			return
		}
		if redisClient != nil {
			limitContext, cancel := context.WithTimeout(request.Request.Context(), 200*time.Millisecond)
			defer cancel()
			for _, scope := range []struct {
				key   string
				limit int
			}{
				{"ip:" + address, customPaymentQueryIPLimit},
				{"global", customPaymentQueryGlobalLimit},
			} {
				result, limitError := limiter.Allow(limitContext, "custom-payment-public-query:"+scope.key, scope.limit, time.Minute)
				if limitError != nil {
					break
				}
				if !result.Allowed {
					customRejectPaymentQuery(request, result.RetryAfter)
					return
				}
			}
		}
		request.Request.Body = http.MaxBytesReader(request.Writer, request.Request.Body, 64<<10)
		request.Next()
	}
}

func customRejectPaymentQuery(request *gin.Context, retryAfter time.Duration) {
	request.Header("Retry-After", strconv.Itoa(max(1, int(math.Ceil(retryAfter.Seconds())))))
	AbortWithError(request, http.StatusTooManyRequests, "RATE_LIMITED", "rate limit exceeded")
}

func CustomPaymentAdminStepUp(stepUp StepUpAuthMiddleware) gin.HandlerFunc {
	return func(request *gin.Context) {
		if request.Request.Method == http.MethodGet || request.Request.Method == http.MethodHead || request.Request.Method == http.MethodOptions {
			request.Next()
			return
		}
		if stepUp == nil {
			AbortWithError(request, http.StatusServiceUnavailable, "STEP_UP_UNAVAILABLE", "Sensitive operation verification unavailable")
			return
		}
		gin.HandlerFunc(stepUp)(request)
	}
}
