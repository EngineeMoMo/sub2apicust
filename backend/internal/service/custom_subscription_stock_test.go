package service

import (
	"context"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/DATA-DOG/go-sqlmock"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/stretchr/testify/require"
)

func TestCustomSubscriptionStockPaidOrderDoesNotChargeAgain(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	defer func() { _ = client.Close() }()
	svc := &SubscriptionService{entClient: client}
	planID := int64(42)
	order := &dbent.PaymentOrder{ID: 7, PlanID: &planID}
	for range 2 {
		mock.ExpectBegin()
		// 唯一账本键防重放；不得读现售套餐或再 UPDATE 库存。
		mock.ExpectExec("INSERT INTO custom_subscription_stock_allocations").WithArgs(int64(1), int64(42), int64(3), int64(4), "payment", "order:7").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err = svc.withSubscriptionUpdateTx(context.Background(), func(ctx context.Context) error {
			return svc.customConsumeSubscriptionStock(customWithPaidSubscriptionStock(ctx, order), 1, 3, 4)
		})
		require.NoError(t, err)
	}
	require.NoError(t, mock.ExpectationsWereMet())
}
