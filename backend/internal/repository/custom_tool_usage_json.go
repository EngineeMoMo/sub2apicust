package repository

import (
	"database/sql"
	"encoding/json"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func nullToolUsageJSON(tools map[string]service.ToolCallCost) any {
	if len(tools) == 0 {
		return nil
	}
	body, err := json.Marshal(tools)
	if err != nil {
		return nil
	}
	return string(body)
}

func toolUsageFromNullJSON(value sql.NullString) map[string]service.ToolCallCost {
	if !value.Valid {
		return nil
	}
	var tools map[string]service.ToolCallCost
	if json.Unmarshal([]byte(value.String), &tools) != nil {
		return nil
	}
	return tools
}
