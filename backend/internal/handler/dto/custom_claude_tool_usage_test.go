package dto

import (
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestCustomClaudeToolDTO(t *testing.T) {
	log := &service.UsageLog{ToolUsage: map[string]service.ToolCallCost{"web_search": {Calls: 1, UnitPrice: 0.01, TotalCost: 0.01, ActualCost: 0.0015}}}
	for _, mapped := range []any{UsageLogFromService(log), UsageLogFromServiceAdmin(log)} {
		body, err := json.Marshal(mapped)
		require.NoError(t, err)
		var decoded map[string]any
		require.NoError(t, json.Unmarshal(body, &decoded))
		tool := decoded["tool_usage"].(map[string]any)["web_search"].(map[string]any)
		require.Equal(t, float64(1), tool["calls"])
		require.Equal(t, 0.0015, tool["actual_cost"])
	}
	body, err := json.Marshal(UsageLogFromService(&service.UsageLog{}))
	require.NoError(t, err)
	require.NotContains(t, string(body), "tool_usage")
}
