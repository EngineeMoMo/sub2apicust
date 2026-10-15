package middleware

import (
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func prepareCustomGroupAccess(c *gin.Context, keys *service.APIKeyService, key *service.APIKey, google bool) (*service.APIKey, bool) {
	prepared, err := keys.PrepareCustomGroupAccess(c.Request.Context(), key)
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
