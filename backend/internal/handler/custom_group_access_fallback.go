package handler

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/requestmodel"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *GatewayHandler) prepareCustomGatewayGroup(c *gin.Context, key *service.APIKey) (*service.APIKey, bool) {
	prepared, err := h.gatewayService.PrepareCustomGatewayGroup(c.Request.Context(), key)
	if err != nil {
		status, code, message, _ := billingErrorDetails(err)
		h.errorResponse(c, status, code, message)
		return nil, false
	}
	if prepared != key {
		subscription, err := h.apiKeyService.CustomGroupFallbackSubscription(c.Request.Context(), prepared)
		if err != nil {
			status, code, message, _ := billingErrorDetails(err)
			h.errorResponse(c, status, code, message)
			return nil, false
		}
		setCustomGatewayGroup(c, prepared)
		c.Set(string(middleware2.ContextKeySubscription), subscription)
	}
	return prepared, true
}

func setCustomGatewayGroup(c *gin.Context, key *service.APIKey) {
	c.Set(string(middleware2.ContextKeyAPIKey), key)
	c.Set(string(middleware2.ContextKeySubscription), nil)
	middleware2.SetOpsFallbackAPIKey(c, key)
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ctxkey.Group, key.Group))
}

func customGatewayGroupBlocksModel(c *gin.Context, key *service.APIKey, body []byte) bool {
	if key == nil || !key.Group.ModelAllowlistEnabled() {
		return false
	}
	for _, model := range requestmodel.FromBodyCandidates(c.FullPath(), c.GetHeader("Content-Type"), body) {
		if !key.Group.ModelAllowlist.Allows(model) {
			service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonLocalModelConfiguration)
			middleware2.MarkIngressRejected(c, middleware2.IngressRejectModelNotAllowed)
			return true
		}
	}
	return false
}
