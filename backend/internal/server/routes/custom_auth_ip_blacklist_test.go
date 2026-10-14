//go:build unit

package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type authIPRouteUserRepo struct {
	service.UserRepository
	user *service.User
}

func (repo *authIPRouteUserRepo) GetByID(context.Context, int64) (*service.User, error) {
	return repo.user, nil
}

func (repo *authIPRouteUserRepo) GetUserAvatar(context.Context, int64) (*service.UserAvatar, error) {
	return nil, nil
}

type authIPRouteRepo struct {
	service.AuthIPBlacklistRepository
}

func (repo *authIPRouteRepo) Config(context.Context) (*service.AuthIPBlacklistConfig, error) {
	return &service.AuthIPBlacklistConfig{Threshold: 10, WindowSeconds: 7200, Version: 1, BlockMode: service.AuthIPBlockModeTimed}, nil
}
func (repo *authIPRouteRepo) UpdateConfig(_ context.Context, config service.AuthIPBlacklistConfig) (*service.AuthIPBlacklistConfig, error) {
	if config.Version != 1 {
		return nil, service.ErrAuthIPBlacklistConflict
	}
	config.Version++
	return &config, nil
}
func (repo *authIPRouteRepo) Unblock(_ context.Context, id, generation int64) (*service.AuthIPBlacklistEntry, error) {
	if generation != 3 {
		return nil, service.ErrAuthIPBlacklistConflict
	}
	return &service.AuthIPBlacklistEntry{ID: id, IP: "61.120.9.25", Generation: 4}, nil
}

func (repo *authIPRouteRepo) ManualBlock(_ context.Context, input service.AuthIPManualBlock) (*service.AuthIPBlacklistEntry, error) {
	if input.IP == "61.120.9.25" {
		return nil, service.ErrAuthIPAlreadyBlocked
	}
	return &service.AuthIPBlacklistEntry{ID: 2, IP: input.IP, Blocked: true, Generation: 1, Source: "manual", Reason: input.Reason}, nil
}

type authIPRouteAuditRepo struct {
	service.AuditLogRepository
	mu   sync.Mutex
	logs []*service.AuditLog
}

func (repo *authIPRouteAuditRepo) BatchInsert(_ context.Context, logs []*service.AuditLog) (int64, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.logs = append(repo.logs, logs...)
	return int64(len(logs)), nil
}

func TestAuthIPBlacklistAdminRoutesRequireRealAdminSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	user := &service.User{ID: 1, Email: "guard@test.invalid", Role: service.RoleUser, Status: service.StatusActive}
	users := service.NewUserService(&authIPRouteUserRepo{user: user}, nil, nil, nil)
	auth := service.NewAuthService(nil, nil, nil, nil, &config.Config{JWT: config.JWTConfig{Secret: "auth-ip-test-secret", ExpireHour: 1}}, nil, nil, nil, nil, nil, nil, nil, nil)
	router := gin.New()
	auditRepo := &authIPRouteAuditRepo{}
	auditService := service.NewAuditLogService(auditRepo, nil)
	auditService.Start()
	audit := middleware.NewAuditLogMiddleware(auditService)
	RegisterAuthIPBlacklistAdminRoutes(router.Group("/api/v1"), admin.NewAuthIPBlacklistHandler(service.NewAuthIPBlacklistService(&authIPRouteRepo{}, nil)), middleware.NewAdminAuthMiddleware(auth, users, nil, nil), audit, nil, nil)
	request := func(method, path, token, body string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.RemoteAddr = "61.120.9.25:1234"
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		router.ServeHTTP(recorder, req)
		return recorder
	}
	require.Equal(t, 401, request(http.MethodGet, "/api/v1/admin/ip-blacklist/config", "", "").Code)
	require.Equal(t, 401, request(http.MethodPost, "/api/v1/admin/ip-blacklist/entries", "", `{}`).Code)
	token, err := auth.GenerateToken(context.Background(), user)
	require.NoError(t, err)
	require.Equal(t, 403, request(http.MethodGet, "/api/v1/admin/ip-blacklist/config", token, "").Code)
	require.Equal(t, 403, request(http.MethodPost, "/api/v1/admin/ip-blacklist/entries", token, `{}`).Code)
	user.Role = service.RoleAdmin
	token, err = auth.GenerateToken(context.Background(), user)
	require.NoError(t, err)
	require.Equal(t, 200, request(http.MethodGet, "/api/v1/admin/ip-blacklist/config", token, "").Code)
	require.Equal(t, 200, request(http.MethodPut, "/api/v1/admin/ip-blacklist/config", token, `{"enabled":true,"threshold":10,"window_seconds":7200,"block_mode":"permanent","version":1,"aggregation_mode":"network","network_ipv4_prefix":24,"network_ipv6_prefix":64}`).Code)
	require.Equal(t, 400, request(http.MethodPut, "/api/v1/admin/ip-blacklist/config", token, `{"threshold":0,"window_seconds":7200,"version":1}`).Code)
	require.Equal(t, 409, request(http.MethodPut, "/api/v1/admin/ip-blacklist/config", token, `{"enabled":true,"threshold":10,"window_seconds":7200,"block_mode":"timed","version":2,"aggregation_mode":"ip","network_ipv4_prefix":24,"network_ipv6_prefix":64}`).Code)
	require.Equal(t, 409, request(http.MethodPost, "/api/v1/admin/ip-blacklist/entries/1/unblock", token, `{"generation":1}`).Code)
	require.Equal(t, 200, request(http.MethodPost, "/api/v1/admin/ip-blacklist/entries/1/unblock", token, `{"generation":3}`).Code)
	require.Equal(t, 200, request(http.MethodPost, "/api/v1/admin/ip-blacklist/entries", token, `{"ip":"::ffff:192.0.2.40","block_mode":"permanent","reason":"人工确认恶意扫描"}`).Code)
	require.Equal(t, 400, request(http.MethodPost, "/api/v1/admin/ip-blacklist/entries", token, `{"ip":"192.0.*.4","block_mode":"permanent","reason":"扫描"}`).Code)
	require.Equal(t, 409, request(http.MethodPost, "/api/v1/admin/ip-blacklist/entries", token, `{"ip":"61.120.9.25","block_mode":"permanent","reason":"扫描"}`).Code)
	require.Equal(t, 400, request(http.MethodPut, "/api/v1/admin/ip-blacklist/config", token, `{"enabled":true,"threshold":10,"window_seconds":7200,"block_mode":"timed","version":1}`).Code)
	auditService.Stop()
	auditRepo.mu.Lock()
	defer auditRepo.mu.Unlock()
	require.Len(t, auditRepo.logs, 9)
	require.Equal(t, "admin.ip_blacklist.config.update", auditRepo.logs[0].Action)
	require.EqualValues(t, 7200, auditRepo.logs[0].Extra["window_seconds"])
	require.EqualValues(t, 2, auditRepo.logs[0].Extra["config_version"])
	require.Equal(t, "permanent", auditRepo.logs[0].Extra["block_mode"])
	require.Equal(t, "network", auditRepo.logs[0].Extra["aggregation_mode"])
	require.Equal(t, "admin.ip_blacklist.manual_block", auditRepo.logs[5].Action)
	require.Equal(t, "192.0.2.40", auditRepo.logs[5].Extra["target_ip"])
	require.Equal(t, "manual", auditRepo.logs[5].Extra["source"])
	require.Equal(t, "人工确认恶意扫描", auditRepo.logs[5].Extra["reason"])
	require.Equal(t, "admin.ip_blacklist.unblock", auditRepo.logs[4].Action)
	require.Equal(t, "61.120.9.25", auditRepo.logs[4].Extra["target_ip"])
	require.EqualValues(t, 4, auditRepo.logs[4].Extra["generation"])
}
