//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/DATA-DOG/go-sqlmock"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type customRedeemCachePublisher struct {
	BillingCache
	invalidated bool
	published   string
}

func (cache *customRedeemCachePublisher) InvalidateSubscriptionCache(context.Context, int64, int64) error {
	cache.invalidated = true
	return nil
}
func (cache *customRedeemCachePublisher) PublishSubscriptionCacheInvalidation(_ context.Context, key string) error {
	cache.published = key
	return nil
}
func (cache *customRedeemCachePublisher) SubscribeSubscriptionCacheInvalidation(context.Context, func(string)) error {
	return nil
}

type customRedeemCacheRepo struct {
	lockingRenewalRepo
	service              *SubscriptionService
	committed            bool
	reloadedBeforeCommit bool
	activeReads          int
}

func (repo *customRedeemCacheRepo) GetActiveByUserIDAndGroupID(context.Context, int64, int64) (*UserSubscription, error) {
	repo.activeReads++
	if repo.committed {
		copy := repo.current
		return &copy, nil
	}
	copy := repo.stale
	return &copy, nil
}

func (repo *customRedeemCacheRepo) GetByID(ctx context.Context, subscriptionID int64) (*UserSubscription, error) {
	if dbent.TxFromContext(ctx) != nil {
		repo.service.subCacheL1.Wait()
		_, err := repo.service.GetActiveSubscription(context.Background(), repo.stale.UserID, repo.stale.GroupID)
		if err != nil {
			return nil, err
		}
		repo.service.subCacheL1.Wait()
		repo.reloadedBeforeCommit = true
	}
	return repo.lockingRenewalRepo.GetByID(ctx, subscriptionID)
}

func TestCustomRedeemCacheRefillInvalidatedAfterCommit(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, sqlDB)))
	t.Cleanup(func() { _ = client.Close() })
	previousExpiry := time.Now().AddDate(0, 0, 3)
	original := UserSubscription{ID: 1, UserID: 10, GroupID: 20, StartsAt: time.Now().AddDate(0, 0, -10), ExpiresAt: previousExpiry, Status: SubscriptionStatusActive}
	repo := &customRedeemCacheRepo{lockingRenewalRepo: lockingRenewalRepo{stale: original, current: original}}
	cfg := &config.Config{}
	cfg.SubscriptionCache.L1Size = 1000
	cfg.SubscriptionCache.L1TTLSeconds = 60
	subscriptions := NewSubscriptionService(&subscriptionGroupRepoStub{group: &Group{ID: 20, SubscriptionType: SubscriptionTypeSubscription}}, repo, nil, client, cfg)
	t.Cleanup(subscriptions.Stop)
	t.Cleanup(subscriptions.subCacheL1.Close)
	repo.service = subscriptions
	publisher := &customRedeemCachePublisher{}
	subscriptions.billingCacheService = &BillingCacheService{cache: publisher}
	groupID := int64(20)
	codes := &redeemRejectRepo{code: RedeemCode{ID: 7, Code: "review-cache-refill", Type: RedeemTypeSubscription, Status: StatusUnused, GroupID: &groupID, ValidityDays: 7}}
	redeem := NewRedeemService(codes, &mockUserRepo{getByIDUser: &User{ID: 10}}, subscriptions, nil, nil, client, nil, nil)

	mock.ExpectBegin()
	expectBulkSubscriptionStock(mock)
	mock.ExpectCommit()
	result, err := redeem.Redeem(context.Background(), 10, codes.code.Code)
	require.NoError(t, err)
	require.Equal(t, StatusUsed, result.Status)
	require.True(t, publisher.invalidated, "billing cache invalidation must complete before returning")
	require.Equal(t, subCacheKey(10, groupID), publisher.published, "post-commit invalidation must notify other instances")
	require.True(t, repo.reloadedBeforeCommit)
	require.Equal(t, previousExpiry.AddDate(0, 0, 7), repo.current.ExpiresAt)
	repo.committed = true
	cached, err := subscriptions.GetActiveSubscription(context.Background(), 10, groupID)
	require.NoError(t, err)
	require.Equal(t, previousExpiry.AddDate(0, 0, 7), cached.ExpiresAt, "post-commit invalidation removes the pre-commit value")
	require.Equal(t, 2, repo.activeReads, "post-commit access must reload the updated subscription")
	require.NoError(t, mock.ExpectationsWereMet())
}
