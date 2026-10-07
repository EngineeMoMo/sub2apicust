//go:build unit

package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// [CUSTOM] 同时覆盖公开 JSON 和 HTML 注入源，防止只隐藏徽标仍泄露版本。
func TestCustomPublicVersionHidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const privateVersion = "0.2.13-custom.private-build"
	svc := service.NewSettingService(&settingHandlerPublicRepoStub{values: map[string]string{
		service.SettingKeySiteName: "魔法家族",
	}}, &config.Config{})
	svc.SetVersion(privateVersion)
	h := NewSettingHandler(svc, privateVersion)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/settings/public", nil)
	h.GetPublicSettings(c)
	require.Equal(t, http.StatusOK, recorder.Code)
	var response struct {
		Data map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, "", response.Data["version"])
	require.Equal(t, "魔法家族", response.Data["site_name"])
	require.NotContains(t, recorder.Body.String(), privateVersion)

	payload, err := svc.GetPublicSettingsForInjection(context.Background())
	require.NoError(t, err)
	encoded, err := json.Marshal(payload)
	require.NoError(t, err)
	var injection map[string]any
	require.NoError(t, json.Unmarshal(encoded, &injection))
	require.Equal(t, "", injection["version"])
	require.Equal(t, "魔法家族", injection["site_name"])
	require.NotContains(t, string(encoded), privateVersion)
}
