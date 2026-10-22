package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCustomAdminAssignmentPreservesSuspensionAfterStaleRead(t *testing.T) {
	now := time.Now()
	current := UserSubscription{
		ID: 27, UserID: 31, GroupID: 33, StartsAt: now.AddDate(0, 0, -10), ExpiresAt: now.AddDate(0, 0, 10),
		Status: SubscriptionStatusSuspended, Notes: "suspended", DailyUsageUSD: 4,
	}
	repo := &lockingRenewalRepo{
		stale:   UserSubscription{ID: 27, UserID: 31, GroupID: 33, ExpiresAt: now.Add(-time.Hour), Status: SubscriptionStatusExpired},
		current: current,
	}
	svc := NewSubscriptionService(nil, repo, nil, nil, nil)
	t.Cleanup(svc.Stop)
	subscription, err := svc.AssignAdminSubscription(context.Background(), &AssignSubscriptionInput{UserID: 31, GroupID: 33, ValidityDays: 3})
	require.ErrorIs(t, err, ErrSubscriptionSuspended)
	require.Nil(t, subscription)
	require.Equal(t, 1, repo.lockReads)
	require.Equal(t, current, repo.current)
}
