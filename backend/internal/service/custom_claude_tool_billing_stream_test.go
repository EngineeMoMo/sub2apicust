//go:build unit

package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const customClaudeSearchStart = `{"type":"message_start","message":{"id":"msg_search","type":"message","role":"assistant","model":"claude-sonnet-4","content":[],"usage":{"input_tokens":127,"server_tool_use":{"web_search_requests":0}}}}`
const customClaudeSearchResult = `{"type":"content_block_start","index":0,"content_block":{"type":"web_search_tool_result","tool_use_id":"search_a","content":[]}}`
const customClaudeSearchEnd = `{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":529}}`

func TestCustomClaudeToolInitialZeroStream(t *testing.T) {
	for _, test := range []struct {
		name   string
		events []string
		want   int
	}{
		{"successful_result", []string{customClaudeSearchResult, customClaudeSearchEnd}, 1},
		{"interrupted_after_result", []string{customClaudeSearchResult}, 1},
		{"duplicate_result", []string{customClaudeSearchResult, customClaudeSearchResult, customClaudeSearchEnd}, 1},
		{"distinct_results", []string{customClaudeSearchResult, strings.ReplaceAll(customClaudeSearchResult, "search_a", "search_b"), customClaudeSearchEnd}, 2},
		{"no_result", []string{customClaudeSearchEnd}, 0},
		{"failed_result", []string{`{"type":"content_block_start","content_block":{"type":"web_search_tool_result","tool_use_id":"search_a","content":{"type":"web_search_tool_result_error","error_code":"unavailable"}}}`, customClaudeSearchEnd}, 0},
		{"server_start_only", []string{`{"type":"content_block_start","content_block":{"type":"server_tool_use","name":"web_search","id":"search_a"}}`, customClaudeSearchEnd}, 0},
		{"client_tool_only", []string{`{"type":"content_block_start","content_block":{"type":"tool_use","name":"web_search","id":"search_a"}}`, customClaudeSearchEnd}, 0},
		{"explicit_final_zero_with_success", []string{customClaudeSearchResult, `{"type":"message_delta","usage":{"server_tool_use":{"web_search_requests":0}}}`}, 1},
		{"explicit_final_count", []string{customClaudeSearchResult, `{"type":"message_delta","usage":{"server_tool_use":{"web_search_requests":2}}}`}, 2},
		{"invalid_final_count", []string{customClaudeSearchResult, `{"type":"message_delta","usage":{"server_tool_use":{"web_search_requests":-1}}}`}, 1},
		{"string_final_count", []string{customClaudeSearchResult, `{"type":"message_delta","usage":{"server_tool_use":{"web_search_requests":"2"}}}`}, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			for name, parse := range map[string]func(string, *ClaudeUsage){"passthrough": parseSSEUsagePassthrough, "generic": (&GatewayService{}).parseSSEUsage} {
				t.Run(name, func(t *testing.T) {
					usage := &ClaudeUsage{}
					parse(customClaudeSearchStart, usage)
					for _, event := range test.events {
						parse(event, usage)
					}
					require.Equal(t, test.want, usage.WebSearchRequests)
					require.Equal(t, test.want, claudeUsageToOpenAIUsage(usage).ClaudeWebSearchRequests)
				})
			}
		})
	}
}

func TestCustomClaudeToolInitialZeroResponsesConversion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var stream strings.Builder
	for _, event := range []string{customClaudeSearchStart, customClaudeSearchResult, customClaudeSearchEnd, `{"type":"message_stop"}`} {
		fmt.Fprintf(&stream, "event: %s\ndata: %s\n\n", gjson.Get(event, "type").String(), event)
	}
	for _, streaming := range []bool{false, true} {
		t.Run(fmt.Sprint(streaming), func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			response := &http.Response{Header: make(http.Header), Body: io.NopCloser(strings.NewReader(stream.String()))}
			gateway := &GatewayService{}
			var result *ForwardResult
			var err error
			if streaming {
				result, err = gateway.handleResponsesStreamingResponse(response, ctx, "claude-sonnet-4", "claude-sonnet-4", nil, time.Now(), apicompat.ResponsesClientToolMapping{})
			} else {
				result, err = gateway.handleResponsesBufferedStreamingResponse(response, ctx, "claude-sonnet-4", "claude-sonnet-4", nil, time.Now(), apicompat.ResponsesClientToolMapping{})
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, 1, result.Usage.WebSearchRequests)
			require.Equal(t, 127, result.Usage.InputTokens)
			require.Equal(t, 529, result.Usage.OutputTokens)
		})
	}
}

func TestCustomClaudeToolInitialZeroSettlement(t *testing.T) {
	for _, subscription := range []bool{false, true} {
		t.Run(fmt.Sprint(subscription), func(t *testing.T) {
			logs := &openAIRecordUsageLogRepoStub{}
			repo := &openAIRecordUsageBillingRepoStub{result: &UsageBillingApplyResult{Applied: true}}
			gateway := newGatewayRecordUsageServiceWithBillingRepoForTest(logs, repo, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{})
			group := &Group{ID: 1, Platform: PlatformAnthropic, RateMultiplier: 0.15, WebSearchPricePerCall: floatPtr(0.05)}
			result := &ForwardResult{RequestID: "initial-zero-search", Model: "claude-sonnet-4"}
			for _, event := range []string{customClaudeSearchStart, customClaudeSearchResult, customClaudeSearchEnd} {
				parseSSEUsagePassthrough(event, &result.Usage)
			}
			input := &RecordUsageInput{Result: result, APIKey: &APIKey{ID: 2, GroupID: &group.ID, Group: group, Quota: 100, RateLimit5h: 100}, User: &User{ID: 3}, Account: &Account{ID: 4, Type: AccountTypeAPIKey}, APIKeyService: &apiKeyQuotaUpdaterStub{}}
			if subscription {
				group.SubscriptionType = "subscription"
				input.Subscription = &UserSubscription{ID: 5}
			}
			require.NoError(t, gateway.RecordUsage(context.Background(), input))
			require.NotNil(t, logs.lastLog)
			require.NotNil(t, repo.lastCmd)
			require.InDelta(t, 0.058316, logs.lastLog.TotalCost, 1e-10)
			require.InDelta(t, 0.0087474, logs.lastLog.ActualCost, 1e-10)
			require.Equal(t, 1, logs.lastLog.ToolUsage["web_search"].Calls)
			require.InDelta(t, 0.05, logs.lastLog.ToolUsage["web_search"].UnitPrice, 1e-10)
			require.InDelta(t, 0.0075, logs.lastLog.ToolUsage["web_search"].ActualCost, 1e-10)
			if subscription {
				require.InDelta(t, 0.0087474, repo.lastCmd.SubscriptionCost, 1e-10)
				require.Zero(t, repo.lastCmd.BalanceCost)
			} else {
				require.InDelta(t, 0.0087474, repo.lastCmd.BalanceCost, 1e-10)
			}
			require.InDelta(t, 0.0087474, repo.lastCmd.APIKeyQuotaCost, 1e-10)
			require.InDelta(t, 0.0087474, repo.lastCmd.APIKeyRateLimitCost, 1e-10)
		})
	}
}
