//go:build unit

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCustomGroupAccessClientFallbackHTTPContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, mode := range []string{"allowed", "denied", "messages", "count_tokens", "forced_platform"} {
		t.Run(mode, func(t *testing.T) {
			database, mock, err := sqlmock.New()
			require.NoError(t, err)
			t.Cleanup(func() { _ = database.Close() })
			sourceID, targetID := int64(1), int64(2)
			user := &service.User{ID: 7, Role: service.RoleUser, Status: service.StatusActive, Balance: 95, Concurrency: 5}
			source := &service.Group{ID: sourceID, Platform: service.PlatformAnthropic, Status: service.StatusActive, Hydrated: true, RateMultiplier: 1, ClaudeCodeOnly: true, FallbackGroupID: &targetID}
			target := &service.Group{ID: targetID, Platform: service.PlatformAnthropic, Status: service.StatusActive, Hydrated: true, RateMultiplier: 5}
			key := &service.APIKey{ID: 10, UserID: user.ID, Key: "fallback-key", Status: service.StatusActive, User: user, GroupID: &sourceID, Group: source}
			repo := &stubApiKeyRepo{getByKey: func(context.Context, string) (*service.APIKey, error) { return key, nil }}
			cfg := &config.Config{RunMode: "standard"}
			keys := service.NewAPIKeyService(repo, nil, nil, nil, nil, nil, cfg)
			service.NewCustomGroupAccessService(database, customAccessMiddlewareGroups{groups: map[int64]*service.Group{sourceID: source, targetID: target}}, customAccessMiddlewareUsers{user: user}, nil, keys)
			deferred := mode == "messages" || mode == "count_tokens" || mode == "forced_platform"
			threshold := 90
			if mode == "denied" {
				threshold = 100
			}
			queries := 2
			if deferred {
				queries = 1
			}
			for index := 0; index < queries; index++ {
				rows := sqlmock.NewRows([]string{"group_id", "monthly_recharge_enabled", "monthly_recharge_min", "balance_enabled", "balance_min", "expires_at", "fallback_group_id", "rate_schedule_enabled", "rate_schedule", "updated_at"}).AddRow(targetID, false, 0, true, threshold, nil, nil, true, `[{"starts_at":"2020-01-01T00:00:00Z","ends_at":"2035-01-01T00:00:00Z","rate_multiplier":0.8}]`, time.Now())
				mock.ExpectQuery("FROM custom_group_access_policies").WillReturnRows(rows)
			}
			if !deferred {
				mock.ExpectQuery("SELECT u.balance").WithArgs(user.ID, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"balance", "net_recharge"}).AddRow(95, 0))
			}
			router := gin.New()
			if mode == "forced_platform" {
				router.Use(ForcePlatform(service.PlatformAntigravity))
			}
			router.Use(gin.HandlerFunc(NewAPIKeyAuthMiddleware(keys, nil, cfg)))
			path := "/v1/responses"
			if mode == "messages" {
				path = "/v1/messages"
			} else if mode == "count_tokens" {
				path = "/v1/messages/count_tokens"
			}
			reached := false
			router.POST(path, func(c *gin.Context) {
				reached = true
				actual, ok := GetAPIKeyFromContext(c)
				require.True(t, ok)
				expectedID, expectedRate := targetID, 0.8
				if deferred {
					expectedID, expectedRate = sourceID, 1
				}
				require.Equal(t, expectedID, *actual.GroupID)
				require.Equal(t, expectedRate, actual.Group.RateMultiplier)
				require.Equal(t, actual.Group, c.Request.Context().Value(ctxkey.Group))
				require.Equal(t, sourceID, *key.GroupID)
				require.Equal(t, float64(5), target.RateMultiplier)
				c.Status(http.StatusOK)
			})
			request := httptest.NewRequest(http.MethodPost, path, nil)
			request.Header.Set("Authorization", "Bearer fallback-key")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.NoError(t, mock.ExpectationsWereMet())
			if mode == "denied" {
				require.Equal(t, http.StatusForbidden, response.Code)
				require.False(t, reached)
			} else {
				require.Equal(t, http.StatusOK, response.Code)
				require.True(t, reached)
			}
		})
	}
}

func TestCustomGroupAccessPaidImageReadAndNewGeneration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		t.Run(method, func(t *testing.T) {
			database, mock, err := sqlmock.New()
			require.NoError(t, err)
			t.Cleanup(func() { _ = database.Close() })
			groupID := int64(1)
			user := &service.User{ID: 7, Role: service.RoleUser, Status: service.StatusActive, Balance: 95, Concurrency: 5}
			group := &service.Group{ID: groupID, Platform: service.PlatformOpenAI, Status: service.StatusActive, Hydrated: true, RateMultiplier: 1}
			key := &service.APIKey{ID: 10, UserID: user.ID, Key: "owned-image-key", Status: service.StatusActive, User: user, GroupID: &groupID, Group: group}
			repo := &stubApiKeyRepo{getByKey: func(context.Context, string) (*service.APIKey, error) { return key, nil }}
			cfg := &config.Config{RunMode: "standard"}
			keys := service.NewAPIKeyService(repo, nil, nil, nil, nil, nil, cfg)
			service.NewCustomGroupAccessService(database, customAccessMiddlewareGroups{groups: map[int64]*service.Group{groupID: group}}, customAccessMiddlewareUsers{user: user}, nil, keys)
			if method == http.MethodPost {
				rows := sqlmock.NewRows([]string{"group_id", "monthly_recharge_enabled", "monthly_recharge_min", "balance_enabled", "balance_min", "expires_at", "fallback_group_id", "rate_schedule_enabled", "rate_schedule", "updated_at"}).AddRow(groupID, false, 0, true, 100, nil, nil, false, "[]", time.Now())
				mock.ExpectQuery("FROM custom_group_access_policies").WillReturnRows(rows)
				mock.ExpectQuery("SELECT u.balance").WithArgs(user.ID, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"balance", "net_recharge"}).AddRow(95, 0))
			}
			router := gin.New()
			router.Use(gin.HandlerFunc(NewAPIKeyAuthMiddleware(keys, nil, cfg)))
			reached := false
			router.Handle(method, "/v1/images/tasks/:task_id", func(c *gin.Context) {
				reached = true
				actual, ok := GetAPIKeyFromContext(c)
				require.True(t, ok)
				require.Equal(t, key.ID, actual.ID)
				require.Equal(t, groupID, *actual.GroupID)
				c.JSON(http.StatusOK, gin.H{"status": "succeeded"})
			})
			request := httptest.NewRequest(method, "/v1/images/tasks/imgtask_owned", nil)
			request.Header.Set("Authorization", "Bearer owned-image-key")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.NoError(t, mock.ExpectationsWereMet())
			if method == http.MethodGet {
				require.Equal(t, http.StatusOK, response.Code)
				require.True(t, reached)
			} else {
				require.Equal(t, http.StatusForbidden, response.Code)
				require.False(t, reached)
				require.Contains(t, response.Body.String(), "GROUP_ACCESS_REQUIREMENTS")
			}
		})
	}
}
