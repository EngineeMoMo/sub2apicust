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

const customClaudeSearchFinalZero = `{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":587,"server_tool_use":{"web_search_requests":0}}}`

func TestCustomClaudeToolSuccessfulResultFloor(t *testing.T) {
	secondResult := strings.ReplaceAll(customClaudeSearchResult, "search_a", "search_b")
	for _, test := range []struct {
		name   string
		events []string
		want   int
	}{
		{"final_zero_after_success", []string{customClaudeSearchResult, customClaudeSearchFinalZero}, 1},
		{"zero_before_success", []string{customClaudeSearchFinalZero, customClaudeSearchResult}, 1},
		{"duplicate_after_final_zero", []string{customClaudeSearchResult, customClaudeSearchFinalZero, customClaudeSearchResult}, 1},
		{"distinct_after_final_zero", []string{customClaudeSearchResult, customClaudeSearchFinalZero, secondResult}, 2},
		{"positive_report_below_results", []string{customClaudeSearchResult, secondResult, `{"type":"message_delta","usage":{"server_tool_use":{"web_search_requests":1}}}`}, 2},
		{"results_after_positive_report", []string{`{"type":"message_delta","usage":{"server_tool_use":{"web_search_requests":1}}}`, customClaudeSearchResult, secondResult}, 2},
		{"positive_report_above_results", []string{customClaudeSearchResult, `{"type":"message_delta","usage":{"server_tool_use":{"web_search_requests":3}}}`, secondResult}, 3},
		{"zero_without_success", []string{customClaudeSearchFinalZero}, 0},
		{"zero_with_failed_result", []string{`{"type":"content_block_start","content_block":{"type":"web_search_tool_result","tool_use_id":"search_a","content":{"type":"web_search_tool_result_error","error_code":"unavailable"}}}`, customClaudeSearchFinalZero}, 0},
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
					require.Equal(t, 127, usage.InputTokens)
				})
			}
		})
	}
}

func TestCustomClaudeToolSuccessfulResultJSONFloor(t *testing.T) {
	for _, test := range []struct {
		name, body string
		want       int
	}{
		{"zero_with_success", `{"usage":{"input_tokens":127,"output_tokens":587,"server_tool_use":{"web_search_requests":0}},"content":[{"type":"web_search_tool_result","tool_use_id":"search_a","content":[]}]}`, 1},
		{"report_below_distinct_results", `{"usage":{"input_tokens":127,"output_tokens":587,"server_tool_use":{"web_search_requests":1}},"content":[{"type":"web_search_tool_result","tool_use_id":"search_a","content":[]},{"type":"web_search_tool_result","tool_use_id":"search_b","content":[]},{"type":"web_search_tool_result","tool_use_id":"search_a","content":[]}]}`, 2},
		{"zero_without_success", `{"usage":{"input_tokens":127,"output_tokens":587,"server_tool_use":{"web_search_requests":0}},"content":[{"type":"server_tool_use","name":"web_search","id":"search_a"}]}`, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			usage := parseClaudeUsageFromResponseBody([]byte(test.body))
			require.Equal(t, test.want, usage.WebSearchRequests)
			require.Equal(t, 127, usage.InputTokens)
			require.Equal(t, 587, usage.OutputTokens)
		})
	}
}

func TestCustomClaudeToolSuccessfulResultResponsesConversion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var stream strings.Builder
	for _, event := range []string{customClaudeSearchStart, customClaudeSearchResult, customClaudeSearchFinalZero, `{"type":"message_stop"}`} {
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
			require.Equal(t, 587, result.Usage.OutputTokens)
		})
	}
}

func TestCustomClaudeToolSuccessfulResultSettlement(t *testing.T) {
	for _, subscription := range []bool{false, true} {
		t.Run(fmt.Sprint(subscription), func(t *testing.T) {
			logs := &openAIRecordUsageLogRepoStub{}
			repo := &openAIRecordUsageBillingRepoStub{result: &UsageBillingApplyResult{Applied: true}}
			gateway := newGatewayRecordUsageServiceWithBillingRepoForTest(logs, repo, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{})
			group := &Group{ID: 1, Platform: PlatformAnthropic, RateMultiplier: 0.15, WebSearchPricePerCall: floatPtr(0.05)}
			result := &ForwardResult{RequestID: "successful-result-final-zero", Model: "claude-sonnet-4"}
			for _, event := range []string{customClaudeSearchStart, customClaudeSearchResult, customClaudeSearchFinalZero} {
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
			require.InDelta(t, 0.059186, logs.lastLog.TotalCost, 1e-10)
			require.InDelta(t, 0.0088779, logs.lastLog.ActualCost, 1e-10)
			require.Equal(t, 1, logs.lastLog.ToolUsage["web_search"].Calls)
			require.InDelta(t, 0.05, logs.lastLog.ToolUsage["web_search"].UnitPrice, 1e-10)
			require.InDelta(t, 0.0075, logs.lastLog.ToolUsage["web_search"].ActualCost, 1e-10)
			if subscription {
				require.InDelta(t, 0.0088779, repo.lastCmd.SubscriptionCost, 1e-10)
				require.Zero(t, repo.lastCmd.BalanceCost)
			} else {
				require.InDelta(t, 0.0088779, repo.lastCmd.BalanceCost, 1e-10)
			}
			require.InDelta(t, 0.0088779, repo.lastCmd.APIKeyQuotaCost, 1e-10)
			require.InDelta(t, 0.0088779, repo.lastCmd.APIKeyRateLimitCost, 1e-10)
		})
	}
}
