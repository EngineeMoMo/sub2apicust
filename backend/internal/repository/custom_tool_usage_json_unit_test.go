//go:build unit

package repository

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestCustomClaudeToolUsageJSONAndInsert(t *testing.T) {
	tools := map[string]service.ToolCallCost{"web_search": {Calls: 2, UnitPrice: 0.01, TotalCost: 0.02, ActualCost: 0.003}}
	jsonValue := nullToolUsageJSON(tools).(string)
	require.Equal(t, tools, toolUsageFromNullJSON(sql.NullString{Valid: true, String: jsonValue}))
	require.Nil(t, nullToolUsageJSON(nil))
	require.Nil(t, toolUsageFromNullJSON(sql.NullString{}))
	require.Nil(t, toolUsageFromNullJSON(sql.NullString{Valid: true, String: "bad json"}))
	prepared := prepareUsageLogInsert(&service.UsageLog{ToolUsage: tools})
	columns := strings.Split(usageLogSelectColumns, ",")
	for index, column := range columns {
		if strings.TrimSpace(column) == "tool_usage" {
			require.Equal(t, "jsonb", usageLogInsertArgTypes[index-1])
			require.JSONEq(t, jsonValue, prepared.args[index-1].(string))
			return
		}
	}
	t.Fatal("tool_usage missing in log query")
}
