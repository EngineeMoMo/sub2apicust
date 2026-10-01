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
	s := NewCustomDedicatedService(db, nil, keys, &GatewayService{}, &OpenAIGatewayService{})
	ctx := context.Background()
	newKey := func(user, id, group int64) *APIKey {
		key, err := keys.PrepareCustomDedicatedBilling(ctx, &APIKey{ID: id, UserID: user, GroupID: &group, User: &User{ID: user}})
		require.NoError(t, err)
		return key
	}
	reset := func() {
		_, err := db.Exec("DELETE FROM custom_dedicated_request_leases; DELETE FROM custom_dedicated_request_counters; DELETE FROM custom_dedicated_billing_policies; UPDATE custom_dedicated_accounts SET revoked_at=NULL,expires_at=NOW()+INTERVAL '1 day'; UPDATE users SET status='active'; DELETE FROM user_allowed_groups; INSERT INTO user_allowed_groups VALUES(11,33),(12,33),(11,55),(12,55),(13,77)")
		require.NoError(t, err)
	}
	t.Run("positive_membership_only", func(t *testing.T) {
		require.True(t, newKey(11, 7, 33).IsCustomDedicatedPrepaid())
		require.True(t, newKey(12, 8, 55).IsCustomDedicatedPrepaid())
		require.False(t, newKey(13, 9, 77).IsCustomDedicatedPrepaid())
		group := int64(33)
		_, err := keys.PrepareCustomDedicatedBilling(ctx, &APIKey{ID: 9, UserID: 13, GroupID: &group})
		require.ErrorIs(t, err, ErrDedicatedAccess)
	})
	t.Run("concurrent_members_groups_and_keys_share_slots", func(t *testing.T) {
		reset()
		var accepted atomic.Int32
		var unexpected atomic.Int32
		var wg sync.WaitGroup
		for i := 0; i < 12; i++ {
			key := newKey(11, 7, 33)
			if i%2 == 1 {
				key = newKey(12, 8, 55)
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := s.AdmitBillingRequest(ctx, key, "/v1/responses", []byte(`{"model":"gpt-test"}`), false, false)
				if err == nil {
					accepted.Add(1)
				} else if err != ErrDedicatedLimit {
					unexpected.Add(1)
				}
			}()
		}
		wg.Wait()
		require.Zero(t, unexpected.Load())
		require.Equal(t, int32(2), accepted.Load())
		var count int
		require.NoError(t, db.QueryRow("SELECT day_requests FROM custom_dedicated_request_counters WHERE binding_id=1").Scan(&count))
		require.Equal(t, 2, count)
	})
	t.Run("hard_daily_admission_and_renewal_does_not_reset", func(t *testing.T) {
		reset()
		p := defaultCustomDedicatedPolicy()
		p.DailyRequestLimit = 2
		_, err := s.UpdateBillingPolicy(ctx, 1, p)
		require.NoError(t, err)
		for i := 0; i < 2; i++ {
			key := newKey(11, 7, 33)
			require.NoError(t, s.AdmitBillingRequest(ctx, key, "/responses", []byte(`{}`), false, false))
			_, err = db.Exec("UPDATE custom_dedicated_request_leases SET finished_at=NOW()")
			require.NoError(t, err)
		}
		_, err = db.Exec("UPDATE custom_dedicated_accounts SET expires_at=NOW()+INTERVAL '2 days',updated_at=NOW()")
		require.NoError(t, err)
		require.ErrorIs(t, s.AdmitBillingRequest(ctx, newKey(12, 8, 55), "/responses", []byte(`{}`), false, false), ErrDedicatedLimit)
		_, err = db.Exec("UPDATE custom_dedicated_request_counters SET day_start=day_start-INTERVAL '1 day'")
		require.NoError(t, err)
		require.NoError(t, s.AdmitBillingRequest(ctx, newKey(12, 8, 55), "/responses", []byte(`{}`), false, false))
	})
	t.Run("rpm_all_aliases_share_one_counter", func(t *testing.T) {
		reset()
		p := defaultCustomDedicatedPolicy()
		p.RPMLimit = 1
		_, err := s.UpdateBillingPolicy(ctx, 1, p)
		require.NoError(t, err)
		require.NoError(t, s.AdmitBillingRequest(ctx, newKey(11, 7, 33), "/v1/messages", []byte(`{}`), false, false))
		require.ErrorIs(t, s.AdmitBillingRequest(ctx, newKey(12, 8, 55), "/backend-api/codex/responses", []byte(`{}`), false, false), ErrDedicatedLimit)
	})
	t.Run("concurrent_last_daily_request_is_consumed_once", func(t *testing.T) {
		reset()
		p := defaultCustomDedicatedPolicy()
		p.ConcurrencyLimit = 20
		p.DailyRequestLimit = 1
		_, err := s.UpdateBillingPolicy(ctx, 1, p)
		require.NoError(t, err)
		var accepted, unexpected atomic.Int32
		var wg sync.WaitGroup
		for i := 0; i < 12; i++ {
			key := newKey(11, 7, 33)
			if i%2 == 1 {
				key = newKey(12, 8, 55)
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := s.AdmitBillingRequest(ctx, key, "/responses", []byte(`{}`), false, false)
				if err == nil {
					accepted.Add(1)
				} else if err != ErrDedicatedLimit {
					unexpected.Add(1)
				}
			}()
		}
		wg.Wait()
		require.Zero(t, unexpected.Load())
		require.Equal(t, int32(1), accepted.Load())
		var count int
		require.NoError(t, db.QueryRow("SELECT day_requests FROM custom_dedicated_request_counters WHERE binding_id=1").Scan(&count))
		require.Equal(t, 1, count)
	})
	t.Run("websocket_every_turn_and_live_policy", func(t *testing.T) {
		reset()
		key := newKey(11, 7, 33)
		require.NoError(t, s.AdmitBillingRequest(ctx, key, "/responses", nil, true, false))
		var count int
		require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM custom_dedicated_request_counters").Scan(&count))
		require.Zero(t, count)
		p := defaultCustomDedicatedPolicy()
		p.DailyRequestLimit = 1
		_, err := s.UpdateBillingPolicy(ctx, 1, p)
		require.NoError(t, err)
		require.NoError(t, s.AdmitBillingRequest(ctx, key, "/responses", []byte(`{"type":"response.create"}`), true, true))
		require.ErrorIs(t, s.AdmitBillingRequest(ctx, key, "/responses", []byte(`{"type":"response.create"}`), true, true), ErrDedicatedLimit)
	})
	t.Run("revoke_expire_remove_and_changed_account_reject_next_request", func(t *testing.T) {
		for _, change := range []string{"UPDATE custom_dedicated_accounts SET revoked_at=NOW()", "UPDATE custom_dedicated_accounts SET expires_at=NOW()-INTERVAL '1 second'", "UPDATE custom_dedicated_accounts SET user_ids='[12]'", "DELETE FROM user_allowed_groups WHERE user_id=11", "UPDATE groups SET fallback_group_id=77 WHERE id=33"} {
			reset()
			key := newKey(11, 7, 33)
			_, err := db.Exec(change)
			require.NoError(t, err)
			require.ErrorIs(t, s.AdmitBillingRequest(ctx, key, "/responses", []byte(`{}`), false, false), ErrDedicatedAccess)
			_, err = db.Exec("UPDATE custom_dedicated_accounts SET user_ids='[11,12]'; UPDATE groups SET fallback_group_id=NULL")
			require.NoError(t, err)
		}
	})
	t.Run("lease_expiry_recovery_and_finished_receipt_retention", func(t *testing.T) {
		reset()
		for i := 0; i < 2; i++ {
			require.NoError(t, s.AdmitBillingRequest(ctx, newKey(11, 7, 33), "/responses", []byte(`{}`), false, false))
		}
		require.ErrorIs(t, s.AdmitBillingRequest(ctx, newKey(12, 8, 55), "/responses", []byte(`{}`), false, false), ErrDedicatedLimit)
		_, err := db.Exec("UPDATE custom_dedicated_request_leases SET expires_at=NOW()-INTERVAL '1 second'")
		require.NoError(t, err)
		require.NoError(t, s.AdmitBillingRequest(ctx, newKey(12, 8, 55), "/responses", []byte(`{}`), false, false))
		var receipts int
		require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM custom_dedicated_request_leases").Scan(&receipts))
		require.Equal(t, 3, receipts)
	})
	t.Run("policy_stale_update_and_images_cannot_bypass", func(t *testing.T) {
		reset()
		p := defaultCustomDedicatedPolicy()
		updated, err := s.UpdateBillingPolicy(ctx, 1, p)
		require.NoError(t, err)
		_, err = s.UpdateBillingPolicy(ctx, 1, p)
		require.ErrorIs(t, err, ErrDedicatedStale)
		key := newKey(11, 7, 33)
		require.ErrorIs(t, s.AdmitBillingRequest(ctx, key, "/responses", []byte(`{"tools":[{"type":"image_generation"}],"tools":[]}`), false, false), ErrDedicatedEndpoint)
		updated.AllowImages = true
		_, err = s.UpdateBillingPolicy(ctx, 1, updated)
		require.NoError(t, err)
		require.NoError(t, s.AdmitBillingRequest(ctx, key, "/v1/images/generations", []byte(`{"model":"test"}`), false, false))
		require.NoError(t, s.AdmitBillingRequest(ctx, newKey(12, 8, 55), "/v1/images/edits", []byte("--test\r\nContent-Disposition: form-data; name=\"model\"\r\n\r\ngpt-image-1\r\n--test--\r\n"), false, false, "multipart/form-data; boundary=test"))
	})
}
