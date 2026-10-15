//go:build unit

package service

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
)

type customAccessFallbackAccounts struct {
	*mockAccountRepoForPlatform
	queriedGroups []int64
}

func (r *customAccessFallbackAccounts) ListSchedulableByGroupIDAndPlatforms(ctx context.Context, groupID int64, platforms []string) ([]Account, error) {
	r.queriedGroups = append(r.queriedGroups, groupID)
	return r.ListModelAvailabilityCandidates(ctx, &groupID, platforms, false)
}

func (r *customAccessFallbackAccounts) ListSchedulableByGroupIDAndPlatform(ctx context.Context, groupID int64, platform string) ([]Account, error) {
	return r.ListSchedulableByGroupIDAndPlatforms(ctx, groupID, []string{platform})
}

func TestCustomGroupAccessLegacyFallbackChecksTarget(t *testing.T) {
	for _, mode := range []string{"balance", "monthly", "exclusive", "inactive"} {
		t.Run(mode, func(t *testing.T) {
			access, store, key := customAccessTestFixture()
			fallbackID := int64(2)
			groups := access.groups.(customAccessTestGroups)
			groups.values[1].ClaudeCodeOnly = true
			groups.values[1].FallbackGroupID = &fallbackID
			groups.values[1].Hydrated = true
			groups.values[2].Hydrated = true
			groups.values[2].IsExclusive = mode == "exclusive"
			if mode == "inactive" {
				groups.values[2].Status = "disabled"
			}
			store.policies[2] = CustomGroupAccessPolicy{GroupID: 2, BalanceEnabled: mode == "balance", BalanceMin: 200, MonthlyRechargeEnabled: mode == "monthly", MonthlyRechargeMin: 200}
			account := Account{ID: 42, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true, AccountGroups: []AccountGroup{{AccountID: 42, GroupID: fallbackID}}}
			accounts := &customAccessFallbackAccounts{mockAccountRepoForPlatform: &mockAccountRepoForPlatform{accounts: []Account{account}, accountsByID: map[int64]*Account{42: &account}}}
			gateway := &GatewayService{groupRepo: groups, accountRepo: accounts, cfg: testConfig(), customGroupAccess: access}
			ctx := context.WithValue(context.Background(), ctxkey.UserID, key.UserID)
			ctx = context.WithValue(ctx, ctxkey.Group, key.Group)
			selected, err := gateway.SelectAccountForModelWithExclusions(ctx, key.GroupID, "", "claude-sonnet-4-6", nil)
			require.Error(t, err)
			require.Nil(t, selected)
			require.Empty(t, accounts.queriedGroups)
			_, err = gateway.PrepareCustomGatewayGroup(ctx, key)
			require.Error(t, err)
		})
	}
}

func TestCustomGroupAccessFallbackExpiryRateAndSnapshot(t *testing.T) {
	access, store, key := customAccessTestFixture()
	groups := access.groups.(customAccessTestGroups)
	fallbackID, finalID := int64(2), int64(3)
	groups.values[1].ClaudeCodeOnly, groups.values[1].FallbackGroupID = true, &fallbackID
	groups.values[3].RateMultiplier = 1
	now := access.now()
	expired := now.Add(-time.Minute)
	store.policies[2] = CustomGroupAccessPolicy{GroupID: 2, ExpiresAt: &expired, FallbackGroupID: &finalID}
	store.policies[3] = CustomGroupAccessPolicy{GroupID: 3, RateScheduleEnabled: true, RateSchedule: []CustomGroupRatePeriod{{StartsAt: now.Add(-time.Hour), EndsAt: now.Add(time.Hour), RateMultiplier: 0.1}}}
	gateway := &GatewayService{groupRepo: groups, customGroupAccess: access}
	ctx := context.WithValue(context.Background(), ctxkey.UserID, key.UserID)
	prepared, err := gateway.PrepareCustomGatewayGroup(ctx, key)
	require.NoError(t, err)
	require.Equal(t, finalID, *prepared.GroupID)
	require.Equal(t, 0.1, prepared.Group.RateMultiplier)
	require.EqualValues(t, 1, *key.GroupID)
	require.Equal(t, float64(1), groups.values[3].RateMultiplier)
	claudeCtx := SetClaudeCodeClient(ctx, true)
	unchanged, err := gateway.PrepareCustomGatewayGroup(claudeCtx, key)
	require.NoError(t, err)
	require.Same(t, key, unchanged)
	direct, err := access.PrepareFallback(ctx, key, groups.values[2])
	require.NoError(t, err)
	require.Equal(t, finalID, *direct.GroupID)
	require.Equal(t, 0.1, direct.Group.RateMultiplier)
}

func TestCustomGroupAccessWarmRPMSnapshot(t *testing.T) {
	for _, mode := range []string{"same-existing", "same-empty", "target-existing", "target-empty", "target-unlimited"} {
		t.Run(mode, func(t *testing.T) {
			access, store, key := customAccessTestFixture()
			access.cache = newCustomGroupAccessCache()
			access.cache.health(true)
			original := 1000
			key.User.UserGroupRPMOverride = &original
			store.policies[1] = CustomGroupAccessPolicy{GroupID: 1, BalanceEnabled: true, BalanceMin: 80}
			var override *int
			if mode == "target-existing" || mode == "target-unlimited" {
				value := 2000
				if mode == "target-unlimited" {
					value = 0
				}
				override = &value
			}
			if mode == "same-empty" {
				key.User.UserGroupRPMOverride = nil
			}
			if mode == "target-existing" || mode == "target-empty" || mode == "target-unlimited" {
				expired, target := access.now().Add(-time.Second), int64(2)
				store.policies[1] = CustomGroupAccessPolicy{GroupID: 1, ExpiresAt: &expired, FallbackGroupID: &target}
			}
			rates := &rpmOverrideRepoStub{override: override}
			access.rates = rates
			billing := newBillingServiceForRPM(t, &userRPMCacheStub{}, rates)
			for request := 0; request < 100; request++ {
				prepared, err := access.PrepareKey(context.Background(), key)
				require.NoError(t, err)
				require.True(t, prepared.User.UserGroupRPMOverrideLoaded)
				if mode == "same-existing" {
					require.Equal(t, &original, prepared.User.UserGroupRPMOverride)
				} else {
					require.Equal(t, override, prepared.User.UserGroupRPMOverride)
				}
				require.NoError(t, billing.checkRPM(context.Background(), prepared.User, prepared.Group))
			}
			if mode == "same-existing" {
				require.Zero(t, atomic.LoadInt32(&rates.calls))
			} else {
				require.EqualValues(t, 1, atomic.LoadInt32(&rates.calls))
				require.NoError(t, access.cache.changed(`{"kind":"user","id":7}`))
				_, err := access.PrepareKey(context.Background(), key)
				require.NoError(t, err)
				require.EqualValues(t, 2, atomic.LoadInt32(&rates.calls))
			}
			require.Equal(t, 1000, original)
		})
	}
}
