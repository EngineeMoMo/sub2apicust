//go:build unit

package repository

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestCustomClaudeToolRealPostgres(t *testing.T) {
	dsn := os.Getenv("CLAUDE_TOOL_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()
	require.NoError(t, ApplyMigrations(ctx, db))
	require.NoError(t, ApplyMigrations(ctx, db))
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	user := client.User.Create().SetEmail("claude-tool@test.invalid").SetPasswordHash("test").SetBalance(1).SaveX(ctx)
	group := client.Group.Create().SetName("tool-test").SetPlatform("anthropic").SetSubscriptionType("subscription").SaveX(ctx)
	key := client.APIKey.Create().SetUserID(user.ID).SetKey("sk-tool-test").SetName("test").SetGroupID(group.ID).SaveX(ctx)
	account := client.Account.Create().SetName("test").SetPlatform("anthropic").SetType("api_key").SetCredentials(map[string]any{}).SaveX(ctx)
	subscription := client.UserSubscription.Create().SetUserID(user.ID).SetGroupID(group.ID).SetStartsAt(time.Now()).SetExpiresAt(time.Now().Add(time.Hour)).SaveX(ctx)
	logs := newUsageLogRepositoryWithSQL(client, db)
	tools := map[string]service.ToolCallCost{"web_search": {Calls: 1, UnitPrice: 0.01, TotalCost: 0.01, ActualCost: 0.0015}}
	makeLog := func(id string) *service.UsageLog {
		return &service.UsageLog{UserID: user.ID, APIKeyID: key.ID, AccountID: account.ID, RequestID: id, Model: "claude-sonnet-4", InputTokens: 132, OutputTokens: 354, TotalCost: 0.015706, ActualCost: 0.0023559, ToolUsage: tools}
	}
	direct := makeLog("direct")
	inserted, err := logs.Create(ctx, direct)
	require.NoError(t, err)
	require.True(t, inserted)
	saved, err := logs.GetByID(ctx, direct.ID)
	require.NoError(t, err)
	require.Equal(t, tools, saved.ToolUsage)
	inserted, err = logs.Create(ctx, direct)
	require.NoError(t, err)
	require.False(t, inserted)
	var wait sync.WaitGroup
	for _, id := range []string{"batch-a", "batch-b", "batch-c"} {
		wait.Add(1)
		go func(id string) { defer wait.Done(); _, err := logs.Create(ctx, makeLog(id)); require.NoError(t, err) }(id)
	}
	wait.Wait()
	require.NoError(t, logs.CreateBestEffort(ctx, makeLog("best-effort")))
	var count int
	require.NoError(t, db.QueryRow("SELECT count(*) FROM usage_logs WHERE tool_usage->'web_search'->>'calls'='1'").Scan(&count))
	require.Equal(t, 5, count)
	historical := makeLog("historical")
	historical.ToolUsage = nil
	_, err = logs.Create(ctx, historical)
	require.NoError(t, err)
	saved, err = logs.GetByID(ctx, historical.ID)
	require.NoError(t, err)
	require.Nil(t, saved.ToolUsage)
	billing := NewUsageBillingRepository(nil, db)
	for _, isSubscription := range []bool{false, true} {
		cmd := &service.UsageBillingCommand{RequestID: "bill-balance", APIKeyID: key.ID, UserID: user.ID, AccountID: account.ID, AccountType: "api_key", Model: "claude-sonnet-4", APIKeyQuotaCost: 0.0023559}
		if isSubscription {
			cmd.RequestID = "bill-subscription"
			cmd.SubscriptionID = &subscription.ID
			cmd.SubscriptionCost = 0.0023559
		} else {
			cmd.BalanceCost = 0.0023559
		}
		result, err := billing.Apply(ctx, cmd)
		require.NoError(t, err)
		require.True(t, result.Applied)
		result, err = billing.Apply(ctx, cmd)
		require.NoError(t, err)
		require.False(t, result.Applied)
	}
	var balance, quota, daily, weekly, monthly float64
	require.NoError(t, db.QueryRow("SELECT balance FROM users WHERE id=$1", user.ID).Scan(&balance))
	require.NoError(t, db.QueryRow("SELECT quota_used FROM api_keys WHERE id=$1", key.ID).Scan(&quota))
	require.NoError(t, db.QueryRow("SELECT daily_usage_usd,weekly_usage_usd,monthly_usage_usd FROM user_subscriptions WHERE id=$1", subscription.ID).Scan(&daily, &weekly, &monthly))
	require.InDelta(t, 1-0.0023559, balance, 1e-10)
	require.InDelta(t, 2*0.0023559, quota, 1e-10)
	for _, value := range []float64{daily, weekly, monthly} {
		require.InDelta(t, 0.0023559, value, 1e-10)
	}
}
