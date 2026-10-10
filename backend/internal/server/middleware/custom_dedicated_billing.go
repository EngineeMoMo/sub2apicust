package middleware

import (
	"net/http"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func customDedicatedBillingAdmission(c *gin.Context, keys *service.APIKeyService, key *service.APIKey) {
	if !key.IsCustomDedicatedPrepaid() {
		c.Next()
		return
	}
	path := c.Request.URL.Path
	normalized := strings.TrimPrefix(path, "/v1")
	if c.Request.Method == http.MethodGet && (normalized == "/usage" || normalized == "/sub2api/billing" || normalized == "/models" || strings.HasPrefix(normalized, "/models/") || isAsyncImageTaskRead(c.Request.Method, path)) {
		c.Next()
		return
	}
	if err := keys.AdmitCustomDedicatedRequest(c.Request.Context(), key, isResponsesWebSocketRoute(c)); err != nil {
		AbortWithError(c, http.StatusForbidden, infraerrors.Reason(err), err.Error())
		return
	}
	c.Next()
}
