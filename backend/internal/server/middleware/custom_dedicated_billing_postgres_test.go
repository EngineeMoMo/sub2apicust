//go:build unit

package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestCustomDedicatedBillingAuthPostgres(t *testing.T) {
	dsn := os.Getenv("DEDICATED_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires an isolated PostgreSQL database")
	}
	u, err := url.Parse(dsn)
	require.NoError(t, err)
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	schema := fmt.Sprintf("dedicated_auth_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	db, err := sql.Open("postgres", u.String())
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
		_, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE")
		require.NoError(t, err)
		_ = admin.Close()
	})
	for _, file := range []string{"../../service/testdata/custom_dedicated_billing.sql", "../../../migrations/241_custom_dedicated_accounts.sql", "../../../migrations/242_custom_dedicated_members.sql", "../../../migrations/243_custom_dedicated_billing.sql"} {
		body, err := os.ReadFile(file)
		require.NoError(t, err)
		_, err = db.Exec(string(body))
		require.NoError(t, err)
	}
	_, err = db.Exec(`INSERT INTO custom_dedicated_accounts(user_id,user_ids,account_id,group_id,label,expires_at) VALUES(11,'[11,12]',22,33,'认证测试',NOW()+INTERVAL '1 day')`)
	require.NoError(t, err)
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{RunMode: config.RunModeStandard}
	user := &service.User{ID: 11, Status: service.StatusActive, AllowedGroups: []int64{33, 55, 77}}
	group := &service.Group{ID: 33, Platform: service.PlatformOpenAI, Status: service.StatusActive, IsExclusive: true, SubscriptionType: service.SubscriptionTypeStandard, Hydrated: true}
	key := &service.APIKey{ID: 7, UserID: 11, Key: "test-dedicated", Status: service.StatusActive, User: user, Group: group, GroupID: &group.ID}
	publicGroup := *group
	publicGroup.ID = 77
	publicKey := *key
	publicKey.Key = "test-public"
	publicKey.Group = &publicGroup
	publicKey.GroupID = &publicGroup.ID
	repo := &stubApiKeyRepo{getByKey: func(_ context.Context, credential string) (*service.APIKey, error) {
		if credential == key.Key {
			return key, nil
		}
		if credential == publicKey.Key {
			return &publicKey, nil
		}
		return nil, service.ErrAPIKeyNotFound
	}}
	keys := service.NewAPIKeyService(repo, nil, nil, nil, nil, nil, cfg)
	dedicated := service.NewCustomDedicatedService(db, nil, keys, &service.GatewayService{}, &service.OpenAIGatewayService{})
	router := gin.New()
	router.Use(gin.HandlerFunc(NewAPIKeyAuthMiddleware(keys, nil, cfg)))
	for _, path := range []string{"/v1/responses", "/v1/messages", "/backend-api/codex/responses", "/v1/images/generations/async"} {
		router.POST(path, func(c *gin.Context) {
			authKey, ok := GetAPIKeyFromContext(c)
			require.True(t, ok)
			if authKey.Key == key.Key {
				require.True(t, authKey.IsCustomDedicatedPrepaid())
				require.NotEmpty(t, authKey.CustomDedicatedUsageRequestID("upstream"))
			}
			body, err := io.ReadAll(c.Request.Body)
			require.NoError(t, err)
			c.Data(http.StatusOK, "application/json", body)
		})
	}
	request := func(path, credential, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
		req.Header.Set("x-api-key", credential)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		return w
	}
	t.Run("zero_balance_admitted_body_restored_and_slot_released", func(t *testing.T) {
		w := request("/v1/responses", key.Key, `{"model":"gpt-test","stream":true}`)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		require.JSONEq(t, `{"model":"gpt-test","stream":true}`, w.Body.String())
		require.False(t, key.IsCustomDedicatedPrepaid(), "authentication cache object must stay ordinary")
		var active int
		require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM custom_dedicated_request_leases WHERE finished_at IS NULL").Scan(&active))
		require.Zero(t, active)
	})
	t.Run("same_user_public_group_still_requires_balance", func(t *testing.T) {
		w := request("/v1/responses", publicKey.Key, `{}`)
		require.Equal(t, http.StatusForbidden, w.Code)
		require.Contains(t, w.Body.String(), "INSUFFICIENT_BALANCE")
		user.Balance = 10
		w = request("/v1/responses", publicKey.Key, `{}`)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		user.Balance = 0
	})
	t.Run("oversized_images_and_unsupported_async_rejected", func(t *testing.T) {
		p, err := dedicated.BillingPolicy(context.Background(), 1)
		require.NoError(t, err)
		p.MaxBodyBytes = 1024
		_, err = dedicated.UpdateBillingPolicy(context.Background(), 1, p)
		require.NoError(t, err)
		require.Equal(t, http.StatusRequestEntityTooLarge, request("/v1/responses", key.Key, strings.Repeat("a", 1025)).Code)
		require.Equal(t, http.StatusForbidden, request("/v1/responses", key.Key, `{"tools":[{"type":"image_generation"}]}`).Code)
		require.Equal(t, http.StatusForbidden, request("/v1/images/generations/async", key.Key, `{}`).Code)
	})
	t.Run("expired_quota_key_and_revoked_entitlement_do_not_become_free", func(t *testing.T) {
		expired := time.Now().Add(-time.Hour)
		key.ExpiresAt = &expired
		require.Equal(t, http.StatusForbidden, request("/v1/messages", key.Key, `{}`).Code)
		key.ExpiresAt = nil
		key.Quota = 1
		key.QuotaUsed = 1
		require.Equal(t, http.StatusTooManyRequests, request("/v1/messages", key.Key, `{}`).Code)
		key.Quota = 0
		key.QuotaUsed = 0
		_, err := db.Exec("UPDATE custom_dedicated_accounts SET revoked_at=NOW()")
		require.NoError(t, err)
		w := request("/backend-api/codex/responses", key.Key, `{}`)
		require.Equal(t, http.StatusForbidden, w.Code)
		require.Contains(t, w.Body.String(), "DEDICATED_ACCOUNT_UNAVAILABLE")
	})
}
