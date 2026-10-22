//go:build unit

package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

type customSequentialSubscriptionRepo struct {
	service.UserSubscriptionRepository
}

func TestCustomRedeemStockBindingFailureDoesNotLeaveCode(t *testing.T) {
	fixture := newCustomRedeemFixture(t, 1, -1)
	ctx := context.Background()
	plan := fixture.client.SubscriptionPlan.Query().OnlyX(ctx)
	_, err := fixture.client.ExecContext(ctx, "DROP TABLE custom_redeem_stock_sources")
	require.NoError(t, err)
	code := &service.RedeemCode{Code: "binding-failure", Type: service.RedeemTypeSubscription, Status: service.StatusUnused, GroupID: &fixture.groupID, ValidityDays: 7}
	err = fixture.redeem.CreateCodeWithStockPlan(ctx, code, plan.ID)
	require.Error(t, err)
	require.Zero(t, fixture.client.RedeemCode.Query().CountX(ctx))
	require.Zero(t, fixture.client.SubscriptionPlan.Query().OnlyX(ctx).StockUsed)
}

func TestCustomRedeemRejectsStockSourceForBalanceBatch(t *testing.T) {
	fixture := newCustomRedeemFixture(t, 1, -1)
	ctx := context.Background()
	plan := fixture.client.SubscriptionPlan.Query().OnlyX(ctx)
	code := &service.RedeemCode{Code: "balance-only", Type: service.RedeemTypeBalance, Status: service.StatusUnused, Value: 1}
	require.NoError(t, fixture.codes.Create(ctx, code))
	_, err := fixture.redeem.BatchUpdate(ctx, &service.RedeemCodeBatchUpdateInput{IDs: []int64{code.ID}, Fields: service.RedeemCodeBatchUpdateFields{PlanID: service.NullableInt64Update{Set: true, Value: &plan.ID}}})
	require.Equal(t, "SUBSCRIPTION_STOCK_PLAN_INVALID", infraerrors.Reason(err))
}

func (repo customSequentialSubscriptionRepo) GetByIDForUpdate(ctx context.Context, id int64) (*service.UserSubscription, error) {
	return repo.GetByID(ctx, id)
}

type customRedeemFixture struct {
	client        *dbent.Client
	userID        int64
	groupID       int64
	codes         service.RedeemCodeRepository
	users         service.UserRepository
	subscriptions *service.SubscriptionService
	redeem        *service.RedeemService
	admin         service.AdminService
}

func newCustomRedeemFixture(t *testing.T, plans int, stockLimit int) *customRedeemFixture {
	t.Helper()
	databaseName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	database, err := sql.Open("sqlite", "file:"+databaseName+"?mode=memory&cache=shared")
	require.NoError(t, err)
	database.SetMaxOpenConns(4)
	t.Cleanup(func() { _ = database.Close() })
	_, err = database.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(entsql.OpenDB(dialect.SQLite, database))))
	t.Cleanup(func() { _ = client.Close() })
	_, err = database.Exec("CREATE TABLE custom_subscription_stock_allocations (id INTEGER PRIMARY KEY AUTOINCREMENT, subscription_id BIGINT NOT NULL, plan_id BIGINT NOT NULL, user_id BIGINT NOT NULL, group_id BIGINT NOT NULL, source TEXT NOT NULL, source_key TEXT UNIQUE)")
	require.NoError(t, err)
	ctx := context.Background()
	user, err := client.User.Create().SetEmail("review@example.invalid").SetPasswordHash("test").Save(ctx)
	require.NoError(t, err)
	group, err := client.Group.Create().SetName("review subscriptions").SetSubscriptionType("subscription").Save(ctx)
	require.NoError(t, err)
	for index := 0; index < plans; index++ {
		_, err = client.SubscriptionPlan.Create().SetName(fmt.Sprintf("plan-%d", index)).SetGroupID(group.ID).SetPrice(10).SetStockLimit(stockLimit).Save(ctx)
		require.NoError(t, err)
	}
	_, err = database.Exec("CREATE TABLE custom_redeem_stock_sources (redeem_code_id BIGINT PRIMARY KEY REFERENCES redeem_codes(id) ON DELETE CASCADE, plan_id BIGINT NOT NULL CHECK (plan_id > 0))")
	require.NoError(t, err)
	codes := NewRedeemCodeRepository(client)
	users := NewUserRepository(client, database)
	subscriptions := service.NewSubscriptionService(NewGroupRepository(client, database), customSequentialSubscriptionRepo{NewUserSubscriptionRepository(client)}, nil, client, nil)
	t.Cleanup(subscriptions.Stop)
	redeem := service.NewRedeemService(codes, users, subscriptions, nil, nil, client, nil, nil)
	admin := service.NewAdminService(nil, nil, NewAdminGroupRepository(client, database), nil, nil, nil, codes, nil, nil, nil, nil, nil, nil, client, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	return &customRedeemFixture{client: client, userID: user.ID, groupID: group.ID, codes: codes, users: users, subscriptions: subscriptions, redeem: redeem, admin: admin}
}

func TestCustomRedeemGenerationRequiresAvailablePlan(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		plans  int
		limit  int
		reason string
	}{
		{"no_plan", 0, -1, "SUBSCRIPTION_STOCK_PLAN_REQUIRED"},
		{"multiple_plans", 2, -1, "SUBSCRIPTION_STOCK_PLAN_REQUIRED"},
		{"unique_available", 1, -1, ""},
		{"sold_out", 1, 0, "PLAN_OUT_OF_STOCK"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			fixture := newCustomRedeemFixture(t, scenario.plans, scenario.limit)
			ctx := context.Background()
			generated, err := fixture.admin.GenerateRedeemCodes(ctx, &service.GenerateRedeemCodesInput{Count: 1, Type: service.RedeemTypeSubscription, GroupID: &fixture.groupID, ValidityDays: 30})
			if scenario.reason != "" {
				require.Equal(t, scenario.reason, infraerrors.Reason(err))
				require.Empty(t, generated)
				require.Zero(t, fixture.client.RedeemCode.Query().CountX(ctx))
			} else {
				require.NoError(t, err)
				require.Len(t, generated, 1)
				plan := fixture.client.SubscriptionPlan.Query().OnlyX(ctx)
				require.Zero(t, plan.StockUsed, "generation must not reserve stock")
				source, err := fixture.codes.(service.CustomRedeemStockRepository).GetStockPlanID(ctx, generated[0].ID)
				require.NoError(t, err)
				require.Equal(t, plan.ID, source)
				used, err := fixture.redeem.Redeem(ctx, fixture.userID, generated[0].Code)
				require.NoError(t, err)
				require.Equal(t, service.StatusUsed, used.Status)
				require.NotNil(t, used.UsedAt)
				require.Equal(t, fixture.userID, *used.UsedBy)
			}
			balance := &service.RedeemCode{Code: "balance", Type: service.RedeemTypeBalance, Status: service.StatusUnused, Value: 2}
			require.NoError(t, fixture.codes.Create(ctx, balance))
			_, err = fixture.redeem.Redeem(ctx, fixture.userID, balance.Code)
			require.NoError(t, err)
		})
	}
}

func TestCustomRedeemExplicitPlanOverridesPreviousSource(t *testing.T) {
	fixture := newCustomRedeemFixture(t, 2, -1)
	ctx := context.Background()
	plans := fixture.client.SubscriptionPlan.Query().AllX(ctx)
	for _, plan := range plans {
		generated, err := fixture.admin.GenerateRedeemCodes(ctx, &service.GenerateRedeemCodesInput{Count: 1, Type: service.RedeemTypeSubscription, GroupID: &fixture.groupID, ValidityDays: 7, PlanID: plan.ID})
		require.NoError(t, err)
		_, err = fixture.redeem.Redeem(ctx, fixture.userID, generated[0].Code)
		require.NoError(t, err)
	}
	require.Equal(t, 1, fixture.client.UserSubscription.Query().CountX(ctx))
	for _, plan := range fixture.client.SubscriptionPlan.Query().AllX(ctx) {
		require.Equal(t, 1, plan.StockUsed)
	}
}

func TestCustomRedeemLegacyStockFailureRollsBack(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		plans  int
		limit  int
		reason string
	}{
		{"no_plan", 0, -1, "SUBSCRIPTION_STOCK_PLAN_REQUIRED"},
		{"multiple_plans", 2, -1, "SUBSCRIPTION_STOCK_PLAN_REQUIRED"},
		{"sold_out", 1, 0, "PLAN_OUT_OF_STOCK"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			fixture := newCustomRedeemFixture(t, scenario.plans, scenario.limit)
			ctx := context.Background()
			code := &service.RedeemCode{Code: "legacy", Type: service.RedeemTypeSubscription, Status: service.StatusUnused, GroupID: &fixture.groupID, ValidityDays: 7}
			require.NoError(t, fixture.codes.Create(ctx, code))
			_, err := fixture.redeem.Redeem(ctx, fixture.userID, code.Code)
			require.Equal(t, scenario.reason, infraerrors.Reason(err))
			stored, err := fixture.codes.GetByID(ctx, code.ID)
			require.NoError(t, err)
			require.Equal(t, service.StatusUnused, stored.Status)
			require.Nil(t, stored.UsedBy)
			require.Zero(t, fixture.client.UserSubscription.Query().CountX(ctx))
			for _, plan := range fixture.client.SubscriptionPlan.Query().AllX(ctx) {
				require.Zero(t, plan.StockUsed)
			}
		})
	}
}

func TestCustomRedeemBatchGroupValidationAndRepair(t *testing.T) {
	fixture := newCustomRedeemFixture(t, 2, -1)
	ctx := context.Background()
	code := &service.RedeemCode{Code: "legacy", Type: service.RedeemTypeSubscription, Status: service.StatusUnused, GroupID: &fixture.groupID, ValidityDays: 7}
	balance := &service.RedeemCode{Code: "balance", Type: service.RedeemTypeBalance, Status: service.StatusUnused, Value: 2}
	require.NoError(t, fixture.codes.Create(ctx, code))
	require.NoError(t, fixture.codes.Create(ctx, balance))
	note := "must not persist"
	_, err := fixture.redeem.BatchUpdate(ctx, &service.RedeemCodeBatchUpdateInput{IDs: []int64{code.ID, balance.ID}, Fields: service.RedeemCodeBatchUpdateFields{Notes: &note, GroupID: service.NullableInt64Update{Set: true}}})
	require.Equal(t, "REDEEM_CODE_GROUP_ID_REQUIRED", infraerrors.Reason(err))
	stored, err := fixture.codes.GetByID(ctx, code.ID)
	require.NoError(t, err)
	require.Equal(t, fixture.groupID, *stored.GroupID)
	require.Empty(t, stored.Notes)
	standard := fixture.client.Group.Create().SetName("standard").SetSubscriptionType("standard").SaveX(ctx)
	_, err = fixture.redeem.BatchUpdate(ctx, &service.RedeemCodeBatchUpdateInput{IDs: []int64{code.ID}, Fields: service.RedeemCodeBatchUpdateFields{GroupID: service.NullableInt64Update{Set: true, Value: &standard.ID}}})
	require.Equal(t, "GROUP_NOT_SUBSCRIPTION_TYPE", infraerrors.Reason(err))
	plan := fixture.client.SubscriptionPlan.Query().FirstX(ctx)
	updated, err := fixture.redeem.BatchUpdate(ctx, &service.RedeemCodeBatchUpdateInput{IDs: []int64{code.ID}, Fields: service.RedeemCodeBatchUpdateFields{GroupID: service.NullableInt64Update{Set: true, Value: &fixture.groupID}, PlanID: service.NullableInt64Update{Set: true, Value: &plan.ID}}})
	require.NoError(t, err)
	require.Equal(t, int64(1), updated.Updated)
	_, err = fixture.redeem.Redeem(ctx, fixture.userID, code.Code)
	require.NoError(t, err)
	_, err = fixture.redeem.BatchUpdate(ctx, &service.RedeemCodeBatchUpdateInput{IDs: []int64{code.ID}, Fields: service.RedeemCodeBatchUpdateFields{PlanID: service.NullableInt64Update{Set: true, Value: &plan.ID}}})
	require.ErrorIs(t, err, service.ErrRedeemCodeUsed)
}

func TestCustomRedeemBatchRepairsMissingGroup(t *testing.T) {
	fixture := newCustomRedeemFixture(t, 2, -1)
	ctx := context.Background()
	code := &service.RedeemCode{Code: "missing-group", Type: service.RedeemTypeSubscription, Status: service.StatusUnused, ValidityDays: 7}
	require.NoError(t, fixture.codes.Create(ctx, code))
	plan := fixture.client.SubscriptionPlan.Query().FirstX(ctx)
	_, err := fixture.redeem.BatchUpdate(ctx, &service.RedeemCodeBatchUpdateInput{IDs: []int64{code.ID}, Fields: service.RedeemCodeBatchUpdateFields{GroupID: service.NullableInt64Update{Set: true, Value: &fixture.groupID}, PlanID: service.NullableInt64Update{Set: true, Value: &plan.ID}}})
	require.NoError(t, err)
	_, err = fixture.redeem.Redeem(ctx, fixture.userID, code.Code)
	require.NoError(t, err)
}

type customRedeemReadRepo struct {
	service.RedeemCodeRepository
	failInside    bool
	externalReads int
}

func (repo *customRedeemReadRepo) GetByID(ctx context.Context, id int64) (*service.RedeemCode, error) {
	if dbent.TxFromContext(ctx) == nil {
		repo.externalReads++
		return nil, errors.New("outside transaction read failed")
	}
	if repo.failInside {
		return nil, errors.New("transaction snapshot read failed")
	}
	return repo.RedeemCodeRepository.GetByID(ctx, id)
}

func TestCustomRedeemSnapshotReadIsTransactional(t *testing.T) {
	for _, failInside := range []bool{false, true} {
		t.Run(fmt.Sprint(failInside), func(t *testing.T) {
			fixture := newCustomRedeemFixture(t, 1, -1)
			ctx := context.Background()
			code := &service.RedeemCode{Code: "snapshot", Type: service.RedeemTypeSubscription, Status: service.StatusUnused, GroupID: &fixture.groupID, ValidityDays: 7}
			require.NoError(t, fixture.codes.Create(ctx, code))
			repo := &customRedeemReadRepo{RedeemCodeRepository: fixture.codes, failInside: failInside}
			redeem := service.NewRedeemService(repo, fixture.users, fixture.subscriptions, nil, nil, fixture.client, nil, nil)
			result, err := redeem.Redeem(ctx, fixture.userID, code.Code)
			stored, readErr := fixture.codes.GetByID(ctx, code.ID)
			require.NoError(t, readErr)
			require.Zero(t, repo.externalReads)
			if failInside {
				require.ErrorContains(t, err, "snapshot read failed")
				require.Nil(t, result)
				require.Equal(t, service.StatusUnused, stored.Status)
				require.Zero(t, fixture.client.UserSubscription.Query().CountX(ctx))
				require.Zero(t, fixture.client.SubscriptionPlan.Query().OnlyX(ctx).StockUsed)
			} else {
				require.NoError(t, err)
				require.Equal(t, service.StatusUsed, result.Status)
				require.Equal(t, stored.UsedAt, result.UsedAt)
				require.Equal(t, 1, fixture.client.UserSubscription.Query().CountX(ctx))
			}
		})
	}
}

func TestCustomRedeemFixedCodeAcceptsZeroValueAndValidatesPlan(t *testing.T) {
	fixture := newCustomRedeemFixture(t, 2, -1)
	ctx := context.Background()
	plan := fixture.client.SubscriptionPlan.Query().FirstX(ctx)
	code := &service.RedeemCode{Code: "fixed", Type: service.RedeemTypeSubscription, Status: service.StatusUnused, GroupID: &fixture.groupID, ValidityDays: 7}
	require.NoError(t, fixture.redeem.CreateCodeWithStockPlan(ctx, code, plan.ID))
	_, err := fixture.redeem.Redeem(ctx, fixture.userID, code.Code)
	require.NoError(t, err)
	other := fixture.client.Group.Create().SetName("other").SetSubscriptionType("subscription").SaveX(ctx)
	invalid := &service.RedeemCode{Code: "invalid", Type: service.RedeemTypeSubscription, GroupID: &other.ID, ValidityDays: 7}
	require.Equal(t, "SUBSCRIPTION_STOCK_PLAN_INVALID", infraerrors.Reason(fixture.redeem.CreateCodeWithStockPlan(ctx, invalid, plan.ID)))
	require.Equal(t, 1, fixture.client.RedeemCode.Query().CountX(ctx))
}
