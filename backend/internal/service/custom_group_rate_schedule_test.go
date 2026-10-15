//go:build unit

package service

import (
	"context"
	"math"
	"testing"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/require"
)

func TestCustomGroupRateScheduleThreePeriodsAndFallback(t *testing.T) {
	svc, store, key := customAccessTestFixture()
	zone := time.FixedZone("Beijing", 8*3600)
	day := time.Date(2026, 10, 15, 0, 0, 0, 0, zone)
	expires := day.Add(23 * time.Hour)
	fallback := int64(2)
	store.policies[1] = CustomGroupAccessPolicy{GroupID: 1, ExpiresAt: &expires, FallbackGroupID: &fallback, RateScheduleEnabled: true, RateSchedule: []CustomGroupRatePeriod{
		{StartsAt: day, EndsAt: day.Add(18 * time.Hour), RateMultiplier: 0.1},
		{StartsAt: day.Add(18 * time.Hour), EndsAt: day.Add(20*time.Hour + 30*time.Minute), RateMultiplier: 0.15},
		{StartsAt: day.Add(20*time.Hour + 30*time.Minute), EndsAt: expires, RateMultiplier: 0.1},
	}}
	svc.cache = newCustomGroupAccessCache()
	svc.cache.health(true)
	resolver := newUserGroupRateResolver(&userGroupRateResolverRepoStub{}, gocache.New(time.Minute, time.Minute), time.Minute, nil, "service.test")
	for _, test := range []struct {
		name  string
		at    time.Time
		rate  float64
		group int64
	}{
		{"before", day.Add(-time.Second), 2, 1},
		{"first", day.Add(17*time.Hour + 59*time.Minute + 59*time.Second), 0.1, 1},
		{"18:00", day.Add(18 * time.Hour), 0.15, 1},
		{"20:30", day.Add(20*time.Hour + 30*time.Minute), 0.1, 1},
		{"23:00", expires, 0.5, 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			svc.now = func() time.Time { return test.at }
			prepared, err := svc.PrepareKey(context.Background(), key)
			require.NoError(t, err)
			require.Equal(t, test.group, *prepared.GroupID)
			rate := resolver.Resolve(context.Background(), key.UserID, *prepared.GroupID, prepared.Group.RateMultiplier)
			text, _ := computePeakAwareMultipliers(prepared, rate, test.at)
			require.Equal(t, test.rate, text)
			require.Equal(t, float64(2), key.Group.RateMultiplier)
		})
	}
}

func TestCustomGroupRateScheduleValidation(t *testing.T) {
	start := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	valid := CustomGroupRatePeriod{StartsAt: start, EndsAt: start.Add(time.Hour), RateMultiplier: 0.1}
	for _, test := range []struct {
		name    string
		periods []CustomGroupRatePeriod
		expires *time.Time
	}{
		{"empty", nil, nil},
		{"overlap", []CustomGroupRatePeriod{valid, valid}, nil},
		{"reversed", []CustomGroupRatePeriod{{StartsAt: valid.EndsAt, EndsAt: valid.StartsAt, RateMultiplier: 0.1}}, nil},
		{"negative", []CustomGroupRatePeriod{{StartsAt: valid.StartsAt, EndsAt: valid.EndsAt, RateMultiplier: -1}}, nil},
		{"nan", []CustomGroupRatePeriod{{StartsAt: valid.StartsAt, EndsAt: valid.EndsAt, RateMultiplier: math.NaN()}}, nil},
		{"after-expiry", []CustomGroupRatePeriod{valid}, &start},
		{"too-many", make([]CustomGroupRatePeriod, 33), nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			require.Error(t, validateCustomGroupRateSchedule(CustomGroupAccessPolicy{RateScheduleEnabled: true, RateSchedule: test.periods, ExpiresAt: test.expires}))
		})
	}
	require.NoError(t, validateCustomGroupRateSchedule(CustomGroupAccessPolicy{RateScheduleEnabled: true, RateSchedule: []CustomGroupRatePeriod{valid}}))
}

func TestCustomGroupRateScheduleDisableAndUserOverride(t *testing.T) {
	policy := CustomGroupAccessPolicy{RateScheduleEnabled: false, RateSchedule: []CustomGroupRatePeriod{{StartsAt: time.Now().Add(-time.Hour), EndsAt: time.Now().Add(time.Hour), RateMultiplier: 0.1}}}
	require.Equal(t, 0.8, policy.rateAt(time.Now(), 0.8))
	userRate := 0.07
	resolver := newUserGroupRateResolver(&userGroupRateResolverRepoStub{rate: &userRate}, gocache.New(time.Minute, time.Minute), time.Minute, nil, "service.test")
	require.Equal(t, userRate, resolver.Resolve(context.Background(), 7, 1, 0.1))
	require.Equal(t, userRate, resolver.Resolve(context.Background(), 7, 1, 0.15))
}

func TestCustomGroupRateScheduleBatchImageSnapshot(t *testing.T) {
	access, store, _ := customAccessTestFixture()
	now := access.now()
	groupID := int64(7)
	store.policies[groupID] = CustomGroupAccessPolicy{GroupID: groupID, RateScheduleEnabled: true, RateSchedule: []CustomGroupRatePeriod{{StartsAt: now.Add(-time.Hour), EndsAt: now.Add(time.Hour), RateMultiplier: 0.15}}}
	svc, repo, _, _, _ := newTestBatchImagePublicService(true)
	svc.GroupAccess = access
	svc.UserGroupRateRepo = nil
	svc.GroupRepo = &publicBatchImageGroupRepo{groups: map[int64]*Group{groupID: {ID: groupID, Platform: PlatformGemini, RateMultiplier: 2, AllowImageGeneration: true, AllowBatchImageGeneration: true, BatchImageDiscountMultiplier: 0.5, BatchImageHoldMultiplier: 0.6}}}
	result, err := svc.Submit(context.Background(), BatchImageOwner{UserID: 11, APIKeyID: 22, GroupID: &groupID}, validBatchImageSubmitRequest(), "")
	require.NoError(t, err)
	require.InDelta(t, 0.15, repo.jobs[result.ID].GroupRateMultiplier, 1e-12)
}
