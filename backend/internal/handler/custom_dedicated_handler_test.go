//go:build unit

package handler

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCustomDedicatedHandlerConfigDetails(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	err := infraerrors.BadRequest("DEDICATED_ACCOUNT_CONFIG", "所选账号还关联其他分组，其他分组编号：55").WithMetadata(map[string]string{"config_issue": "account_other_groups", "resource_ids": "55"})
	customDedicatedError(ctx, err)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), "所选账号还关联其他分组")
	require.Contains(t, recorder.Body.String(), `"config_issue":"account_other_groups"`)
	require.Contains(t, recorder.Body.String(), `"resource_ids":"55"`)
}

func TestCustomDedicatedHandlerUnauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &CustomDedicatedHandler{}
	router := gin.New()
	router.GET("/dedicated-accounts", handler.UserList)
	router.GET("/dedicated-accounts/:id/usage", handler.UserUsage)
	for _, path := range []string{"/dedicated-accounts", "/dedicated-accounts/1/usage"} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		require.Equal(t, http.StatusUnauthorized, recorder.Code)
	}
}

func TestCustomDedicatedHandlerOtherOwnerNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	dedicated := service.NewCustomDedicatedService(db, nil, &service.APIKeyService{}, &service.GatewayService{}, &service.OpenAIGatewayService{})
	handler := NewCustomDedicatedHandler(dedicated)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 99})
		c.Next()
	})
	router.GET("/dedicated-accounts/:id/usage", handler.UserUsage)
	mock.ExpectQuery("SELECT .* FROM custom_dedicated_accounts WHERE id=").WithArgs(int64(1), int64(99)).WillReturnError(sql.ErrNoRows)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/dedicated-accounts/1/usage?user_id=11", nil))
	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.NotContains(t, recorder.Body.String(), "account_id")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomDedicatedHandlerInvalidPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &CustomDedicatedHandler{}
	router := gin.New()
	router.GET("/admin/dedicated-accounts", handler.AdminList)
	for _, page := range []string{"0", "-1", "no", "1000001"} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/admin/dedicated-accounts?page="+page, nil))
		require.Equal(t, http.StatusBadRequest, recorder.Code)
	}
}

func TestCustomDedicatedHandlerEditErrors(t *testing.T) {
	for _, test := range []struct {
		err    error
		status int
		reason string
	}{
		{service.ErrDedicatedStale, http.StatusConflict, "DEDICATED_ACCOUNT_STALE"},
		{service.ErrDedicatedRestore, http.StatusBadRequest, "DEDICATED_ACCOUNT_RESTORE"},
		{service.ErrDedicatedPolicy, http.StatusBadRequest, "DEDICATED_BILLING_POLICY"},
	} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		customDedicatedError(ctx, test.err)
		require.Equal(t, test.status, recorder.Code)
		require.Contains(t, recorder.Body.String(), test.reason)
	}
}

func TestCustomDedicatedHandlerBillingPolicyRejectsInvalidPayloadAndID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dedicated := service.NewCustomDedicatedService(nil, nil, &service.APIKeyService{}, &service.GatewayService{}, &service.OpenAIGatewayService{})
	handler := NewCustomDedicatedHandler(dedicated)
	router := gin.New()
	router.PUT("/admin/dedicated-accounts/:id/billing-policy", handler.BillingPolicy)
	for _, tc := range []struct{ id, body string }{{"no", `{}`}, {"1", `{"concurrency_limit":0}`}, {"1", `{"concurrency_limit":"bad"}`}} {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/admin/dedicated-accounts/"+tc.id+"/billing-policy", strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
	}
}
