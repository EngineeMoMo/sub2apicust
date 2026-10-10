//go:build unit

package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func customDedicatedBillingPG(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DEDICATED_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires an isolated PostgreSQL database")
	}
	u, err := url.Parse(dsn)
	require.NoError(t, err)
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	schema := fmt.Sprintf("dedicated_billing_%d", time.Now().UnixNano())
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
	fixture, err := os.ReadFile("testdata/custom_dedicated_billing.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(fixture))
	require.NoError(t, err)
	for _, file := range []string{"241_custom_dedicated_accounts.sql", "242_custom_dedicated_members.sql", "243_custom_dedicated_billing.sql"} {
		body, err := os.ReadFile("../../migrations/" + file)
		require.NoError(t, err)
		_, err = db.Exec(string(body))
		require.NoError(t, err)
	}
	_, err = db.Exec(`INSERT INTO custom_dedicated_accounts(user_id,user_ids,account_id,group_id,label,expires_at) VALUES(11,'[11,12]',22,33,'共享测试包号',NOW()+INTERVAL '1 day')`)
	require.NoError(t, err)
	return db
}

func TestCustomDedicatedBillingPostgres(t *testing.T) {
	db := customDedicatedBillingPG(t)
	keys := &APIKeyService{}
	svc := NewCustomDedicatedService(db, nil, keys, &GatewayService{}, &OpenAIGatewayService{})
	ctx := context.Background()
	newKey := func(user, id, group int64) *APIKey {
		key, err := keys.PrepareCustomDedicatedBilling(ctx, &APIKey{ID: id, UserID: user, GroupID: &group, User: &User{ID: user}})
		require.NoError(t, err)
		return key
	}
	t.Run("positive_membership_only", func(t *testing.T) {
		require.True(t, newKey(11, 7, 33).IsCustomDedicatedPrepaid())
		require.True(t, newKey(12, 8, 55).IsCustomDedicatedPrepaid())
		require.False(t, newKey(13, 9, 77).IsCustomDedicatedPrepaid())
		group := int64(33)
		_, err := keys.PrepareCustomDedicatedBilling(ctx, &APIKey{ID: 9, UserID: 13, GroupID: &group})
		require.ErrorIs(t, err, ErrDedicatedAccess)
	})
	t.Run("old_limits_and_counters_no_longer_control_admission", func(t *testing.T) {
		_, err := db.Exec(`INSERT INTO custom_dedicated_billing_policies(binding_id,concurrency_limit,rpm_limit,daily_request_limit,max_body_bytes,allow_images) VALUES(1,1,1,1,1024,FALSE);
INSERT INTO custom_dedicated_request_counters(binding_id,minute_start,minute_requests,day_start,day_requests) VALUES(1,date_trunc('minute',NOW()),1000000,date_trunc('day',NOW()),1000000)`)
		require.NoError(t, err)
		var accepted atomic.Int32
		var workers sync.WaitGroup
		for index := 0; index < 12; index++ {
			key := newKey(11, 7, 33)
			if index%2 == 1 {
				key = newKey(12, 8, 55)
			}
			workers.Add(1)
			go func() {
				defer workers.Done()
				if svc.AdmitBillingRequest(ctx, key, false) == nil {
					accepted.Add(1)
				}
			}()
		}
		workers.Wait()
		require.Equal(t, int32(12), accepted.Load())
		for index := 0; index < 31; index++ {
			require.NoError(t, svc.AdmitBillingRequest(ctx, newKey(11, 7, 33), false))
		}
		var active, receipts, requests int
		require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM custom_dedicated_request_leases WHERE finished_at IS NULL").Scan(&active))
		require.Zero(t, active)
		require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM custom_dedicated_request_leases").Scan(&receipts))
		require.Equal(t, 43, receipts)
		require.NoError(t, db.QueryRow("SELECT day_requests FROM custom_dedicated_request_counters WHERE binding_id=1").Scan(&requests))
		require.Equal(t, 1000000, requests)
	})
	t.Run("websocket_receipt_keeps_distinct_turn_identity", func(t *testing.T) {
		key := newKey(11, 7, 33)
		require.NoError(t, svc.AdmitBillingRequest(ctx, key, true))
		require.NotEmpty(t, key.CustomDedicatedUsageRequestID("turn-1"))
		require.NotEqual(t, key.CustomDedicatedUsageRequestID("turn-1"), key.CustomDedicatedUsageRequestID("turn-2"))
	})
	t.Run("revoke_expire_remove_and_changed_account_reject_next_request", func(t *testing.T) {
		for _, change := range []string{"UPDATE custom_dedicated_accounts SET revoked_at=NOW()", "UPDATE custom_dedicated_accounts SET expires_at=NOW()-INTERVAL '1 second'", "UPDATE custom_dedicated_accounts SET user_ids='[12]'", "DELETE FROM user_allowed_groups WHERE user_id=11", "UPDATE groups SET fallback_group_id=77 WHERE id=33"} {
			key := newKey(11, 7, 33)
			_, err := db.Exec(change)
			require.NoError(t, err)
			require.ErrorIs(t, svc.AdmitBillingRequest(ctx, key, false), ErrDedicatedAccess)
			_, err = db.Exec("UPDATE custom_dedicated_accounts SET user_ids='[11,12]',revoked_at=NULL,expires_at=NOW()+INTERVAL '1 day'; UPDATE groups SET fallback_group_id=NULL; DELETE FROM user_allowed_groups; INSERT INTO user_allowed_groups VALUES(11,33),(12,33),(11,55),(12,55),(13,77)")
			require.NoError(t, err)
		}
	})
}
