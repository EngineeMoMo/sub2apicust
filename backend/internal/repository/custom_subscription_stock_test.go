//go:build unit

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/usersubscription"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestCustomSubscriptionStockPostgres(t *testing.T) {
	dsn := os.Getenv("DEDICATED_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer admin.Close()
	name := fmt.Sprintf("manual_stock_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE DATABASE " + name)
	require.NoError(t, err)
	defer func() { _, e := admin.Exec("DROP DATABASE " + name); require.NoError(t, e) }()
	u, err := url.Parse(dsn)
	require.NoError(t, err)
	u.Path = "/" + name
	db, err := sql.Open("postgres", u.String())
	require.NoError(t, err)
	defer db.Close()
	ctx := context.Background()
	require.NoError(t, ApplyMigrations(ctx, db))
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	svc := service.NewSubscriptionService(NewGroupRepository(client, db), NewUserSubscriptionRepository(client), nil, client, nil)
	defer svc.Stop()
	group := client.Group.Create().SetName("manual stock").SetSubscriptionType("subscription").SaveX(ctx)
	plan := client.SubscriptionPlan.Create().SetName("limited").SetGroupID(group.ID).SetPrice(60).SetStockLimit(2).SaveX(ctx)
	user := func() *dbent.User {
		return client.User.Create().SetEmail(fmt.Sprintf("u%d@test.invalid", time.Now().UnixNano())).SetPasswordHash("test").SaveX(ctx)
	}
	used := func() int { return client.SubscriptionPlan.GetX(ctx, plan.ID).StockUsed }
	input := func(uid int64) *service.AssignSubscriptionInput {
		return &service.AssignSubscriptionInput{UserID: uid, GroupID: group.ID, PlanID: plan.ID, ValidityDays: 30}
	}
	u1 := user()
	sub, err := svc.AssignSubscription(ctx, input(u1.ID))
	require.NoError(t, err)
	require.Equal(t, 1, used())
	_, err = svc.AssignSubscription(ctx, input(u1.ID))
	require.NoError(t, err)
	require.Equal(t, 1, used(), "幂等复用不得重复扣库存")
	// 显式外层事务回滚时订阅和计数一起撤销。
	u2 := user()
	tx, err := client.Tx(ctx)
	require.NoError(t, err)
	_, err = svc.AssignSubscription(dbent.NewTxContext(ctx, tx), input(u2.ID))
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())
	require.Equal(t, 1, used())
	require.False(t, client.UserSubscription.Query().Where(usersubscription.UserIDEQ(u2.ID)).ExistX(ctx))
	_, err = svc.ExtendSubscription(ctx, sub.ID, 7)
	require.NoError(t, err)
	require.Equal(t, 2, used(), "续期根据账本找到原套餐")
	before := client.UserSubscription.GetX(ctx, sub.ID).ExpiresAt
	_, err = svc.ExtendSubscription(ctx, sub.ID, 7)
	require.Error(t, err)
	require.True(t, before.Equal(client.UserSubscription.GetX(ctx, sub.ID).ExpiresAt), "缺库存不得延长权益")
	_, err = svc.AssignSubscription(ctx, input(u2.ID))
	require.Error(t, err)
	require.False(t, client.UserSubscription.Query().Where(usersubscription.UserIDEQ(u2.ID)).ExistX(ctx))
	_, err = svc.ExtendSubscription(ctx, sub.ID, -1)
	require.NoError(t, err)
	require.Equal(t, 2, used())

	t.Run("批量部分成功与重试不重扣", func(t *testing.T) {
		client.SubscriptionPlan.UpdateOneID(plan.ID).SetStockLimit(3).ExecX(ctx)
		u3 := user()
		request := &service.BulkAssignSubscriptionInput{UserIDs: []int64{u2.ID, u3.ID}, GroupID: group.ID, PlanID: plan.ID, ValidityDays: 30}
		result, e := svc.BulkAssignSubscription(ctx, request)
		require.NoError(t, e)
		require.Equal(t, 1, result.SuccessCount)
		require.Equal(t, 1, result.FailedCount)
		require.Equal(t, 3, used())
		result, e = svc.BulkAssignSubscription(ctx, request)
		require.NoError(t, e)
		require.Equal(t, 1, result.ReusedCount)
		require.Equal(t, 1, result.FailedCount)
		require.Equal(t, 3, used())
	})
	t.Run("自动分配唯一套餐与歧义拒绝", func(t *testing.T) {
		client.SubscriptionPlan.UpdateOneID(plan.ID).SetStockLimit(10).ExecX(ctx)
		request := input(user().ID)
		request.PlanID = 0
		_, e := svc.AssignSubscription(ctx, request)
		require.NoError(t, e)
		other := client.SubscriptionPlan.Create().SetName("other").SetGroupID(group.ID).SetPrice(90).SaveX(ctx)
		request = input(user().ID)
		request.PlanID = 0
		_, e = svc.AssignSubscription(ctx, request)
		require.ErrorContains(t, e, "无法唯一确定")
		client.SubscriptionPlan.DeleteOneID(other.ID).ExecX(ctx)
	})
	t.Run("购买与管理员并发争抢最后一份", func(t *testing.T) {
		n := used()
		client.SubscriptionPlan.UpdateOneID(plan.ID).SetStockLimit(n + 1).ExecX(ctx)
		users := make([]int64, 12)
		for i := range users {
			users[i] = user().ID
		}
		var success atomic.Int32
		var wg sync.WaitGroup
		for i, uid := range users {
			wg.Add(1)
			go func(i int, uid int64) {
				defer wg.Done()
				var e error
				if i%2 == 0 {
					_, e = svc.AssignSubscription(ctx, input(uid))
				} else {
					_, e = client.PaymentOrder.Create().SetUserID(uid).SetUserEmail("test@test.invalid").SetUserName("test").SetAmount(60).SetPayAmount(60).SetRechargeCode(fmt.Sprintf("manual-race-%d", i)).SetPaymentType("test").SetPaymentTradeNo("").SetClientIP("127.0.0.1").SetSrcHost("test").SetOrderType("subscription").SetPlanID(plan.ID).SetSubscriptionGroupID(group.ID).SetStatus("PENDING").SetExpiresAt(time.Now().Add(30 * time.Minute)).Save(ctx)
				}
				if e == nil {
					success.Add(1)
				} else if !strings.Contains(e.Error(), "PLAN_OUT_OF_STOCK") {
					t.Errorf("unexpected race error: %v", e)
				}
			}(i, uid)
		}
		wg.Wait()
		require.Equal(t, int32(1), success.Load())
		require.Equal(t, n+1, used())
	})
	t.Run("历史人工归属回填且重跑不重扣", func(t *testing.T) {
		legacyGroup := client.Group.Create().SetName("legacy manual").SetSubscriptionType("subscription").SaveX(ctx)
		legacyPlan := client.SubscriptionPlan.Create().SetName("legacy").SetGroupID(legacyGroup.ID).SetPrice(60).SetStockLimit(0).SaveX(ctx)
		legacy := client.UserSubscription.Create().SetUserID(user().ID).SetGroupID(legacyGroup.ID).SetStartsAt(time.Now()).SetExpiresAt(time.Now().AddDate(0, 0, 30)).SaveX(ctx)
		ambiguousGroup := client.Group.Create().SetName("ambiguous history").SetSubscriptionType("subscription").SaveX(ctx)
		ambiguousPlan := client.SubscriptionPlan.Create().SetName("first").SetGroupID(ambiguousGroup.ID).SetPrice(60).SaveX(ctx)
		client.SubscriptionPlan.Create().SetName("second").SetGroupID(ambiguousGroup.ID).SetPrice(90).SaveX(ctx)
		client.UserSubscription.Create().SetUserID(user().ID).SetGroupID(ambiguousGroup.ID).SetStartsAt(time.Now()).SetExpiresAt(time.Now().AddDate(0, 0, 30)).SaveX(ctx)
		migration, e := os.ReadFile(filepath.Join("..", "..", "migrations", "246_custom_subscription_manual_stock.sql"))
		require.NoError(t, e)
		for range 2 {
			tx, e := db.BeginTx(ctx, nil)
			require.NoError(t, e)
			_, e = tx.Exec(string(migration))
			require.NoError(t, e)
			require.NoError(t, tx.Commit())
			require.Equal(t, 1, client.SubscriptionPlan.GetX(ctx, legacyPlan.ID).StockUsed)
			require.Zero(t, client.SubscriptionPlan.GetX(ctx, ambiguousPlan.ID).StockUsed, "多套餐历史不能猜测扣减")
		}
		var count int
		require.NoError(t, db.QueryRow("SELECT count(*) FROM custom_subscription_stock_allocations WHERE subscription_id=$1", legacy.ID).Scan(&count))
		require.Equal(t, 1, count)
	})
	t.Run("并发重复分配过期订阅只续期扣减一次", func(t *testing.T) {
		n := used()
		client.SubscriptionPlan.UpdateOneID(plan.ID).SetStockLimit(n + 1).ExecX(ctx)
		client.UserSubscription.UpdateOneID(sub.ID).SetStatus("expired").SetExpiresAt(time.Now().Add(-time.Hour)).ExecX(ctx)
		var wg sync.WaitGroup
		for range 12 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, e := svc.AssignSubscription(ctx, input(u1.ID))
				if e != nil {
					t.Errorf("duplicate assignment: %v", e)
				}
			}()
		}
		wg.Wait()
		require.Equal(t, n+1, used())
	})
}
