package service

import (
	"math"
	"strconv"

	"github.com/tidwall/gjson"
)

// ToolCallCost 保存实际调用量及本次定价快照，金额包含在 TotalCost/ActualCost 中。
type ToolCallCost struct {
	Calls      int     `json:"calls"`
	UnitPrice  float64 `json:"unit_price"`
	TotalCost  float64 `json:"total_cost"`
	ActualCost float64 `json:"actual_cost"`
}

func (u *ClaudeUsage) observeWebSearchCount(node gjson.Result) {
	value := node.Get("server_tool_use.web_search_requests")
	if value.Type != gjson.Number {
		return
	}
	count, err := strconv.ParseInt(value.Raw, 10, 32)
	if err != nil || count < 0 {
		return
	}
	if !u.webSearchCountObserved {
		u.WebSearchRequests = int(count)
		u.webSearchCountObserved = true
	} else if int(count) > u.WebSearchRequests {
		u.WebSearchRequests = int(count)
	}
}

func (u *ClaudeUsage) observeWebSearchResult(block gjson.Result) {
	if block.Get("type").String() != "web_search_tool_result" || !block.Get("content").IsArray() {
		return
	}
	id := block.Get("tool_use_id").String()
	if id == "" {
		return
	}
	if u.webSearchResultIDs == nil {
		u.webSearchResultIDs = make(map[string]struct{})
	}
	u.webSearchResultIDs[id] = struct{}{}
	if !u.webSearchCountObserved {
		u.WebSearchRequests = len(u.webSearchResultIDs)
	}
}

func observeClaudeToolResponse(body []byte, usage *ClaudeUsage) {
	if usage == nil {
		return
	}
	parsed := gjson.ParseBytes(body)
	usage.observeWebSearchCount(parsed.Get("usage"))
	parsed.Get("content").ForEach(func(_, block gjson.Result) bool {
		usage.observeWebSearchResult(block)
		return true
	})
}

func observeClaudeToolEvent(data string, usage *ClaudeUsage) {
	if usage == nil {
		return
	}
	parsed := gjson.Parse(data)
	switch parsed.Get("type").String() {
	case "message_start":
		usage.observeWebSearchCount(parsed.Get("message.usage"))
	case "message_delta":
		usage.observeWebSearchCount(parsed.Get("usage"))
	case "content_block_start":
		usage.observeWebSearchResult(parsed.Get("content_block"))
	}
}

func addClaudeWebSearchCost(cost *CostBreakdown, calls int, price *float64, multiplier float64) *CostBreakdown {
	if calls <= 0 {
		return cost
	}
	unitPrice := defaultWebSearchPricePerCall
	if price != nil {
		unitPrice = *price
	}
	if unitPrice < 0 || math.IsNaN(unitPrice) || math.IsInf(unitPrice, 0) {
		unitPrice = defaultWebSearchPricePerCall
	}
	if cost == nil {
		cost = &CostBreakdown{}
	}
	total := float64(calls) * unitPrice
	if multiplier < 0 {
		multiplier = 0
	}
	actual := total * multiplier
	cost.TotalCost += total
	cost.ActualCost += actual
	cost.ToolUsage = map[string]ToolCallCost{"web_search": {Calls: calls, UnitPrice: unitPrice, TotalCost: total, ActualCost: actual}}
	return cost
}

func toolUsageTotalCost(tools map[string]ToolCallCost) float64 {
	total := 0.0
	for _, tool := range tools {
		total += tool.TotalCost
	}
	return total
}

func clearToolUsageActualCost(log *UsageLog) {
	for name, tool := range log.ToolUsage {
		tool.ActualCost = 0
		log.ToolUsage[name] = tool
	}
}
