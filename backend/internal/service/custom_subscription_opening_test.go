//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type openingSubscriptionRepo struct {
	userSubRepoNoop
	saved        UserSubscription
	weeklyResets int
}

func (r *openingSubscriptionRepo) ActivateWindows(_ context.Context, _ int64, daily, periodic time.Time) error {
	r.saved.DailyWindowStart = &daily
	r.saved.WeeklyWindowStart = &periodic
	r.saved.MonthlyWindowStart = &periodic
	return nil
}

func (r *openingSubscriptionRepo) ResetWeeklyUsage(_ context.Context, _ int64, _ *time.Time, start time.Time) error {
	r.saved.WeeklyWindowStart = &start
	r.saved.WeeklyUsageUSD = 0
	r.weeklyResets++
	return nil
}

func TestCustomSubscriptionOpeningMaintenanceDoesNotResetFollowingRequest(t *testing.T) {
	start := time.Now().UTC().Add(-10 * 24 * time.Hour)
	repo := &openingSubscriptionRepo{saved: UserSubscription{ID: 1, StartsAt: start, ExpiresAt: start.Add(60 * 24 * time.Hour)}}
	svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)
	initial := repo.saved
	refreshed, err := svc.EnsureWindowMaintenance(context.Background(), &initial)
	require.NoError(t, err)
	require.Equal(t, start.Add(7*24*time.Hour), *refreshed.WeeklyWindowStart)
	require.Equal(t, start, *refreshed.MonthlyWindowStart)
	repo.saved.WeeklyUsageUSD = 9
	refreshed, err = svc.EnsureWindowMaintenance(context.Background(), &repo.saved)
	require.NoError(t, err)
	require.Equal(t, 9.0, refreshed.WeeklyUsageUSD)
	require.Equal(t, 1, repo.weeklyResets)
}

func (r *openingSubscriptionRepo) Create(_ context.Context, sub *UserSubscription) error {
	r.saved = *sub
	r.saved.ID = 1
	sub.ID = 1
	return nil
}

func (r *openingSubscriptionRepo) GetByID(context.Context, int64) (*UserSubscription, error) {
	sub := r.saved
	return &sub, nil
}

func TestCustomSubscriptionOpeningInitializesQuotaWithoutUsage(t *testing.T) {
	repo := &openingSubscriptionRepo{}
	svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)
	sub, err := svc.createSubscription(context.Background(), &AssignSubscriptionInput{UserID: 1, GroupID: 2, ValidityDays: 30})
	require.NoError(t, err)
	require.Equal(t, sub.StartsAt, *sub.WeeklyWindowStart)
	require.Equal(t, sub.StartsAt, *sub.MonthlyWindowStart)
	require.NotNil(t, sub.DailyWindowStart)
	require.Equal(t, sub.StartsAt.AddDate(0, 0, 30), sub.ExpiresAt)
	require.Zero(t, sub.WeeklyUsageUSD)
	require.Zero(t, sub.MonthlyUsageUSD)
	limit := 60.0
	p := svc.calculateProgress(sub, &Group{WeeklyLimitUSD: &limit, MonthlyLimitUSD: &limit})
	require.Equal(t, sub.StartsAt.Add(7*24*time.Hour), p.Weekly.ResetsAt)
	require.Equal(t, sub.StartsAt.Add(30*24*time.Hour), p.Monthly.ResetsAt)
}

func TestCustomSubscriptionOpeningProjectionMatchesListAndPreservesSource(t *testing.T) {
	start := time.Date(2026, 10, 1, 21, 4, 0, 0, time.UTC)
	sub := UserSubscription{StartsAt: start, ExpiresAt: start.Add(60 * 24 * time.Hour),
		WeeklyWindowStart: &start, MonthlyWindowStart: &start, WeeklyUsageUSD: 50, MonthlyUsageUSD: 90}
	now := start.Add(35*24*time.Hour + time.Hour)
	svc := &SubscriptionService{now: func() time.Time { return now }}
	limit := 100.0
	list := []UserSubscription{sub}
	normalizeExpiredWindowsAt(list, now)
	p := svc.calculateProgress(&sub, &Group{WeeklyLimitUSD: &limit, MonthlyLimitUSD: &limit})
	require.Equal(t, *list[0].WeeklyWindowStart, p.Weekly.WindowStart)
	require.Equal(t, *list[0].MonthlyWindowStart, p.Monthly.WindowStart)
	require.Equal(t, start.Add(42*24*time.Hour), p.Weekly.ResetsAt)
	require.Equal(t, start.Add(60*24*time.Hour), p.Monthly.ResetsAt)
	require.Zero(t, p.Weekly.UsedUSD)
	require.Zero(t, p.Monthly.UsedUSD)
	require.Equal(t, 50.0, sub.WeeklyUsageUSD)
	require.Equal(t, 90.0, sub.MonthlyUsageUSD)
	require.Equal(t, start, *sub.WeeklyWindowStart)
}
