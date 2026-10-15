package middleware

import (
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"strings"
)

func prepareCustomGroupAccess(c *gin.Context, keys *service.APIKeyService, key *service.APIKey, google bool) (*service.APIKey, bool) {
	prepared, err := keys.PrepareCustomGroupAccess(c.Request.Context(), key)
	// 非messages协议不等待Claude客户端识别；切组须发生在订阅及计费上下文装配前。
	path := c.Request.URL.Path
	if err == nil && !strings.HasSuffix(path, "/messages") && !strings.HasSuffix(path, "/messages/count_tokens") {
		prepared, err = keys.PrepareCustomGroupClientFallback(c.Request.Context(), prepared)
	}
	if err == nil {
		SetOpsFallbackAPIKey(c, prepared)
		return prepared, true
	}
	statusCode, status := infraerrors.ToHTTP(err)
	if statusCode >= 500 {
		status.Reason = "GROUP_ACCESS_UNAVAILABLE"
		status.Message = "暂时无法校验分组开放条件，请稍后重试"
	} else {
		service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonAPIKeyGroupUnavailable)
		MarkIngressRejected(c, IngressRejectGroupNotAllowed)
	}
	if google {
		abortWithGoogleError(c, statusCode, status.Message)
	} else {
		AbortWithError(c, statusCode, status.Reason, status.Message)
	}
	return nil, false
}
