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

type customAccessMiddlewareGroups struct {
	service.GroupRepository
	groups map[int64]*service.Group
}

func (r customAccessMiddlewareGroups) GetByIDLite(_ context.Context, id int64) (*service.Group, error) {
	if group := r.groups[id]; group != nil {
		return group, nil
	}
	return nil, service.ErrGroupNotFound
}

type customAccessMiddlewareUsers struct {
	service.UserRepository
	user *service.User
}

func (r customAccessMiddlewareUsers) GetByID(context.Context, int64) (*service.User, error) {
	copy := *r.user
	return &copy, nil
}

func TestCustomGroupAccessHTTPRefundAndExpiry(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, google := range []bool{false, true} {
		for _, expired := range []bool{false, true} {
			name := "normal"
			if google {
				name = "google"
			}
			if expired {
				name += "_expiry"
			} else {
				name += "_refund"
			}
			t.Run(name, func(t *testing.T) {
				database, mock, err := sqlmock.New()
				require.NoError(t, err)
				t.Cleanup(func() { _ = database.Close() })
				sourceID, targetID := int64(1), int64(2)
				user := &service.User{ID: 7, Role: service.RoleUser, Status: service.StatusActive, Balance: 100, Concurrency: 5}
				source := &service.Group{ID: sourceID, Name: "original", Platform: service.PlatformGemini, Status: service.StatusActive, Hydrated: true, RateMultiplier: 2}
				target := &service.Group{ID: targetID, Name: "fallback", Platform: service.PlatformGemini, Status: service.StatusActive, Hydrated: true, RateMultiplier: 0.5}
				original := &service.APIKey{ID: 10, UserID: user.ID, Key: "group-policy-key", Status: service.StatusActive, User: user, GroupID: &sourceID, Group: source}
				repo := &stubApiKeyRepo{getByKey: func(context.Context, string) (*service.APIKey, error) { return original, nil }}
				cfg := &config.Config{RunMode: "standard"}
				keys := service.NewAPIKeyService(repo, nil, nil, nil, nil, nil, cfg)
				service.NewCustomGroupAccessService(database, customAccessMiddlewareGroups{groups: map[int64]*service.Group{1: source, 2: target}}, customAccessMiddlewareUsers{user: user}, nil, keys)
				rows := sqlmock.NewRows([]string{"group_id", "monthly_recharge_enabled", "monthly_recharge_min", "balance_enabled", "balance_min", "expires_at", "fallback_group_id", "rate_schedule_enabled", "rate_schedule", "updated_at"})
				if expired {
					rows.AddRow(sourceID, false, 0, false, 0, time.Now().Add(-time.Minute), targetID, false, "[]", time.Now())
				} else {
					rows.AddRow(sourceID, true, 100, false, 0, nil, nil, false, "[]", time.Now())
				}
				mock.ExpectQuery("FROM custom_group_access_policies").WillReturnRows(rows)
				if !expired {
					mock.ExpectQuery("SELECT u.balance").WithArgs(user.ID, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"balance", "net_recharge"}).AddRow(100, 0))
				}
				router := gin.New()
				if google {
					router.Use(APIKeyAuthGoogle(keys, cfg))
				} else {
					router.Use(gin.HandlerFunc(NewAPIKeyAuthMiddleware(keys, nil, cfg)))
				}
				reached := false
				router.GET("/v1/sub2api/billing", func(c *gin.Context) {
					reached = true
					value, ok := c.Get(string(ContextKeyAPIKey))
					require.True(t, ok)
					resolved := value.(*service.APIKey)
					require.Equal(t, targetID, *resolved.GroupID)
					require.Equal(t, 0.5, resolved.Group.RateMultiplier)
					group := c.Request.Context().Value(ctxkey.Group).(*service.Group)
					require.Equal(t, targetID, group.ID)
					c.Status(http.StatusOK)
				})
				request := httptest.NewRequest(http.MethodGet, "/v1/sub2api/billing", nil)
				request.Header.Set("Authorization", "Bearer group-policy-key")
				response := httptest.NewRecorder()
				router.ServeHTTP(response, request)
				if expired {
					require.Equal(t, http.StatusOK, response.Code, response.Body.String())
					require.True(t, reached)
				} else {
					require.Equal(t, http.StatusForbidden, response.Code, response.Body.String())
					require.False(t, reached)
				}
				require.Equal(t, sourceID, *original.GroupID)
				require.Equal(t, float64(2), original.Group.RateMultiplier)
				require.NoError(t, mock.ExpectationsWereMet())
			})
		}
	}
}
