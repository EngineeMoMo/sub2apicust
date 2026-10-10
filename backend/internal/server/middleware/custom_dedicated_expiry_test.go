//go:build unit

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCustomDedicatedExpiryAndPublicGroupBilling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, sample := range []struct {
		name    string
		groupID int64
		balance float64
		expired bool
		status  int
		reason  string
	}{
		{"valid_dedicated_zero_balance", 33, 0, false, http.StatusOK, ""},
		{"expired_dedicated_zero_balance", 33, 0, true, http.StatusForbidden, "DEDICATED_ACCOUNT_UNAVAILABLE"},
		{"expired_dedicated_positive_balance", 33, 10, true, http.StatusForbidden, "DEDICATED_ACCOUNT_UNAVAILABLE"},
		{"expired_related_group", 55, 10, true, http.StatusForbidden, "DEDICATED_ACCOUNT_UNAVAILABLE"},
		{"public_group_zero_balance", 77, 0, false, http.StatusForbidden, "INSUFFICIENT_BALANCE"},
		{"public_group_zero_balance_after_expiry", 77, 0, true, http.StatusForbidden, "INSUFFICIENT_BALANCE"},
		{"public_group_paid_after_expiry", 77, 10, true, http.StatusOK, ""},
	} {
		t.Run(sample.name, func(t *testing.T) {
			database, mock, err := sqlmock.New()
			require.NoError(t, err)
			t.Cleanup(func() { _ = database.Close() })
			cfg := &config.Config{RunMode: config.RunModeStandard}
			user := &service.User{ID: 11, Status: service.StatusActive, Balance: sample.balance, AllowedGroups: []int64{33, 55, 77}}
			group := &service.Group{ID: sample.groupID, Platform: service.PlatformOpenAI, Status: service.StatusActive, IsExclusive: true, SubscriptionType: service.SubscriptionTypeStandard, Hydrated: true}
			cachedKey := &service.APIKey{ID: 7, UserID: user.ID, Key: "expiry-test-key", Status: service.StatusActive, User: user, Group: group, GroupID: &group.ID}
			repo := &stubApiKeyRepo{getByKey: func(context.Context, string) (*service.APIKey, error) { return cachedKey, nil }}
			keys := service.NewAPIKeyService(repo, nil, nil, nil, nil, nil, cfg)
			service.NewCustomDedicatedService(database, nil, keys, &service.GatewayService{}, &service.OpenAIGatewayService{})
			now := time.Now()
			expires := now.Add(time.Hour)
			if sample.expired {
				expires = now.Add(-time.Second)
			}
			bindingRows := func(include bool) *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "user_id", "account_id", "group_id", "label", "expires_at", "revoked_at", "updated_at", "user_ids"})
				if include {
					rows.AddRow(1, 11, 22, 33, "包号", expires, nil, now, []byte(`[11]`))
				}
				return rows
			}
			dedicated := sample.groupID != 77
			mock.ExpectQuery(`SELECT .* FROM custom_dedicated_accounts WHERE [(]group_id`).WithArgs(sample.groupID, int64(0), int64(0)).WillReturnRows(bindingRows(dedicated))
			if dedicated && !sample.expired {
				mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"valid"}).AddRow(true))
			}
			if !dedicated || !sample.expired {
				mock.ExpectQuery("SELECT .* FROM custom_dedicated_accounts WHERE deleted_at IS NULL").WithArgs(sample.groupID).WillReturnRows(bindingRows(dedicated))
				if dedicated {
					mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"valid"}).AddRow(true))
					mock.ExpectBegin()
					mock.ExpectQuery("SELECT .* FROM custom_dedicated_accounts WHERE id=.*FOR UPDATE").WithArgs(int64(1)).WillReturnRows(bindingRows(true))
					mock.ExpectQuery("SELECT clock_timestamp").WillReturnRows(sqlmock.NewRows([]string{"accepted_at"}).AddRow(now))
					mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"valid"}).AddRow(true))
					mock.ExpectExec("INSERT INTO custom_dedicated_request_leases").WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectCommit()
				}
			}
			reached := false
			router := gin.New()
			router.Use(gin.HandlerFunc(NewAPIKeyAuthMiddleware(keys, nil, cfg)))
			router.POST("/v1/responses", func(ctx *gin.Context) {
				reached = true
				key, ok := GetAPIKeyFromContext(ctx)
				require.True(t, ok)
				require.Equal(t, dedicated, key.IsCustomDedicatedPrepaid())
				ctx.Status(http.StatusOK)
			})
			request := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(`{}`))
			request.Header.Set("x-api-key", cachedKey.Key)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			require.Equal(t, sample.status, recorder.Code, recorder.Body.String())
			require.Equal(t, sample.status == http.StatusOK, reached)
			if sample.reason != "" {
				require.Contains(t, recorder.Body.String(), sample.reason)
			}
			require.False(t, cachedKey.IsCustomDedicatedPrepaid())
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
