//go:build unit

package admin

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestRegistrationIPSettingsSaveAndPreserve(t *testing.T) {
	for _, stored := range []string{"true", "false"} {
		t.Run(stored, func(t *testing.T) {
			handler, repo := newStepUpSwitchTestHandler(t, map[string]string{
				service.SettingKeyRegistrationIPLimitEnabled: stored,
			})
			response := doUpdateSettings(t, handler, map[string]any{"site_name": "魔法家族"}, nil)
			require.Equal(t, http.StatusOK, response.Code)
			require.Equal(t, stored, repo.values[service.SettingKeyRegistrationIPLimitEnabled])
			response = doUpdateSettings(t, handler, map[string]any{"registration_ip_limit_enabled": stored != "true"}, nil)
			require.Equal(t, http.StatusOK, response.Code)
			require.Equal(t, map[string]string{"true": "false", "false": "true"}[stored], repo.values[service.SettingKeyRegistrationIPLimitEnabled])
			var payload struct {
				Data struct {
					Enabled bool `json:"registration_ip_limit_enabled"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
			require.Equal(t, stored != "true", payload.Data.Enabled)
		})
	}
}
