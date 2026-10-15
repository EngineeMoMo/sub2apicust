//go:build unit

package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestCustomClaudeToolResponseUsage(t *testing.T) {
	for _, test := range []struct {
		name, body string
		want       int
	}{
		{"reported", `{"usage":{"server_tool_use":{"web_search_requests":2}}}`, 2},
		{"negative", `{"usage":{"server_tool_use":{"web_search_requests":-1}}}`, 0},
		{"fraction", `{"usage":{"server_tool_use":{"web_search_requests":1.5}}}`, 0},
		{"string", `{"usage":{"server_tool_use":{"web_search_requests":"2"}}}`, 0},
		{"overflow", `{"usage":{"server_tool_use":{"web_search_requests":2147483648}}}`, 0},
		{"client_tool", `{"content":[{"type":"tool_use","name":"web_search","id":"a"}]}`, 0},
		{"server_start", `{"content":[{"type":"server_tool_use","name":"web_search","id":"a"}]}`, 0},
		{"successful_fallback", `{"content":[{"type":"web_search_tool_result","tool_use_id":"a","content":[]},{"type":"web_search_tool_result","tool_use_id":"a","content":[]}]}`, 1},
		{"error_free", `{"content":[{"type":"web_search_tool_result","tool_use_id":"a","content":{"type":"web_search_tool_result_error","error_code":"unavailable"}}]}`, 0},
		{"reported_zero_with_success", `{"usage":{"server_tool_use":{"web_search_requests":0}},"content":[{"type":"web_search_tool_result","tool_use_id":"a","content":[]}]}`, 1},
		{"no_usage", `{"usage":{"input_tokens":132,"output_tokens":354}}`, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			usage := parseClaudeUsageFromResponseBody([]byte(test.body))
			require.Equal(t, test.want, usage.WebSearchRequests)
			require.Equal(t, test.want, claudeUsageToOpenAIUsage(usage).ClaudeWebSearchRequests)
		})
	}
}

func TestCustomClaudeToolStreamCumulative(t *testing.T) {
	for name, parse := range map[string]func(string, *ClaudeUsage){"passthrough": parseSSEUsagePassthrough, "generic": (&GatewayService{}).parseSSEUsage} {
		t.Run(name, func(t *testing.T) {
			usage := &ClaudeUsage{}
			parse(`{"type":"message_start","message":{"usage":{"input_tokens":132,"server_tool_use":{"web_search_requests":0}}}}`, usage)
			for _, count := range []int{1, 1, 3, 3} {
				parse(fmt.Sprintf(`{"type":"message_delta","usage":{"output_tokens":354,"server_tool_use":{"web_search_requests":%d}}}`, count), usage)
			}
			require.Equal(t, 3, usage.WebSearchRequests)
			require.Equal(t, 132, usage.InputTokens)
			require.Equal(t, 354, usage.OutputTokens)
			require.True(t, usage.hasObservedTokens())
		})
	}
}

func TestCustomClaudeToolStreamFallbackAndReports(t *testing.T) {
	usage := &ClaudeUsage{}
	successful := `{"type":"content_block_start","content_block":{"type":"web_search_tool_result","tool_use_id":"a","content":[]}}`
	observeClaudeToolEvent(successful, usage)
	observeClaudeToolEvent(successful, usage)
	require.Equal(t, 1, usage.WebSearchRequests)
	observeClaudeToolEvent(`{"type":"message_delta","usage":{"server_tool_use":{"web_search_requests":0}}}`, usage)
	require.Equal(t, 1, usage.WebSearchRequests)
	require.True(t, (&ClaudeUsage{WebSearchRequests: 1}).hasObservedTokens())
}

func TestCustomClaudeToolBothGatewayPrices(t *testing.T) {
	billing := NewBillingService(&config.Config{}, nil)
	for _, test := range []struct {
		name              string
		price             *float64
		multiplier, extra float64
	}{
		{"default", nil, 0.15, 0.01},
		{"override", floatPtr(0.02), 0.15, 0.02},
		{"free", floatPtr(0), 0.15, 0},
		{"effective_peak_rate", nil, 0.3, 0.01},
		{"zero_multiplier", nil, 0, 0.01},
	} {
		t.Run(test.name, func(t *testing.T) {
			group := &Group{ID: 1, Platform: PlatformAnthropic, WebSearchPricePerCall: test.price}
			key := &APIKey{ID: 2, Group: group, GroupID: &group.ID}
			result := &ForwardResult{Model: "claude-sonnet-4", Usage: ClaudeUsage{InputTokens: 132, OutputTokens: 354, WebSearchRequests: 1}}
			gw := &GatewayService{billingService: billing}
			cost := gw.calculateRecordUsageCost(context.Background(), result, key, result.Model, test.multiplier, test.multiplier, time.Now())
			want := 0.005706 + test.extra
			require.InDelta(t, want, cost.TotalCost, 1e-10)
			require.InDelta(t, want*test.multiplier, cost.ActualCost, 1e-10)
			require.Equal(t, 1, cost.ToolUsage["web_search"].Calls)
			require.InDelta(t, test.extra, cost.ToolUsage["web_search"].UnitPrice, 1e-10)
			openAI := &OpenAIGatewayService{billingService: billing}
			native := &OpenAIForwardResult{Model: result.Model, Usage: claudeUsageToOpenAIUsage(&result.Usage)}
			nativeCost, err := openAI.calculateOpenAIRecordUsageCost(context.Background(), native, key, []string{result.Model}, test.multiplier, test.multiplier, test.multiplier, test.multiplier, UsageTokens{InputTokens: 132, OutputTokens: 354}, "", nil, false, time.Now())
			require.NoError(t, err)
			require.InDelta(t, cost.ActualCost, nativeCost.ActualCost, 1e-10)
			require.Equal(t, cost.ToolUsage, nativeCost.ToolUsage)
		})
	}
}

func TestCustomClaudeToolRecordUsageBalanceAndSubscription(t *testing.T) {
	for _, subscription := range []bool{false, true} {
		t.Run(fmt.Sprint(subscription), func(t *testing.T) {
			logs := &openAIRecordUsageLogRepoStub{}
			repo := &openAIRecordUsageBillingRepoStub{result: &UsageBillingApplyResult{Applied: true}}
			svc := newGatewayRecordUsageServiceWithBillingRepoForTest(logs, repo, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{})
			group := &Group{ID: 1, Platform: PlatformAnthropic, RateMultiplier: 0.15}
			input := &RecordUsageInput{Result: &ForwardResult{RequestID: "tool-bill", Model: "claude-sonnet-4", Usage: ClaudeUsage{InputTokens: 132, OutputTokens: 354, WebSearchRequests: 1}}, APIKey: &APIKey{ID: 2, GroupID: &group.ID, Group: group, Quota: 100}, User: &User{ID: 3}, Account: &Account{ID: 4, Type: AccountTypeAPIKey}, APIKeyService: &apiKeyQuotaUpdaterStub{}}
			if subscription {
				group.SubscriptionType = "subscription"
				input.Subscription = &UserSubscription{ID: 5}
			}
			require.NoError(t, svc.RecordUsage(context.Background(), input))
			require.InDelta(t, 0.0023559, logs.lastLog.ActualCost, 1e-10)
			require.InDelta(t, 0.0015, logs.lastLog.ToolUsage["web_search"].ActualCost, 1e-10)
			if subscription {
				require.InDelta(t, 0.0023559, repo.lastCmd.SubscriptionCost, 1e-10)
				require.Zero(t, repo.lastCmd.BalanceCost)
			} else {
				require.InDelta(t, 0.0023559, repo.lastCmd.BalanceCost, 1e-10)
			}
			require.InDelta(t, 0.0023559, repo.lastCmd.APIKeyQuotaCost, 1e-10)
		})
	}
}

func TestCustomClaudeToolStatsAndFailureSnapshot(t *testing.T) {
	billing := NewBillingService(&config.Config{}, nil)
	log := &UsageLog{ToolUsage: map[string]ToolCallCost{"web_search": {Calls: 1, UnitPrice: 0.01, TotalCost: 0.01, ActualCost: 0.0015}}}
	applyAccountStatsCost(context.Background(), log, nil, billing, 1, 2, "claude-sonnet-4", "", UsageTokens{InputTokens: 132, OutputTokens: 354}, 0.015706, time.Now(), true)
	require.Nil(t, log.AccountStatsCost)
	for _, applyCustomerPricing := range []bool{false, true} {
		channel := &Channel{ID: 1, Status: StatusActive, ApplyPricingToAccountStats: applyCustomerPricing}
		cs := newTestChannelServiceForStats(t, channel, 2, "anthropic")
		applyAccountStatsCost(context.Background(), log, cs, billing, 1, 2, "claude-sonnet-4", "", UsageTokens{InputTokens: 132, OutputTokens: 354}, 0.015706, time.Now(), true)
		require.NotNil(t, log.AccountStatsCost)
		require.InDelta(t, 0.015706, *log.AccountStatsCost, 1e-10)
	}
	clearToolUsageActualCost(log)
	require.Zero(t, log.ToolUsage["web_search"].ActualCost)
	require.Equal(t, 0.01, log.ToolUsage["web_search"].TotalCost)
}
