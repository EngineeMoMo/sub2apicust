//go:build unit

package handler

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

// 经真实资格解析生成私有凭证，不通过公开字段伪造免费资格。
func customInflightPrepaidKey(t *testing.T) *service.APIKey {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	keys := &service.APIKeyService{}
	service.NewCustomDedicatedService(db, nil, keys, &service.GatewayService{}, &service.OpenAIGatewayService{})
	now := time.Now()
	mock.ExpectQuery("SELECT .* FROM custom_dedicated_accounts WHERE deleted_at IS NULL").WithArgs(int64(33)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "user_id", "account_id", "group_id", "label", "expires_at", "revoked_at", "updated_at", "user_ids"}).
			AddRow(1, 11, 22, 33, "包号", now.Add(time.Hour), nil, now, []byte(`[11]`)))
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"valid"}).AddRow(true))
	mock.ExpectQuery("SELECT .* FROM custom_dedicated_billing_policies").WithArgs(int64(1)).WillReturnError(sql.ErrNoRows)
	groupID := int64(33)
	key, err := keys.PrepareCustomDedicatedBilling(context.Background(), &service.APIKey{
		ID: 7, UserID: 11, User: &service.User{ID: 11, Balance: 0}, GroupID: &groupID,
		Group: &service.Group{ID: groupID, SubscriptionType: service.SubscriptionTypeStandard, RateMultiplier: 1},
	})
	require.NoError(t, err)
	require.True(t, key.IsCustomDedicatedPrepaid())
	require.NoError(t, mock.ExpectationsWereMet())
	return key
}

func TestCustomDedicatedInflightSkipsOnlyAuthenticatedPrepaid(t *testing.T) {
	key := customInflightPrepaidKey(t)
	cache := newHandlerInflightCache(0)
	cfg := &config.Config{}
	cfg.Billing.InflightReservation = config.InflightReservationConfig{Enabled: true, TTLSeconds: 60, FailClosedOnUnpriced: true}
	billing := service.NewBillingCacheService(cache, nil, nil, nil, nil, nil, cfg, nil)
	t.Cleanup(billing.Stop)
	// 即使同用户有普通请求占款，也不能占用包号的余额或拒绝第二个请求。
	held, err := billing.ReserveInflight(context.Background(), key.User, key.Group, nil, 1)
	require.NoError(t, err)
	t.Cleanup(held.HandlerDone)
	for _, priced := range []bool{true, false} {
		est := &countingEstimator{cost: 1, priced: priced}
		for range 2 {
			_, done, err := reserveInflightBalanceCtx(context.Background(), billing, est, key, nil, tokenInflightEstimate("m", nil))
			require.NoError(t, err)
			done()
		}
		require.Zero(t, est.calls)
		require.Equal(t, 1, cache.count())
	}
	// HTTP 使用同一入口；普通Key和凭证身份不匹配的Key必须仍执行余额预占。
	done, err := reserveInflightBalance(newInflightTestGinContext(), billing, &countingEstimator{cost: 1, priced: true}, key, nil, tokenInflightEstimate("m", nil))
	require.NoError(t, err)
	done()
	ordinary := &service.APIKey{ID: 8, UserID: key.UserID, User: key.User, GroupID: key.GroupID, Group: key.Group}
	tampered := *key
	tampered.ID++
	for _, billable := range []*service.APIKey{ordinary, &tampered} {
		_, _, err := reserveInflightBalanceCtx(context.Background(), billing, &countingEstimator{cost: 1, priced: true}, billable, nil, tokenInflightEstimate("m", nil))
		require.ErrorIs(t, err, service.ErrInsufficientBalance)
	}
}
