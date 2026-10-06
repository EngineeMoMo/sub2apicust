//go:build unit

package routes

import (
	"context"
	"database/sql"
	"strconv"
	"sync"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/ent/paymentauditlog"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

// 使用真实仓储和隔离 SQLite，检查金额拒绝与并发通知下的实际余额。
func TestCustomAlipaySecurityExactAmountAndConcurrentCredit(t *testing.T) {
	db, err := sql.Open("sqlite", "file:"+t.Name()+"?mode=memory&cache=shared")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(entsql.OpenDB(dialect.SQLite, db))))
	t.Cleanup(func() { _ = client.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	u, err := client.User.Create().SetEmail("credit@example.test").SetUsername("credit").SetPasswordHash("test").Save(ctx)
	require.NoError(t, err)
	o, err := client.PaymentOrder.Create().SetUserID(u.ID).SetUserEmail(u.Email).SetUserName(u.Username).
		SetAmount(10).SetPayAmount(10).SetFeeRate(0).SetRechargeCode("ALIPAY-CONCURRENT").SetOutTradeNo("alipay-concurrent").
		SetPaymentType(payment.TypeAlipay).SetProviderKey(payment.TypeAlipay).SetPaymentTradeNo("").SetClientIP("127.0.0.1").SetSrcHost("merchant.example").
		SetOrderType(payment.OrderTypeBalance).SetStatus(service.OrderStatusPending).SetExpiresAt(time.Now().Add(time.Hour)).Save(ctx)
	require.NoError(t, err)
	users := repository.NewUserRepository(client, db)
	redeem := service.NewRedeemService(repository.NewRedeemCodeRepository(client), users, nil, nil, nil, client, nil, nil)
	svc := service.NewPaymentService(client, payment.NewRegistry(), nil, redeem, nil, nil, users, nil, nil)
	notify := func(amount float64) error {
		return svc.HandlePaymentNotification(ctx, &payment.PaymentNotification{OrderID: o.OutTradeNo, TradeNo: "official-trade", Amount: amount, Status: payment.NotificationStatusSuccess}, payment.TypeAlipay)
	}
	for _, amount := range []float64{9.99, 10.01, 10.001} {
		require.Error(t, notify(amount), "支付宝不能容忍少付、多付或小数分")
	}
	fresh, err := client.User.Get(ctx, u.ID)
	require.NoError(t, err)
	require.Zero(t, fresh.Balance)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); <-start; _ = notify(10) }()
	}
	close(start)
	wg.Wait()
	// 有租约冲突的回调可重试，最终重复通知均应成功且不再加钱。
	for i := 0; i < 3; i++ {
		require.NoError(t, notify(10))
	}
	fresh, err = client.User.Get(ctx, u.ID)
	require.NoError(t, err)
	require.Equal(t, float64(10), fresh.Balance)
	order, err := client.PaymentOrder.Get(ctx, o.ID)
	require.NoError(t, err)
	require.Equal(t, service.OrderStatusCompleted, order.Status)
	count, err := client.RedeemCode.Query().Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	count, err = client.PaymentAuditLog.Query().Where(paymentauditlog.OrderIDEQ(strconv.FormatInt(o.ID, 10)), paymentauditlog.ActionEQ("RECHARGE_SUCCESS")).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}
