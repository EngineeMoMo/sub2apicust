//go:build unit

package handler

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

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
