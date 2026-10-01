package middleware

import (
	"bytes"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStudioAuditOmitsPrivateBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := &auditCaptureRepository{}
	auditService := service.NewAuditLogService(repository, nil)
	auditService.Start()
	router := gin.New()
	router.Use(func(ctx *gin.Context) { ctx.Set(string(ContextKeyUser), AuthSubject{UserID: 1}); ctx.Next() })
	router.Use(gin.HandlerFunc(NewAuditLogMiddleware(auditService)))
	for _, path := range []string{"/api/v1/studio/submissions", "/api/v1/admin/studio/submissions/:id/review"} {
		router.POST(path, func(ctx *gin.Context) { ctx.Status(http.StatusCreated) })
	}
	for _, path := range []string{"/api/v1/studio/submissions", "/api/v1/admin/studio/submissions/abc/review"} {
		request := httptest.NewRequest("POST", path, bytes.NewBufferString("private-prompt-and-file-canary"))
		request.Header.Set("Content-Type", "multipart/form-data; boundary=test")
		router.ServeHTTP(httptest.NewRecorder(), request)
	}
	auditService.Stop()
	repository.mu.Lock()
	defer repository.mu.Unlock()
	require.Len(t, repository.logs, 2)
	for _, entry := range repository.logs {
		require.Equal(t, "<credential-bearing body omitted>", entry.RequestBody)
		require.NotContains(t, entry.RequestBody, "canary")
		require.Equal(t, http.StatusCreated, entry.StatusCode)
	}
}
