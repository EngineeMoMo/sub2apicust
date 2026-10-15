//go:build unit

package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCustomGroupExpiryClockDelay(t *testing.T) {
	now := time.Now()
	for _, test := range []struct {
		name string
		next sql.NullTime
		want time.Duration
	}{
		{"none", sql.NullTime{}, 30 * time.Second},
		{"future", sql.NullTime{Time: now.Add(time.Hour), Valid: true}, 30 * time.Second},
		{"boundary", sql.NullTime{Time: now.Add(100 * time.Millisecond), Valid: true}, 100 * time.Millisecond},
		{"due", sql.NullTime{Time: now.Add(-time.Second), Valid: true}, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, customGroupExpiryDelay(test.next, now))
		})
	}
}

func TestCustomGroupExpiryOldAuthSnapshotAfterDeletion(t *testing.T) {
	access, store, key := customAccessTestFixture()
	expired, target := access.now().Add(-time.Second), int64(2)
	store.policies[1] = CustomGroupAccessPolicy{GroupID: 1, ExpiresAt: &expired, FallbackGroupID: &target}
	delete(access.groups.(customAccessTestGroups).values, 1)
	prepared, err := access.PrepareKey(context.Background(), key)
	require.NoError(t, err)
	require.Equal(t, target, *prepared.GroupID)
	require.Equal(t, float64(0.5), prepared.Group.RateMultiplier)
	require.EqualValues(t, 1, *key.GroupID)
}

func TestCustomGroupRateAvailableGroupsUsesServerClock(t *testing.T) {
	access, store, key := customAccessTestFixture()
	groups := access.groups.(customAccessTestGroups)
	groups.values[1].RateMultiplier = 1
	start := access.now()
	end := start.Add(time.Minute)
	store.policies[1] = CustomGroupAccessPolicy{GroupID: 1, RateScheduleEnabled: true, RateSchedule: []CustomGroupRatePeriod{{StartsAt: start, EndsAt: end, RateMultiplier: 0.8}}}
	keys := &APIKeyService{userRepo: access.users, groupRepo: groups, userSubRepo: customAccessTestSubs{}, customGroupAccess: access}
	for _, test := range []struct {
		name string
		at   time.Time
		want float64
	}{
		{"before", start.Add(-time.Nanosecond), 1}, {"start", start, 0.8}, {"inside", start.Add(time.Second), 0.8}, {"end", end, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			access.now = func() time.Time { return test.at }
			available, err := keys.GetAvailableGroups(context.Background(), key.UserID)
			require.NoError(t, err)
			found := false
			for _, group := range available {
				if group.ID == 1 {
					found = true
					require.Equal(t, test.want, group.RateMultiplier)
				}
			}
			require.True(t, found)
			require.Equal(t, float64(1), groups.values[1].RateMultiplier)
		})
	}
}
