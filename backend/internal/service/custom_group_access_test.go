//go:build unit

package service

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type customAccessTestStore struct {
	policies   map[int64]CustomGroupAccessPolicy
	facts      customGroupAccessFacts
	listErr    error
	factsErr   error
	factsCalls int
	start      time.Time
	end        time.Time
}

func (s *customAccessTestStore) List(context.Context) (map[int64]CustomGroupAccessPolicy, error) {
	return s.policies, s.listErr
}
func (s *customAccessTestStore) Facts(_ context.Context, _ int64, start, end time.Time) (customGroupAccessFacts, error) {
	s.factsCalls++
	s.start, s.end = start, end
	return s.facts, s.factsErr
}
func (s *customAccessTestStore) Save(_ context.Context, policy CustomGroupAccessPolicy) (*CustomGroupAccessPolicy, error) {
	s.policies[policy.GroupID] = policy
	return &policy, nil
}

type customAccessTestGroups struct {
	GroupRepository
	values map[int64]*Group
}

func (r customAccessTestGroups) GetByID(_ context.Context, groupID int64) (*Group, error) {
	if group := r.values[groupID]; group != nil {
		return group, nil
	}
	return nil, ErrGroupNotFound
}

func (r customAccessTestGroups) GetByIDLite(ctx context.Context, groupID int64) (*Group, error) {
	return r.GetByID(ctx, groupID)
}
func (r customAccessTestGroups) ListActive(context.Context) ([]Group, error) {
	var result []Group
	for _, group := range r.values {
		if group.IsActive() {
			result = append(result, *group)
		}
	}
	return result, nil
}

type customAccessTestUsers struct {
	UserRepository
	user *User
}

func (r customAccessTestUsers) GetByID(context.Context, int64) (*User, error) {
	cloned := *r.user
	return &cloned, nil
}

type customAccessTestSubs struct {
	UserSubscriptionRepository
	active map[int64]bool
}

type customAccessTestKeys struct {
	APIKeyRepository
	key *APIKey
}

func (r customAccessTestKeys) GetByID(context.Context, int64) (*APIKey, error) {
	cloned := *r.key
	return &cloned, nil
}

func (r customAccessTestKeys) GetByKeyForAuth(context.Context, string) (*APIKey, error) {
	cloned := *r.key
	return &cloned, nil
}

func (r customAccessTestSubs) GetActiveByUserIDAndGroupID(_ context.Context, _, groupID int64) (*UserSubscription, error) {
	if r.active[groupID] {
		return &UserSubscription{GroupID: groupID}, nil
	}
	return nil, errors.New("no active subscription")
}
func (r customAccessTestSubs) ListActiveByUserID(context.Context, int64) ([]UserSubscription, error) {
	return nil, nil
}

func customAccessTestFixture() (*CustomGroupAccessService, *customAccessTestStore, *APIKey) {
	now := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	store := &customAccessTestStore{policies: map[int64]CustomGroupAccessPolicy{}, facts: customGroupAccessFacts{Balance: 100, MonthlyRecharge: 100}}
	groups := customAccessTestGroups{values: map[int64]*Group{
		1: {ID: 1, Platform: PlatformAnthropic, Status: StatusActive, RateMultiplier: 2},
		2: {ID: 2, Platform: PlatformAnthropic, Status: StatusActive, RateMultiplier: 0.5},
		3: {ID: 3, Platform: PlatformAnthropic, Status: StatusActive},
	}}
	user := &User{ID: 7, Status: StatusActive, Balance: 100}
	svc := &CustomGroupAccessService{store: store, groups: groups, users: customAccessTestUsers{user: user}, subscriptions: customAccessTestSubs{}, now: func() time.Time { return now }}
	groupID := int64(1)
	key := &APIKey{ID: 4, UserID: user.ID, User: user, GroupID: &groupID, Group: groups.values[groupID], Status: StatusActive}
	return svc, store, key
}

func TestCustomGroupAccessIndependentThresholds(t *testing.T) {
	for _, test := range []struct {
		name             string
		monthly, balance bool
		paid, remaining  float64
		allowed          bool
	}{
		{"both disabled", false, false, 0, 0, true},
		{"monthly only exact", true, false, 100, 0, true},
		{"monthly only low", true, false, 99.99, 1000, false},
		{"balance only exact", false, true, 0, 50, true},
		{"balance only low", false, true, 1000, 49.99, false},
		{"both exact", true, true, 100, 50, true},
		{"both paid low", true, true, 99, 100, false},
		{"both balance low", true, true, 100, 49, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			svc, store, key := customAccessTestFixture()
			store.policies[1] = CustomGroupAccessPolicy{GroupID: 1, MonthlyRechargeEnabled: test.monthly, MonthlyRechargeMin: 100, BalanceEnabled: test.balance, BalanceMin: 50}
			store.facts = customGroupAccessFacts{Balance: test.remaining, MonthlyRecharge: test.paid}
			prepared, err := svc.PrepareKey(context.Background(), key)
			if test.allowed {
				require.NoError(t, err)
				require.Equal(t, int64(1), *prepared.GroupID)
			} else {
				require.ErrorIs(t, err, ErrGroupAccessRequirements)
			}
			require.Equal(t, float64(100), key.User.Balance)
		})
	}
}

func TestCustomGroupAccessRefundRechecksExistingKey(t *testing.T) {
	svc, store, key := customAccessTestFixture()
	store.policies[1] = CustomGroupAccessPolicy{GroupID: 1, MonthlyRechargeEnabled: true, MonthlyRechargeMin: 100}
	_, err := svc.PrepareKey(context.Background(), key)
	require.NoError(t, err)
	store.facts.MonthlyRecharge = 0
	_, err = svc.PrepareKey(context.Background(), key)
	require.ErrorIs(t, err, ErrGroupAccessRequirements)
	require.Equal(t, 2, store.factsCalls)
	denied, err := svc.DeniedGroupIDs(context.Background(), key.UserID)
	require.NoError(t, err)
	require.True(t, denied[1])
}

func TestCustomGroupAccessExpiryUsesTargetWithoutMutatingBinding(t *testing.T) {
	svc, store, key := customAccessTestFixture()
	expires := svc.now()
	fallback := int64(2)
	store.policies[1] = CustomGroupAccessPolicy{GroupID: 1, ExpiresAt: &expires, FallbackGroupID: &fallback, BalanceEnabled: true, BalanceMin: 999}
	prepared, err := svc.PrepareKey(context.Background(), key)
	require.NoError(t, err)
	require.Equal(t, int64(2), *prepared.GroupID)
	require.Equal(t, int64(2), prepared.Group.ID)
	require.Equal(t, 0.5, prepared.Group.RateMultiplier)
	require.Equal(t, int64(1), *key.GroupID)
	require.Equal(t, float64(2), key.Group.RateMultiplier)
	denied, err := svc.DeniedGroupIDs(context.Background(), key.UserID)
	require.NoError(t, err)
	require.True(t, denied[1])
	require.False(t, denied[2])
	require.Zero(t, store.factsCalls)
}

func TestCustomGroupAccessBeforeExpiryKeepsOriginal(t *testing.T) {
	svc, store, key := customAccessTestFixture()
	expires := svc.now().Add(time.Nanosecond)
	fallback := int64(2)
	store.policies[1] = CustomGroupAccessPolicy{GroupID: 1, ExpiresAt: &expires, FallbackGroupID: &fallback}
	prepared, err := svc.PrepareKey(context.Background(), key)
	require.NoError(t, err)
	require.Equal(t, int64(1), *prepared.GroupID)
}

func TestCustomGroupAccessFallbackFailsClosed(t *testing.T) {
	for _, scenario := range []string{"exclusive", "restricted public", "subscription", "inactive", "other platform", "target threshold", "cycle", "missing target", "no target"} {
		t.Run(scenario, func(t *testing.T) {
			svc, store, key := customAccessTestFixture()
			expires := svc.now()
			fallback := int64(2)
			store.policies[1] = CustomGroupAccessPolicy{GroupID: 1, ExpiresAt: &expires, FallbackGroupID: &fallback}
			groups := svc.groups.(customAccessTestGroups)
			switch scenario {
			case "exclusive":
				groups.values[2].IsExclusive = true
			case "restricted public":
				svc.users.(customAccessTestUsers).user.RestrictPublicGroups = true
				svc.users.(customAccessTestUsers).user.AllowedGroups = []int64{1}
			case "subscription":
				groups.values[2].SubscriptionType = "subscription"
			case "inactive":
				groups.values[2].Status = "inactive"
			case "other platform":
				groups.values[2].Platform = PlatformOpenAI
			case "target threshold":
				store.policies[2] = CustomGroupAccessPolicy{GroupID: 2, MonthlyRechargeEnabled: true, MonthlyRechargeMin: 101}
			case "cycle":
				source := int64(1)
				store.policies[2] = CustomGroupAccessPolicy{GroupID: 2, ExpiresAt: &expires, FallbackGroupID: &source}
			case "missing target":
				delete(groups.values, 2)
			case "no target":
				policy := store.policies[1]
				policy.FallbackGroupID = nil
				store.policies[1] = policy
			}
			_, err := svc.PrepareKey(context.Background(), key)
			require.Error(t, err)
		})
	}
}

func TestCustomGroupAccessVisibilityAndFailures(t *testing.T) {
	svc, store, key := customAccessTestFixture()
	store.policies[1] = CustomGroupAccessPolicy{GroupID: 1, MonthlyRechargeEnabled: true, MonthlyRechargeMin: 10}
	store.policies[2] = CustomGroupAccessPolicy{GroupID: 2, BalanceEnabled: true, BalanceMin: 20}
	store.policies[3] = CustomGroupAccessPolicy{GroupID: 3}
	denied, err := svc.DeniedGroupIDs(context.Background(), 0)
	require.NoError(t, err)
	require.Equal(t, map[int64]bool{1: true, 2: true}, denied)
	require.Zero(t, store.factsCalls)
	denied, err = svc.DeniedGroupIDs(context.Background(), key.UserID)
	require.NoError(t, err)
	require.Empty(t, denied)
	require.Equal(t, 1, store.factsCalls)
	store.factsErr = errors.New("database unavailable")
	_, err = svc.PrepareKey(context.Background(), key)
	require.Error(t, err)
	_, err = svc.DeniedGroupIDs(context.Background(), key.UserID)
	require.Error(t, err)
	store.listErr = errors.New("policy query failed")
	_, err = svc.PrepareKey(context.Background(), key)
	require.Error(t, err)
}

func TestCustomGroupAccessBeijingMonth(t *testing.T) {
	for _, instant := range []time.Time{
		time.Date(2026, 9, 30, 15, 59, 59, 0, time.UTC),
		time.Date(2026, 9, 30, 16, 0, 0, 0, time.UTC),
		time.Date(2026, 12, 31, 16, 0, 0, 0, time.UTC),
	} {
		start, end := customGroupAccessMonth(instant)
		local := instant.In(time.FixedZone("Beijing", 28800))
		require.Equal(t, local.Month(), start.Month())
		require.Equal(t, local.Year(), start.Year())
		require.Equal(t, 1, start.Day())
		require.Equal(t, 0, start.Hour())
		require.False(t, instant.Before(start))
		require.True(t, instant.Before(end))
	}
}

func TestCustomGroupAccessValidatesInputsAndGraph(t *testing.T) {
	for _, amount := range []float64{-1, math.NaN(), math.Inf(1), 0.001, 1e13, 0} {
		require.Error(t, validateCustomGroupAccessPolicy(CustomGroupAccessPolicy{GroupID: 1, MonthlyRechargeEnabled: true, MonthlyRechargeMin: amount}))
	}
	expires := time.Now()
	source, target := int64(1), int64(2)
	require.Error(t, validateCustomGroupAccessPolicy(CustomGroupAccessPolicy{GroupID: 1, ExpiresAt: &expires}))
	require.Error(t, validateCustomGroupAccessPolicy(CustomGroupAccessPolicy{GroupID: 1, FallbackGroupID: &target}))
	require.Error(t, validateCustomGroupAccessPolicy(CustomGroupAccessPolicy{GroupID: 1, ExpiresAt: &expires, FallbackGroupID: &source}))
	require.NoError(t, validateCustomGroupAccessPolicy(CustomGroupAccessPolicy{GroupID: 1, ExpiresAt: &expires, FallbackGroupID: &target}))
	require.Error(t, validateCustomGroupAccessGraph(map[int64]CustomGroupAccessPolicy{1: {FallbackGroupID: &target}, 2: {FallbackGroupID: &source}}))
	require.NoError(t, validateCustomGroupAccessGraph(map[int64]CustomGroupAccessPolicy{1: {FallbackGroupID: &target}}))
}

func TestCustomGroupAccessAPIKeyBindingAndList(t *testing.T) {
	svc, store, key := customAccessTestFixture()
	store.policies[1] = CustomGroupAccessPolicy{GroupID: 1, MonthlyRechargeEnabled: true, MonthlyRechargeMin: 101}
	keys := NewAPIKeyService(customAccessTestKeys{key: key}, svc.users, svc.groups, svc.subscriptions, nil, nil, nil)
	keys.customGroupAccess = svc
	_, err := keys.Create(context.Background(), key.UserID, CreateAPIKeyRequest{Name: "blocked", GroupID: key.GroupID})
	require.ErrorIs(t, err, ErrGroupAccessRequirements)
	_, err = keys.Update(context.Background(), key.ID, key.UserID, UpdateAPIKeyRequest{GroupID: key.GroupID})
	require.ErrorIs(t, err, ErrGroupAccessRequirements)
	groups, err := keys.GetAvailableGroups(context.Background(), key.UserID)
	require.NoError(t, err)
	for _, group := range groups {
		require.NotEqual(t, int64(1), group.ID)
	}
}

func TestCustomGroupAccessWebSocketRechecksRefundAndExpiry(t *testing.T) {
	svc, store, original := customAccessTestFixture()
	original.Key = "custom-access-existing-key"
	keys := NewAPIKeyService(customAccessTestKeys{key: original}, svc.users, svc.groups, svc.subscriptions, nil, nil, nil)
	keys.customGroupAccess = svc
	store.policies[1] = CustomGroupAccessPolicy{GroupID: 1, MonthlyRechargeEnabled: true, MonthlyRechargeMin: 100}
	require.NoError(t, keys.ValidateCustomGroupAccessContinuation(context.Background(), original))
	store.facts.MonthlyRecharge = 0
	require.ErrorIs(t, keys.ValidateCustomGroupAccessContinuation(context.Background(), original), ErrGroupAccessRequirements)
	expires, target := svc.now(), int64(2)
	store.policies[1] = CustomGroupAccessPolicy{GroupID: 1, ExpiresAt: &expires, FallbackGroupID: &target}
	require.ErrorIs(t, keys.ValidateCustomGroupAccessContinuation(context.Background(), original), ErrGroupAccessRouteChanged)
	reconnected, err := keys.PrepareCustomGroupAccess(context.Background(), original)
	require.NoError(t, err)
	require.Equal(t, int64(2), *reconnected.GroupID)
	require.NoError(t, keys.ValidateCustomGroupAccessContinuation(context.Background(), reconnected))
}
