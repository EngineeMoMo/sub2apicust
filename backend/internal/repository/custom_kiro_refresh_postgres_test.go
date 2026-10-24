//go:build unit

package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestCustomKiroRefreshPostgresCAS(t *testing.T) {
	dsn := os.Getenv("CUSTOM_KIRO_TEST_DSN")
	if dsn == "" {
		if os.Getenv("CI") != "" {
			t.Fatal("CI必须配置CUSTOM_KIRO_TEST_DSN以验证Kiro凭据CAS")
		}
		t.Skip("CUSTOM_KIRO_TEST_DSN 未配置，未执行真实 PostgreSQL CAS 回归")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	db.SetMaxOpenConns(1)
	ctx := context.Background()
	var database string
	require.NoError(t, db.QueryRowContext(ctx, "SELECT current_database()").Scan(&database))
	require.Equal(t, "kiro_native_test", database, "仅允许独立测试库")
	schema := fmt.Sprintf("kiro_cas_%d", time.Now().UnixNano())
	_, err = db.ExecContext(ctx, "CREATE SCHEMA "+schema)
	require.NoError(t, err)
	defer func() { _, _ = db.ExecContext(context.Background(), "DROP SCHEMA "+schema+" CASCADE") }()
	_, err = db.ExecContext(ctx, "SET search_path="+schema)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `CREATE TABLE accounts(id BIGINT PRIMARY KEY,credentials JSONB NOT NULL,proxy_id BIGINT,platform TEXT,type TEXT,deleted_at TIMESTAMPTZ,updated_at TIMESTAMPTZ DEFAULT NOW(),status TEXT DEFAULT 'active',schedulable BOOLEAN DEFAULT TRUE);
 CREATE TABLE scheduler_outbox(id BIGSERIAL PRIMARY KEY,event_type TEXT NOT NULL,account_id BIGINT CHECK(account_id<>99),group_id BIGINT,payload JSONB);`)
	require.NoError(t, err)
	repository := newAccountRepositoryWithSQL(nil, db, nil)
	old := map[string]any{"access_token": "old", "refresh_token": "old-refresh", "client_id": "client"}
	rotated := map[string]any{"access_token": "new", "refresh_token": "rotated-refresh", "client_id": "client"}
	oldJSON, _ := json.Marshal(old)
	create := func(id int64) {
		_, insertErr := db.ExecContext(ctx, "INSERT INTO accounts(id,credentials,platform,type) VALUES($1,$2,$3,$4)", id, string(oldJSON), service.PlatformKiro, service.AccountTypeOAuth)
		require.NoError(t, insertErr)
	}
	t.Run("matching_credentials_and_null_proxy", func(t *testing.T) {
		create(1)
		_, err := db.ExecContext(ctx, "UPDATE accounts SET status='error',schedulable=false WHERE id=1")
		require.NoError(t, err)
		applied, err := repository.UpdateKiroOAuthCredentialsIfUnchanged(ctx, 1, old, nil, rotated)
		require.NoError(t, err)
		require.True(t, applied)
		var token, status string
		var schedulable bool
		require.NoError(t, db.QueryRowContext(ctx, "SELECT credentials->>'refresh_token',status,schedulable FROM accounts WHERE id=1").Scan(&token, &status, &schedulable))
		require.Equal(t, "rotated-refresh", token)
		require.Equal(t, "error", status)
		require.False(t, schedulable)
		var count int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM scheduler_outbox WHERE account_id=1").Scan(&count))
		require.Equal(t, 1, count)
	})
	for index, mutation := range []string{
		"credentials=jsonb_set(credentials,'{refresh_token}','\"admin-refresh\"')",
		"proxy_id=9", "platform='anthropic'", "type='apikey'", "deleted_at=NOW()",
	} {
		t.Run(mutation, func(t *testing.T) {
			id := int64(index + 2)
			create(id)
			_, err := db.ExecContext(ctx, "UPDATE accounts SET "+mutation+" WHERE id=$1", id)
			require.NoError(t, err)
			applied, err := repository.UpdateKiroOAuthCredentialsIfUnchanged(ctx, id, old, nil, rotated)
			require.NoError(t, err)
			require.False(t, applied)
			var count int
			require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM scheduler_outbox WHERE account_id=$1", id).Scan(&count))
			require.Zero(t, count)
		})
	}
	t.Run("matching_configured_proxy", func(t *testing.T) {
		create(20)
		proxyID := int64(9)
		_, err := db.ExecContext(ctx, "UPDATE accounts SET proxy_id=9 WHERE id=20")
		require.NoError(t, err)
		applied, err := repository.UpdateKiroOAuthCredentialsIfUnchanged(ctx, 20, old, &proxyID, rotated)
		require.NoError(t, err)
		require.True(t, applied)
	})
	t.Run("concurrent_rotation_applies_once", func(t *testing.T) {
		create(55)
		var wait sync.WaitGroup
		var appliedCount atomic.Int32
		errors := make(chan error, 12)
		for range 12 {
			wait.Add(1)
			go func() {
				defer wait.Done()
				applied, err := repository.UpdateKiroOAuthCredentialsIfUnchanged(ctx, 55, old, nil, rotated)
				if err != nil {
					errors <- err
				}
				if applied {
					appliedCount.Add(1)
				}
			}()
		}
		wait.Wait()
		close(errors)
		for err := range errors {
			require.NoError(t, err)
		}
		require.Equal(t, int32(1), appliedCount.Load())
	})
	t.Run("outbox_failure_rolls_back_rotation", func(t *testing.T) {
		create(99)
		applied, err := repository.UpdateKiroOAuthCredentialsIfUnchanged(ctx, 99, old, nil, rotated)
		require.Error(t, err)
		require.False(t, applied)
		var token string
		require.NoError(t, db.QueryRowContext(ctx, "SELECT credentials->>'refresh_token' FROM accounts WHERE id=99").Scan(&token))
		require.Equal(t, "old-refresh", token)
	})
}
