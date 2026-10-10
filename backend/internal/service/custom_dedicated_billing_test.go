//go:build unit

package service

import (
	"context"
	"errors"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"math"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func customDedicatedBillingTestKey() *APIKey {
	group := int64(33)
	return &APIKey{ID: 7, UserID: 11, GroupID: &group, Quota: 100, RateLimit5h: 10, User: &User{ID: 11}, customDedicatedBilling: &customDedicatedBillingGrant{BindingID: 1, AccountID: 22, UserID: 11, APIKeyID: 7, GroupID: 33, LeaseID: "admitted"}}
}

func TestCustomDedicatedAdmissionCreatesReceiptWithoutPolicyOrCounters(t *testing.T) {
	for _, websocket := range []bool{false, true} {
		database, mock, err := sqlmock.New()
		require.NoError(t, err)
		key := customDedicatedBillingTestKey()
		key.customDedicatedBilling.LeaseID = ""
		acceptedAt := time.Now()
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT .* FROM custom_dedicated_accounts WHERE id=.*FOR UPDATE").WithArgs(int64(1)).WillReturnRows(customDedicatedTestRows(customDedicatedTestBinding(acceptedAt)))
		mock.ExpectQuery("SELECT clock_timestamp").WillReturnRows(sqlmock.NewRows([]string{"accepted_at"}).AddRow(acceptedAt))
		mock.ExpectQuery(regexp.QuoteMeta(customDedicatedAccessSQL)).WithArgs(int64(22), int64(33), "[11]", int64(11)).WillReturnRows(sqlmock.NewRows([]string{"valid"}).AddRow(true))
		mock.ExpectExec("INSERT INTO custom_dedicated_request_leases.*finished_at").WithArgs(sqlmock.AnyArg(), int64(1), int64(22), int64(11), int64(7), int64(33), acceptedAt).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
		svc := &CustomDedicatedService{db: database}
		require.NoError(t, svc.AdmitBillingRequest(context.Background(), key, websocket))
		require.Len(t, key.customDedicatedBilling.LeaseID, 48)
		require.Equal(t, websocket, key.customDedicatedBilling.WebSocket)
		require.NoError(t, mock.ExpectationsWereMet())
		mock.ExpectClose()
		require.NoError(t, database.Close())
	}
}

func TestCustomDedicatedBillingSeparateChargeAndMeter(t *testing.T) {
	for _, actual := range []float64{0, 9} {
		p := &postUsageBillingParams{APIKey: customDedicatedBillingTestKey(), User: &User{ID: 11, Balance: 0}, Account: &Account{ID: 22, Type: AccountTypeOAuth}, Cost: &CostBreakdown{TotalCost: 2, ActualCost: actual}, APIKeyService: apiKeyQuotaUpdaterStub{}}
		log := &UsageLog{ActualCost: actual, TotalCost: 2, BillingType: BillingTypeBalance}
		require.NoError(t, prepareCustomDedicatedSettlement(p, log))
		cmd := buildUsageBillingCommand("accepted", log, p)
		require.Zero(t, cmd.BalanceCost)
		require.Zero(t, cmd.SubscriptionCost)
		require.Equal(t, actual, cmd.APIKeyQuotaCost)
		require.Equal(t, actual, cmd.APIKeyRateLimitCost)
		require.Equal(t, float64(2), cmd.DedicatedReferenceCost)
		require.Equal(t, int8(2), log.BillingType)
		require.Zero(t, log.ActualCost)
		require.Equal(t, float64(2), log.TotalCost)
	}
	ordinary := customDedicatedBillingTestKey()
	ordinary.customDedicatedBilling = nil
	p := &postUsageBillingParams{APIKey: ordinary, User: &User{ID: 11}, Account: &Account{ID: 22}, Cost: &CostBreakdown{TotalCost: 2, ActualCost: 9}, APIKeyService: apiKeyQuotaUpdaterStub{}}
	require.Equal(t, float64(9), buildUsageBillingCommand("ordinary", nil, p).BalanceCost)
}

func TestCustomDedicatedBillingRejectWrongOrMissingAdmission(t *testing.T) {
	for _, mutation := range []func(*postUsageBillingParams){
		func(p *postUsageBillingParams) { p.Account.ID = 99 },
		func(p *postUsageBillingParams) { p.User.ID = 12 },
		func(p *postUsageBillingParams) { p.APIKey.customDedicatedBilling.LeaseID = "" },
		func(p *postUsageBillingParams) { p.APIKey.ID = 8 },
		func(p *postUsageBillingParams) { group := int64(99); p.APIKey.GroupID = &group },
		func(p *postUsageBillingParams) { p.IsSubscriptionBill = true },
		func(p *postUsageBillingParams) { p.Cost.TotalCost = math.NaN() },
	} {
		p := &postUsageBillingParams{APIKey: customDedicatedBillingTestKey(), User: &User{ID: 11}, Account: &Account{ID: 22}, Cost: &CostBreakdown{TotalCost: 2}}
		mutation(p)
		require.ErrorIs(t, prepareCustomDedicatedSettlement(p, nil), ErrDedicatedAccess)
	}
	p := &postUsageBillingParams{APIKey: customDedicatedBillingTestKey(), User: &User{ID: 11}, Account: &Account{ID: 22}, Cost: &CostBreakdown{TotalCost: 2}}
	_, err := applyUsageBilling(context.Background(), "request", nil, p, &billingDeps{}, nil)
	require.ErrorIs(t, err, ErrDedicatedAccess)
}

func TestCustomDedicatedPrepareNoBindingAndCacheIsolation(t *testing.T) {
	for _, hasBinding := range []bool{false, true} {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		key := customDedicatedBillingTestKey()
		key.customDedicatedBilling = nil
		binding := customDedicatedTestBinding(time.Now())
		rows := sqlmock.NewRows([]string{"id", "user_id", "account_id", "group_id", "label", "expires_at", "revoked_at", "updated_at", "user_ids"})
		if hasBinding {
			rows = customDedicatedTestRows(binding)
		}
		mock.ExpectQuery("SELECT .* FROM custom_dedicated_accounts WHERE deleted_at IS NULL").WithArgs(int64(33)).WillReturnRows(rows)
		if hasBinding {
			mock.ExpectQuery(regexp.QuoteMeta(customDedicatedAccessSQL)).WillReturnRows(sqlmock.NewRows([]string{"valid"}).AddRow(true))
		}
		keys := &APIKeyService{customDedicated: &CustomDedicatedService{db: db}}
		result, err := keys.PrepareCustomDedicatedBilling(context.Background(), key)
		require.NoError(t, err)
		require.Equal(t, hasBinding, result.IsCustomDedicatedPrepaid())
		require.False(t, key.IsCustomDedicatedPrepaid())
		if hasBinding {
			require.NotSame(t, key, result)
		}
		mock.ExpectClose()
		require.NoError(t, db.Close())
		require.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestCustomDedicatedPrepareDatabaseErrorFailsClosed(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	mock.ExpectQuery("SELECT .* FROM custom_dedicated_accounts").WillReturnError(errors.New("database unavailable"))
	key := customDedicatedBillingTestKey()
	key.customDedicatedBilling = nil
	result, err := (&APIKeyService{customDedicated: &CustomDedicatedService{db: db}}).PrepareCustomDedicatedBilling(context.Background(), key)
	require.ErrorIs(t, err, ErrDedicatedAccess)
	require.Nil(t, result)
}

func TestCustomDedicatedZeroBalanceEligibility(t *testing.T) {
	cache := NewBillingCacheService(&balanceEligibilityCacheStub{}, nil, nil, nil, nil, nil, &config.Config{}, nil)
	defer cache.Stop()
	key := customDedicatedBillingTestKey()
	key.RateLimit5h = 0
	require.NoError(t, cache.CheckBillingEligibility(context.Background(), key.User, key, nil, nil, "openai"))
	key.customDedicatedBilling = nil
	require.Error(t, cache.CheckBillingEligibility(context.Background(), key.User, key, nil, nil, "openai"))
}

func TestCustomDedicatedExecutionIdentity(t *testing.T) {
	key := customDedicatedBillingTestKey()
	require.Equal(t, key.CustomDedicatedUsageRequestID("provider-a"), key.CustomDedicatedUsageRequestID("provider-b"))
	first := key.CustomDedicatedUsageRequestID("provider-a")
	key.customDedicatedBilling.LeaseID = "another-execution"
	require.NotEqual(t, first, key.CustomDedicatedUsageRequestID("provider-a"))
	key.customDedicatedBilling.WebSocket = true
	require.NotEqual(t, key.CustomDedicatedUsageRequestID("turn-a"), key.CustomDedicatedUsageRequestID("turn-b"))
	require.Empty(t, key.CustomDedicatedUsageRequestID(""))
}

func TestCustomDedicatedRecordUsagePlatformsAndWebSocket(t *testing.T) {
	for _, protocol := range []string{"claude", "openai", "websocket"} {
		t.Run(protocol, func(t *testing.T) {
			logs := &openAIRecordUsageLogRepoStub{inserted: true}
			billing := &openAIRecordUsageBillingRepoStub{}
			users := &openAIRecordUsageUserRepoStub{}
			subs := &openAIRecordUsageSubRepoStub{}
			key := customDedicatedBillingTestKey()
			key.customDedicatedBilling.WebSocket = protocol == "websocket"
			account := &Account{ID: 22, Type: AccountTypeOAuth}
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // worker结算不依赖HTTP生命周期。
			if protocol == "claude" {
				svc := newGatewayRecordUsageServiceWithBillingRepoForTest(logs, billing, users, subs)
				require.NoError(t, svc.RecordUsage(ctx, &RecordUsageInput{Result: &ForwardResult{RequestID: "client-reused", Model: "claude-sonnet-4", Usage: ClaudeUsage{InputTokens: 100, OutputTokens: 50}}, APIKey: key, User: key.User, Account: account, APIKeyService: &openAIRecordUsageAPIKeyQuotaStub{}}))
			} else {
				svc := newOpenAIRecordUsageServiceWithBillingRepoForTest(logs, billing, users, subs, nil)
				require.NoError(t, svc.RecordUsage(ctx, &OpenAIRecordUsageInput{Result: &OpenAIForwardResult{RequestID: "upstream-turn-1", Model: "gpt-5.1", Usage: OpenAIUsage{InputTokens: 100, OutputTokens: 50}, OpenAIWSMode: protocol == "websocket"}, APIKey: key, User: key.User, Account: account, APIKeyService: &openAIRecordUsageAPIKeyQuotaStub{}}))
			}
			require.Equal(t, 1, billing.calls)
			require.NoError(t, billing.lastCtxErr)
			require.Zero(t, billing.lastCmd.BalanceCost)
			require.Zero(t, billing.lastCmd.SubscriptionCost)
			require.Equal(t, BillingTypeDedicated, billing.lastCmd.BillingType)
			require.Equal(t, BillingTypeDedicated, logs.lastLog.BillingType)
			require.Zero(t, logs.lastLog.ActualCost)
			require.Greater(t, logs.lastLog.TotalCost, 0.0)
			require.Equal(t, key.CustomDedicatedUsageRequestID("upstream-turn-1"), billing.lastCmd.RequestID)
			require.Equal(t, QuantizeUsageBillingAmount(logs.lastLog.TotalCost*1.1), billing.lastCmd.APIKeyQuotaCost)
			require.Equal(t, 0, users.deductCalls)
			require.Equal(t, 0, subs.incrementCalls)
		})
	}
}

func TestCustomDedicatedSettlementDoesNotTouchBalanceCache(t *testing.T) {
	cache := &balanceEligibilityCacheStub{balance: 10}
	svc := NewBillingCacheService(cache, nil, nil, nil, nil, nil, &config.Config{}, nil)
	key := customDedicatedBillingTestKey()
	p := &postUsageBillingParams{Cost: &CostBreakdown{TotalCost: 2, ActualCost: 9}, APIKey: key, User: key.User, Account: &Account{ID: 22}}
	finalizePostUsageBilling(context.Background(), p, &billingDeps{billingCacheService: svc, deferredService: &DeferredService{}}, &UsageBillingApplyResult{Applied: true})
	svc.Stop() // 等待缓存写队列完成后再断言。
	require.Zero(t, cache.deductCalls.Load())
	require.Zero(t, cache.invalidateCalls.Load())
	postUsageBilling(context.Background(), p, &billingDeps{}) // 旧兜底同样不执行余额扣款。
}
