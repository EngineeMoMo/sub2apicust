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
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type customAdminLoginTotpCache struct {
	service.TotpCache
	err      error
	sessions int
}

func (cache *customAdminLoginTotpCache) SetLoginSession(context.Context, string, *service.TotpLoginSession, time.Duration) error {
	cache.sessions++
	return cache.err
}

func TestCustomAdminLoginCannotBypassBoundTotpWithGlobalSwitchOff(t *testing.T) {
	for _, scenario := range []struct {
		name           string
		missingService bool
		cacheError     error
		status         int
	}{
		{"challenge", false, nil, http.StatusOK},
		{"missing_service", true, nil, http.StatusServiceUnavailable},
		{"cache_unavailable", false, errors.New("cache unavailable"), http.StatusInternalServerError},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			user := &service.User{ID: 1, Email: "admin@example.com", Role: service.RoleAdmin, Status: service.StatusActive, TotpEnabled: true}
			require.NoError(t, user.SetPassword("test-password"))
			repo := &userHandlerRepoStub{user: user}
			cfg := &config.Config{JWT: config.JWTConfig{Secret: "custom-admin-totp-test-secret", ExpireHour: 1}}
			settings := service.NewSettingService(&oauthCaptchaSettingRepo{values: map[string]string{service.SettingKeyTotpEnabled: "false"}}, cfg)
			auth := service.NewAuthService(nil, repo, nil, nil, cfg, nil, nil, nil, nil, nil, nil, nil, nil)
			cache := &customAdminLoginTotpCache{err: scenario.cacheError}
			handler := &AuthHandler{cfg: cfg, authService: auth, settingSvc: settings}
			if !scenario.missingService {
				handler.totpService = service.NewTotpService(repo, nil, cache, settings, nil, nil)
			}
			response := httptest.NewRecorder()
			request, _ := gin.CreateTestContext(response)
			request.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"admin@example.com","password":"test-password"}`))
			request.Request.Header.Set("Content-Type", "application/json")
			handler.Login(request)
			require.Equal(t, scenario.status, response.Code)
			require.NotContains(t, response.Body.String(), "access_token")
			if scenario.status == http.StatusOK {
				require.Contains(t, response.Body.String(), `"requires_2fa":true`)
				require.Equal(t, 1, cache.sessions)
			}
		})
	}
}
