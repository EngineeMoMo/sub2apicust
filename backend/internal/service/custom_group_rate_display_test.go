//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCustomGroupRateAdminSchedules(t *testing.T) {
	access, store, _ := customAccessTestFixture()
	period := CustomGroupRatePeriod{StartsAt: access.now(), EndsAt: access.now().Add(time.Hour), RateMultiplier: 0.8}
	store.policies[1] = CustomGroupAccessPolicy{GroupID: 1, RateScheduleEnabled: true, RateSchedule: []CustomGroupRatePeriod{period}}
	store.policies[2] = CustomGroupAccessPolicy{GroupID: 2, RateSchedule: []CustomGroupRatePeriod{period}}
	display, err := access.AdminRateSchedules(context.Background())
	require.NoError(t, err)
	require.Equal(t, access.now(), display.ServerTime)
	require.Len(t, display.Schedules, 1)
	require.True(t, display.Schedules[1].RateScheduleEnabled)
	require.Equal(t, []CustomGroupRatePeriod{period}, display.Schedules[1].RateSchedule)
	store.listErr = errors.New("database unavailable")
	display, err = access.AdminRateSchedules(context.Background())
	require.ErrorIs(t, err, store.listErr)
	require.Nil(t, display)
}
