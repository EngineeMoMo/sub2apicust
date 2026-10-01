package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"github.com/Wei-Shaw/sub2api/internal/pkg/requestmodel"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// 原有鉴权已完成；普通组直接进入原handler。发现／用量读取不消耗包号请求次数。
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
	websocket := isResponsesWebSocketRoute(c)
	var payload []byte
	if c.Request.Method != http.MethodGet {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, key.CustomDedicatedMaxBodyBytes())
		var err error
		payload, err = httputil.ReadRequestBodyWithPrealloc(c.Request)
		if err != nil {
			status := http.StatusBadRequest
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				status = http.StatusRequestEntityTooLarge
			}
			AbortWithError(c, status, "DEDICATED_REQUEST_BODY", "包号请求读取失败或超过请求体限制")
			return
		}
		requestmodel.ResetRequestBody(c.Request, payload)
	}
	if err := keys.AdmitCustomDedicatedHTTP(c.Request.Context(), key, path, payload, websocket, c.GetHeader("Content-Type")); err != nil {
		status := http.StatusForbidden
		if errors.Is(err, service.ErrDedicatedLimit) {
			status = http.StatusTooManyRequests
		}
		if errors.Is(err, service.ErrDedicatedPolicy) {
			status = http.StatusBadRequest
		}
		AbortWithError(c, status, infraerrors.Reason(err), err.Error())
		return
	}
	requestCtx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()
	c.Request = c.Request.WithContext(requestCtx)
	release := keys.HoldCustomDedicatedLease(requestCtx, key, cancel)
	defer release()
	c.Next()
}
