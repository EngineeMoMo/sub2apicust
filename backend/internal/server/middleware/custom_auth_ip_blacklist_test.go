//go:build unit

package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAuthIPBlacklistRouteCoverage(t *testing.T) {
	for _, path := range []string{"login", "register", "login/2fa", "passkey/login/begin", "passkey/login/finish", "send-verify-code", "forgot-password", "reset-password", "oauth/pending/bind-login", "oauth/google/start", "oauth/wechat/create-account"} {
		protected, count := AuthIPBlacklistRoute(http.MethodPost, "/api/v1/auth/"+path)
		require.True(t, protected, path)
		require.True(t, count, path)
	}
	for _, path := range []string{"refresh", "logout", "validate-promo-code", "validate-invitation-code", "oauth/wechat/payment/start", "oauth/wechat/payment/callback"} {
		protected, _ := AuthIPBlacklistRoute(http.MethodPost, "/api/v1/auth/"+path)
		require.False(t, protected, path)
	}
	for _, path := range []string{"oauth/google/start", "oauth/linuxdo/callback", "oauth/wechat/bind/start"} {
		protected, count := AuthIPBlacklistRoute(http.MethodGet, "/api/v1/auth/"+path)
		require.True(t, protected, path)
		require.False(t, count, path)
	}
	for _, path := range []string{"/v1/chat/completions", "/api/v1/admin/ip-blacklist/config", "/api/v1/auth/oauth/wechat/payment/callback"} {
		protected, _ := AuthIPBlacklistRoute(http.MethodGet, path)
		require.False(t, protected, path)
	}
}

type authIPMiddlewareRepo struct {
	service.AuthIPBlacklistRepository
	blocked bool
	failure error
}

func (repo *authIPMiddlewareRepo) Snapshot(context.Context, string) (*service.AuthIPBlacklistSnapshot, error) {
	snapshot := &service.AuthIPBlacklistSnapshot{Config: service.AuthIPBlacklistConfig{Enabled: true, Threshold: 10, WindowSeconds: 7200, Version: 1, BlockMode: service.AuthIPBlockModeTimed}}
	if repo.blocked {
		snapshot.Entry = &service.AuthIPBlacklistEntry{IP: "192.0.2.1", Blocked: true}
	}
	return snapshot, repo.failure
}

func (repo *authIPMiddlewareRepo) Block(_ context.Context, ip, path string, _ service.AuthIPBlacklistSnapshot, events service.AuthIPWindowCount) (*service.AuthIPBlacklistDecision, error) {
	repo.blocked = true
	return &service.AuthIPBlacklistDecision{NewlyBlocked: true, Entry: &service.AuthIPBlacklistEntry{
		ID: 1, IP: ip, Blocked: true, Generation: 1, TriggerPath: path, TriggerCount: events.Count, WindowSeconds: 7200,
	}}, nil
}

type authIPMiddlewareCounter struct{ calls int }

func (counter *authIPMiddlewareCounter) Time(context.Context) (time.Time, error) {
	return time.Now().UTC(), nil
}

func (counter *authIPMiddlewareCounter) CountAt(ctx context.Context, ip string, generation int64, config service.AuthIPBlacklistConfig, _ time.Time) (*service.AuthIPWindowCount, error) {
	return counter.Count(ctx, ip, generation, config)
}

func (counter *authIPMiddlewareCounter) Count(context.Context, string, int64, service.AuthIPBlacklistConfig) (*service.AuthIPWindowCount, error) {
	counter.calls++
	return &service.AuthIPWindowCount{Count: counter.calls}, nil
}

func TestAuthIPBlacklistMiddlewareResponseAndSuccessfulRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &authIPMiddlewareRepo{}
	counter := &authIPMiddlewareCounter{}
	router := gin.New()
	require.NoError(t, router.SetTrustedProxies(nil))
	router.Use(NewAuthIPBlacklistGuard(service.NewAuthIPBlacklistService(repo, counter)))
	for _, path := range []string{"/api/v1/auth/login", "/api/v1/auth/register", "/v1/chat/completions"} {
		router.POST(path, func(ctx *gin.Context) { ctx.Status(http.StatusOK) })
	}
	router.GET("/api/v1/auth/oauth/google/start", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })
	request := func(method, path string) *httptest.ResponseRecorder {
		result := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, nil)
		req.RemoteAddr = "192.0.2.1:1234"
		router.ServeHTTP(result, req)
		return result
	}
	require.Equal(t, 200, request(http.MethodPost, "/api/v1/auth/login").Code)
	require.Equal(t, 200, request(http.MethodPost, "/api/v1/auth/register").Code)
	for range 30 {
		require.Equal(t, 200, request(http.MethodPost, "/api/v1/auth/login").Code)
	}
	require.Zero(t, counter.calls, "连续成功登录不计数，不触发黑名单")
	require.Equal(t, 200, request(http.MethodGet, "/api/v1/auth/oauth/google/start").Code)
	require.Zero(t, counter.calls)
	repo.blocked = true
	denied := request(http.MethodPost, "/api/v1/auth/login")
	require.Equal(t, 403, denied.Code)
	require.Contains(t, denied.Body.String(), "IP_BLOCKED")
	require.Equal(t, 403, request(http.MethodGet, "/api/v1/auth/oauth/google/start").Code)
	require.Equal(t, 200, request(http.MethodPost, "/v1/chat/completions").Code)
	repo.blocked = false
	repo.failure = errors.New("offline")
	require.Equal(t, 503, request(http.MethodPost, "/api/v1/auth/login").Code)
	repo.failure = nil
	offline := gin.New()
	offline.POST("/api/v1/auth/login", NewAuthIPBlacklistGuard(service.NewAuthIPBlacklistService(repo, nil)))
	denied = httptest.NewRecorder()
	offline.ServeHTTP(denied, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil))
	require.Equal(t, 429, denied.Code)
	require.Equal(t, "60", denied.Header().Get("Retry-After"))
}

func TestAuthIPBlacklistAutomaticBlockIsAudited(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auditRepo := &auditCaptureRepository{}
	auditService := service.NewAuditLogService(auditRepo, nil)
	auditService.Start()
	router := gin.New()
	router.Use(gin.HandlerFunc(NewAuditLogMiddleware(auditService)))
	router.Use(NewAuthIPBlacklistGuard(service.NewAuthIPBlacklistService(&authIPMiddlewareRepo{}, &authIPMiddlewareCounter{calls: 9})))
	router.POST("/api/v1/auth/login", func(ctx *gin.Context) { response.BadRequest(ctx, "Invalid request") })
	for range 2 {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		req.RemoteAddr = "192.0.2.1:1234"
		router.ServeHTTP(recorder, req)
		require.Equal(t, 403, recorder.Code)
	}
	auditService.Stop()
	auditRepo.mu.Lock()
	defer auditRepo.mu.Unlock()
	require.Len(t, auditRepo.logs, 2)
	require.Equal(t, "auth.ip_blacklist.block", auditRepo.logs[0].Action)
	require.EqualValues(t, 10, auditRepo.logs[0].Extra["trigger_count"])
	require.Equal(t, "192.0.2.1", auditRepo.logs[0].Extra["target_ip"])
	require.Equal(t, "auth.ip_blacklist.reject", auditRepo.logs[1].Action)
}

func TestAuthIPBlacklistOnlyCountsAuthenticationFailures(t *testing.T) {
	for _, test := range []struct {
		status          int
		reason, message string
		counted         bool
	}{
		{200, "", "success", false},
		{401, "INVALID_CREDENTIALS", "invalid email or password", true},
		{401, "PASSKEY_VERIFICATION_FAILED", "passkey verification failed", true},
		{400, "", "Invalid request: missing password", true},
		{400, "", "Invalid request: json: cannot unmarshal number into Go struct field LoginRequest.turnstile_token of type string", true},
		{400, "", "Invalid request: json: cannot unmarshal number into Go struct field LoginRequest.tencent_captcha_ticket of type string", true},
		{400, "", "Invalid request: Key: 'Login2FARequest.TotpCode' Error:Field validation failed", true},
		{400, "INVALID_EMAIL", "invalid email", false},
		{400, "", "Email already exists", false},
		{400, "CAPTCHA_INVALID", "invalid captcha", false},
		{401, "INVALID_TOKEN", "invalid token", false},
		{401, "TOTP_INVALID", "invalid TOTP code", false},
		{400, "VERIFY_CODE_INVALID", "invalid verification code", false},
		{429, "", "rate limited", false},
		{500, "", "mail service failed", false},
		{503, "", "unavailable", false},
	} {
		body, err := json.Marshal(response.Response{Code: test.status, Reason: test.reason, Message: test.message})
		require.NoError(t, err)
		require.Equal(t, test.counted, authIPCountFailure(test.status, body), test.message)
	}
}

func TestAuthIPBlacklistFailureResponsesAndLargeSuccessRemainIntact(t *testing.T) {
	repo := &authIPMiddlewareRepo{}
	counter := &authIPMiddlewareCounter{}
	router := gin.New()
	router.Use(NewAuthIPBlacklistGuard(service.NewAuthIPBlacklistService(repo, counter)))
	large := strings.Repeat("ok", 200000)
	router.POST("/api/v1/auth/login", func(ctx *gin.Context) {
		if ctx.Query("success") == "1" {
			ctx.String(200, large)
			return
		}
		response.ErrorFrom(ctx, service.ErrInvalidCredentials)
	})
	for attempt := 1; attempt <= 10; attempt++ {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		req.RemoteAddr = "192.0.2.1:1234"
		router.ServeHTTP(recorder, req)
		if attempt < 10 {
			require.Equal(t, 401, recorder.Code)
			require.Contains(t, recorder.Body.String(), "INVALID_CREDENTIALS")
		} else {
			require.Equal(t, 403, recorder.Code)
			require.Contains(t, recorder.Body.String(), "IP_BLOCKED")
			require.NotContains(t, recorder.Body.String(), "INVALID_CREDENTIALS")
		}
	}
	repo.blocked = false
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/login?success=1", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	router.ServeHTTP(recorder, req)
	require.Equal(t, 200, recorder.Code)
	require.Equal(t, large, recorder.Body.String())
	require.Equal(t, 10, counter.calls)
}

func TestAuthIPBlacklistMissingCaptchaProofClassification(t *testing.T) {
	for _, reason := range []string{"TURNSTILE_VERIFICATION_FAILED", "TENCENT_CAPTCHA_VERIFICATION_FAILED", "ALIYUN_CAPTCHA_VERIFICATION_FAILED", "UNRELATED"} {
		for _, status := range []int{200, 400, 401, 429, 500, 503} {
			for _, marker := range []string{"", service.AuthIPFailureMissingCaptchaProof, "unknown"} {
				body, err := json.Marshal(response.Response{
					Code: status, Reason: reason, Message: "captcha verification failed",
					Metadata: map[string]string{service.AuthIPFailureMetadataKey: marker},
				})
				require.NoError(t, err)
				counted := status == http.StatusBadRequest && reason != "UNRELATED" && marker == service.AuthIPFailureMissingCaptchaProof
				require.Equal(t, counted, authIPCountFailure(status, body), "status=%d reason=%s marker=%s", status, reason, marker)
			}
		}
	}
}
