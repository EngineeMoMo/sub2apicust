//go:build unit

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func dedicatedBillingRepositoryPG(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DEDICATED_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires an isolated PostgreSQL database")
	}
	u, err := url.Parse(dsn)
	require.NoError(t, err)
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	schema := fmt.Sprintf("dedicated_settlement_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	db, err := sql.Open("postgres", u.String())
	require.NoError(t, err)
	db.SetMaxOpenConns(20)
	t.Cleanup(func() {
		_ = db.Close()
		_, dropErr := admin.Exec("DROP SCHEMA " + schema + " CASCADE")
		require.NoError(t, dropErr)
		_ = admin.Close()
	})
	for _, file := range []string{"../service/testdata/custom_dedicated_billing.sql", "../../migrations/241_custom_dedicated_accounts.sql", "../../migrations/242_custom_dedicated_members.sql", "../../migrations/243_custom_dedicated_billing.sql"} {
		body, err := os.ReadFile(file)
		require.NoError(t, err)
		_, err = db.Exec(string(body))
		require.NoError(t, err)
	}
	_, err = db.Exec(`INSERT INTO custom_dedicated_accounts(user_id,user_ids,account_id,group_id,label,expires_at) VALUES(11,'[11,12]',22,33,'结算测试',NOW()+INTERVAL '1 day')`)
	require.NoError(t, err)
	return db
}

func TestCustomDedicatedBillingRepositoryPostgres(t *testing.T) {
	db := dedicatedBillingRepositoryPG(t)
	keys := &service.APIKeyService{}
	service.NewCustomDedicatedService(db, nil, keys, &service.GatewayService{}, &service.OpenAIGatewayService{})
	r := NewUsageBillingRepository(nil, db)
	ctx := context.Background()
	admit := func(ws bool) *service.UsageBillingCommand {
		group := int64(33)
		key, err := keys.PrepareCustomDedicatedBilling(ctx, &service.APIKey{ID: 7, UserID: 11, GroupID: &group, User: &service.User{ID: 11}})
		require.NoError(t, err)
		require.NoError(t, keys.AdmitCustomDedicatedRequest(ctx, key, ws))
		requestID := key.CustomDedicatedUsageRequestID("upstream-turn-1")
		leaseID := strings.Split(requestID, ":")[1]
		return &service.UsageBillingCommand{RequestID: requestID, APIKeyID: 7, UserID: 11, AccountID: 22, AccountType: service.AccountTypeOAuth, Model: "gpt-test", BillingType: service.BillingTypeDedicated, InputTokens: 100, OutputTokens: 50, APIKeyQuotaCost: 2, APIKeyRateLimitCost: 2, DedicatedBindingID: 1, DedicatedGroupID: 33, DedicatedLeaseID: leaseID, DedicatedReferenceCost: 2}
	}
	reset := func() {
		_, err := db.Exec(`DELETE FROM custom_dedicated_billing_usage; DELETE FROM usage_billing_dedup; DELETE FROM custom_dedicated_request_leases; DELETE FROM custom_dedicated_request_counters; UPDATE users SET balance=0; UPDATE api_keys SET quota_used=0,usage_5h=0,usage_1d=0,usage_7d=0; UPDATE custom_dedicated_accounts SET revoked_at=NULL,expires_at=NOW()+INTERVAL '1 day',user_ids='[11,12]',account_id=22`)
		require.NoError(t, err)
	}
	state := func(balance, meter float64, records, claims int) {
		t.Helper()
		var gotBalance, quota, h5, d1, d7, reference float64
		require.NoError(t, db.QueryRow("SELECT balance FROM users WHERE id=11").Scan(&gotBalance))
		require.NoError(t, db.QueryRow("SELECT quota_used,usage_5h,usage_1d,usage_7d FROM api_keys WHERE id=7").Scan(&quota, &h5, &d1, &d7))
		var gotRecords, gotClaims int
		require.NoError(t, db.QueryRow("SELECT COUNT(*),COALESCE(SUM(reference_cost),0) FROM custom_dedicated_billing_usage").Scan(&gotRecords, &reference))
		require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM usage_billing_dedup").Scan(&gotClaims))
		require.Equal(t, balance, gotBalance)
		for _, value := range []float64{quota, h5, d1, d7, reference} {
			require.Equal(t, meter, value)
		}
		require.Equal(t, records, gotRecords)
		require.Equal(t, claims, gotClaims)
	}
	t.Run("zero_balance_kept_reference_and_key_limits_increment", func(t *testing.T) {
		reset()
		result, err := r.Apply(ctx, admit(false))
		require.NoError(t, err)
		require.True(t, result.Applied)
		require.Nil(t, result.NewBalance)
		state(0, 2, 1, 1)
	})
	t.Run("same_callback_concurrent_only_once", func(t *testing.T) {
		reset()
		cmd := *admit(false)
		var applied, failures atomic.Int32
		var wg sync.WaitGroup
		for i := 0; i < 12; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				copyCmd := cmd
				result, err := r.Apply(ctx, &copyCmd)
				if err != nil {
					failures.Add(1)
				} else if result.Applied {
					applied.Add(1)
				}
			}()
		}
		wg.Wait()
		require.Zero(t, failures.Load())
		require.Equal(t, int32(1), applied.Load())
		state(0, 2, 1, 1)
	})
	t.Run("repeated_client_id_cannot_dedup_independent_execution", func(t *testing.T) {
		reset()
		one, two := admit(false), admit(false)
		require.NotEqual(t, one.RequestID, two.RequestID)
		for _, cmd := range []*service.UsageBillingCommand{one, two} {
			result, err := r.Apply(ctx, cmd)
			require.NoError(t, err)
			require.True(t, result.Applied)
		}
		state(0, 4, 2, 2)
	})
	t.Run("websocket_distinct_turns_and_repeated_callback", func(t *testing.T) {
		reset()
		one := admit(true)
		two := *one
		two.RequestID = "dedicated:" + one.DedicatedLeaseID + ":" + strings.Repeat("b", 64)
		for _, cmd := range []*service.UsageBillingCommand{one, &two, one} {
			_, err := r.Apply(ctx, cmd)
			require.NoError(t, err)
		}
		state(0, 4, 2, 2)
	})
	t.Run("accepted_receipt_remains_prepaid_after_revoke_expire_or_transfer", func(t *testing.T) {
		for _, change := range []string{"revoked_at=NOW()", "expires_at=NOW()-INTERVAL '1 day'", "user_ids='[12]'", "account_id=44"} {
			reset()
			cmd := admit(false)
			_, err := db.Exec("UPDATE custom_dedicated_accounts SET " + change)
			require.NoError(t, err)
			_, err = db.Exec("UPDATE custom_dedicated_request_leases SET finished_at=NOW(),expires_at=NOW()-INTERVAL '1 second'")
			require.NoError(t, err)
			result, err := r.Apply(ctx, cmd)
			require.NoError(t, err)
			require.True(t, result.Applied)
			state(0, 2, 1, 1)
		}
	})
	t.Run("forged_receipt_charge_or_reference_rejected_transactionally", func(t *testing.T) {
		mutations := []func(*service.UsageBillingCommand){
			func(c *service.UsageBillingCommand) { c.DedicatedLeaseID = "forged" },
			func(c *service.UsageBillingCommand) { c.AccountID = 44 },
			func(c *service.UsageBillingCommand) { c.UserID = 12 },
			func(c *service.UsageBillingCommand) { c.DedicatedGroupID = 55 },
			func(c *service.UsageBillingCommand) { c.BalanceCost = 1 },
			func(c *service.UsageBillingCommand) { c.BillingType = 0 },
			func(c *service.UsageBillingCommand) { c.RequestID = "client-id" },
			func(c *service.UsageBillingCommand) { c.DedicatedReferenceCost = -1 },
		}
		for _, mutate := range mutations {
			reset()
			cmd := admit(false)
			mutate(cmd)
			_, err := r.Apply(ctx, cmd)
			require.ErrorIs(t, err, service.ErrDedicatedAccess)
			state(0, 0, 0, 0)
		}
	})
	t.Run("ledger_error_rolls_back_dedup_and_key_effects", func(t *testing.T) {
		reset()
		cmd := admit(false)
		_, err := db.Exec("ALTER TABLE custom_dedicated_billing_usage ADD CONSTRAINT test_ledger_failure CHECK (reference_cost=0)")
		require.NoError(t, err)
		_, err = r.Apply(ctx, cmd)
		require.Error(t, err)
		state(0, 0, 0, 0)
		_, err = db.Exec("ALTER TABLE custom_dedicated_billing_usage DROP CONSTRAINT test_ledger_failure")
		require.NoError(t, err)
		result, err := r.Apply(ctx, cmd)
		require.NoError(t, err)
		require.True(t, result.Applied)
		state(0, 2, 1, 1)
	})
	t.Run("ordinary_group_retains_balance_billing", func(t *testing.T) {
		reset()
		_, err := db.Exec("UPDATE users SET balance=10 WHERE id=11")
		require.NoError(t, err)
		result, err := r.Apply(ctx, &service.UsageBillingCommand{RequestID: "ordinary-execution", APIKeyID: 7, UserID: 11, AccountID: 44, Model: "gpt-test", BalanceCost: 3})
		require.NoError(t, err)
		require.True(t, result.Applied)
		require.Equal(t, 7.0, *result.NewBalance)
		state(7, 0, 0, 1)
	})
}
