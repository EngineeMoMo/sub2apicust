//go:build unit

package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCustomGroupAccessFallbackModelAndContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, mode := range []string{"allowed", "blocked", "duplicate", "case_variant", "disabled"} {
		t.Run(mode, func(t *testing.T) {
			context, _ := gin.CreateTestContext(httptest.NewRecorder())
			context.Request = httptest.NewRequest("POST", "/v1/messages", nil)
			context.Request.Header.Set("Content-Type", "application/json")
			groupID := int64(2)
			group := &service.Group{ID: groupID, ModelAllowlist: service.GroupModelAllowlist{Enabled: mode != "disabled", Models: []string{"claude-sonnet-*"}}}
			key := &service.APIKey{GroupID: &groupID, Group: group}
			body := []byte(`{"model":"claude-sonnet-4-6"}`)
			switch mode {
			case "blocked", "disabled":
				body = []byte(`{"model":"claude-opus-4-6"}`)
			case "duplicate":
				body = []byte(`{"model":"claude-sonnet-4-6","model":"claude-opus-4-6"}`)
			case "case_variant":
				body = []byte(`{"model":"claude-sonnet-4-6","Model":"claude-opus-4-6"}`)
			}
			require.Equal(t, mode != "allowed" && mode != "disabled", customGatewayGroupBlocksModel(context, key, body))
			context.Set(string(middleware2.ContextKeySubscription), &service.UserSubscription{GroupID: 1})
			setCustomGatewayGroup(context, key)
			actual, ok := middleware2.GetAPIKeyFromContext(context)
			require.True(t, ok)
			require.Same(t, key, actual)
			subscription, exists := context.Get(string(middleware2.ContextKeySubscription))
			require.True(t, exists)
			require.Nil(t, subscription)
			require.Same(t, group, context.Request.Context().Value(ctxkey.Group))
		})
	}
}
