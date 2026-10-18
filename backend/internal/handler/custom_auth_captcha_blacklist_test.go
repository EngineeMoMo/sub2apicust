//go:build unit

package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type authCaptchaProbeRepository struct {
	service.AuthIPBlacklistRepository
	enabled bool
	entry   *service.AuthIPBlacklistEntry
}

func (repo *authCaptchaProbeRepository) Snapshot(context.Context, string) (*service.AuthIPBlacklistSnapshot, error) {
	return &service.AuthIPBlacklistSnapshot{
		Config: service.AuthIPBlacklistConfig{Enabled: repo.enabled, Threshold: 10, WindowSeconds: 7200, Version: 1, BlockMode: service.AuthIPBlockModePermanent},
		Entry:  repo.entry,
	}, nil
}

func (repo *authCaptchaProbeRepository) Block(_ context.Context, ip, path string, _ service.AuthIPBlacklistSnapshot, count service.AuthIPWindowCount) (*service.AuthIPBlacklistDecision, error) {
	repo.entry = &service.AuthIPBlacklistEntry{ID: 1, IP: ip, Blocked: true, Generation: 1, TriggerPath: path, TriggerCount: count.Count, WindowSeconds: 7200}
	return &service.AuthIPBlacklistDecision{Entry: repo.entry, NewlyBlocked: true}, nil
}

type authCaptchaProbeCounter struct{ calls int }

func (counter *authCaptchaProbeCounter) Time(context.Context) (time.Time, error) {
	return time.Now().UTC(), nil
}

func (counter *authCaptchaProbeCounter) Count(context.Context, string, int64, service.AuthIPBlacklistConfig) (*service.AuthIPWindowCount, error) {
	counter.calls++
	return &service.AuthIPWindowCount{Count: counter.calls}, nil
}

func (counter *authCaptchaProbeCounter) CountAt(ctx context.Context, ip string, generation int64, cfg service.AuthIPBlacklistConfig, _ time.Time) (*service.AuthIPWindowCount, error) {
	return counter.Count(ctx, ip, generation, cfg)
}

type authCaptchaProbeVerifier struct {
	calls int
	err   error
}

func (verifier *authCaptchaProbeVerifier) VerifyToken(context.Context, string, string, string) (*service.TurnstileVerifyResponse, error) {
	verifier.calls++
	return &service.TurnstileVerifyResponse{Success: false}, verifier.err
}

func TestCustomAuthCaptchaLoginBlacklist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name       string
		automatic  bool
		configured bool
		proof      string
		verifyErr  error
		status     int
		blocks     bool
	}{
		{name: "缺少验证凭证的默认账号扫描达到阈值封禁", automatic: true, configured: true, status: 400, blocks: true},
		{name: "人机挑战未通过不计数", automatic: true, configured: true, proof: "rejected-user-proof", status: 400},
		{name: "验证服务故障不计数", automatic: true, configured: true, proof: "user-proof", verifyErr: errors.New("verifier offline"), status: 500},
		{name: "服务端未配置不计数", automatic: true, status: 503},
		{name: "自动规则关闭仍不新增封禁", configured: true, status: 400},
	} {
		t.Run(test.name, func(t *testing.T) {
			values := map[string]string{service.SettingKeyTurnstileEnabled: "true"}
			if test.configured {
				values[service.SettingKeyTurnstileSecretKey] = "local-test-secret"
			}
			cfg := &config.Config{}
			settings := service.NewSettingService(&oauthCaptchaSettingRepo{values: values}, cfg)
			verifier := &authCaptchaProbeVerifier{err: test.verifyErr}
			turnstile := service.NewTurnstileService(settings, verifier)
			auth := service.NewAuthService(nil, nil, nil, nil, cfg, settings, nil, turnstile, nil, nil, nil, nil, nil)
			handler := &AuthHandler{authService: auth, settingSvc: settings, cfg: cfg}
			repo := &authCaptchaProbeRepository{enabled: test.automatic}
			counter := &authCaptchaProbeCounter{}
			router := gin.New()
			require.NoError(t, router.SetTrustedProxies(nil))
			router.POST("/api/v1/auth/login", middleware.NewAuthIPBlacklistGuard(service.NewAuthIPBlacklistService(repo, counter)), handler.Login)
			payload := `{"email":"admin@sub2api.local","password":"local-test-password"}`
			if test.proof != "" {
				payload = strings.TrimSuffix(payload, "}") + `,"turnstile_token":"` + test.proof + `"}`
			}
			for attempt := 1; attempt <= 12; attempt++ {
				recorder := httptest.NewRecorder()
				request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(payload))
				request.Header.Set("Content-Type", "application/json")
				request.Header.Set("User-Agent", "Go-http-client/1.1")
				request.RemoteAddr = "198.51.100.24:12345"
				router.ServeHTTP(recorder, request)
				if test.blocks && attempt >= 10 {
					require.Equal(t, http.StatusForbidden, recorder.Code)
					require.Contains(t, recorder.Body.String(), "IP_BLOCKED")
				} else {
					require.Equal(t, test.status, recorder.Code)
				}
			}
			if test.blocks {
				require.Equal(t, 10, counter.calls)
				require.NotNil(t, repo.entry)
				require.Nil(t, repo.entry.BlockedUntil)
			} else {
				require.Zero(t, counter.calls)
				require.Nil(t, repo.entry)
			}
			if test.proof == "" {
				require.Zero(t, verifier.calls)
			}
		})
	}
}
