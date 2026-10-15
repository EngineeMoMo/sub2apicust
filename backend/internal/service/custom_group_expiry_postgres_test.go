//go:build unit

package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func customGroupExpiryPostgresFixture(t *testing.T) (*customGroupAccessSQLStore, time.Time) {
	t.Helper()
	dsn := os.Getenv("CUSTOM_GROUP_ACCESS_TEST_DSN")
	if dsn == "" {
		t.Skip("CUSTOM_GROUP_ACCESS_TEST_DSN未配置，未验证真实分组平移事务")
	}
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		var err error
		dsn, err = pq.ParseURL(dsn)
		require.NoError(t, err)
	}
	database, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })
	var name string
	require.NoError(t, database.QueryRow("SELECT current_database()").Scan(&name))
	require.Equal(t, "group_access_test", name)
	schema := fmt.Sprintf("group_expiry_%d", time.Now().UnixNano())
	_, err = database.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = database.Exec("DROP SCHEMA " + schema + " CASCADE") })
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(8)
	_, err = db.Exec(`CREATE TABLE groups(id BIGINT PRIMARY KEY,platform TEXT NOT NULL DEFAULT 'openai',status TEXT NOT NULL DEFAULT 'active',subscription_type TEXT NOT NULL DEFAULT 'standard',fallback_group_id BIGINT,fallback_group_id_on_invalid_request BIGINT,deleted_at TIMESTAMPTZ,updated_at TIMESTAMPTZ DEFAULT NOW());
 CREATE TABLE users(id BIGINT PRIMARY KEY,balance NUMERIC,status TEXT DEFAULT 'active',deleted_at TIMESTAMPTZ);
 CREATE TABLE payment_orders(id BIGSERIAL PRIMARY KEY,user_id BIGINT,paid_at TIMESTAMPTZ,order_type TEXT,status TEXT);
 CREATE TABLE api_keys(id BIGINT PRIMARY KEY,key TEXT NOT NULL,user_id BIGINT NOT NULL REFERENCES users(id),group_id BIGINT REFERENCES groups(id),quota NUMERIC DEFAULT 100,quota_used NUMERIC DEFAULT 7,status TEXT DEFAULT 'active',expires_at TIMESTAMPTZ DEFAULT '2030-01-01',deleted_at TIMESTAMPTZ,updated_at TIMESTAMPTZ DEFAULT NOW());
 CREATE TABLE user_allowed_groups(user_id BIGINT REFERENCES users(id),group_id BIGINT REFERENCES groups(id),created_at TIMESTAMPTZ DEFAULT NOW(),PRIMARY KEY(user_id,group_id));
 CREATE TABLE user_group_rate_multipliers(user_id BIGINT,group_id BIGINT,rate_multiplier NUMERIC);
 CREATE TABLE user_subscriptions(id BIGINT PRIMARY KEY,user_id BIGINT,group_id BIGINT,deleted_at TIMESTAMPTZ);
 CREATE TABLE custom_dedicated_accounts(id BIGINT PRIMARY KEY,group_id BIGINT);
 CREATE TABLE account_groups(account_id BIGINT,group_id BIGINT);
 CREATE TABLE composite_model_routes(group_id BIGINT,deleted_at TIMESTAMPTZ);
 CREATE TABLE settings(key TEXT PRIMARY KEY,value TEXT,updated_at TIMESTAMPTZ DEFAULT NOW());
 CREATE TABLE usage_logs(id BIGINT PRIMARY KEY,group_id BIGINT REFERENCES groups(id),actual_cost NUMERIC);
 CREATE TABLE scheduler_outbox(id BIGSERIAL PRIMARY KEY,event_type TEXT,group_id BIGINT);
 INSERT INTO groups(id) VALUES(1),(2),(3),(4);
 INSERT INTO users(id,balance) VALUES(7,100),(8,80),(9,50),(10,40);
 INSERT INTO api_keys(id,key,user_id,group_id,status) VALUES(1,'key-source-a',7,1,'active'),(2,'key-source-b',8,1,'disabled'),(3,'key-other',7,4,'active');
 INSERT INTO api_keys(id,key,user_id,group_id,deleted_at) VALUES(4,'key-deleted',10,1,NOW());
 INSERT INTO user_allowed_groups(user_id,group_id) VALUES(7,1),(9,1),(7,2);
 INSERT INTO user_group_rate_multipliers VALUES(7,1,0.07),(7,2,0.3);
 INSERT INTO account_groups VALUES(11,1),(12,2);
 INSERT INTO composite_model_routes VALUES(1,NULL);
 INSERT INTO usage_logs VALUES(1,1,0.1234);
 UPDATE groups SET fallback_group_id=1 WHERE id=4;
 UPDATE groups SET fallback_group_id_on_invalid_request=1 WHERE id=2;`)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO settings(key,value) VALUES($1,$2)", SettingKeyContentModerationConfig, `{"group_ids":[1,2]}`)
	require.NoError(t, err)
	migration, err := os.ReadFile("../../migrations/257_custom_group_access_policy.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(migration))
	require.NoError(t, err)
	migration, err = os.ReadFile("../../migrations/258_custom_group_expiry_binding_guard.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(migration))
	require.NoError(t, err)
	now := time.Now().UTC().Truncate(time.Microsecond)
	store := &customGroupAccessSQLStore{db: db}
	target := int64(2)
	expires := now.Add(-time.Second)
	_, err = store.Save(context.Background(), CustomGroupAccessPolicy{GroupID: 1, ExpiresAt: &expires, FallbackGroupID: &target})
	require.NoError(t, err)
	return store, now
}

func TestCustomGroupExpiryPostgresPersistentMoveAndConcurrency(t *testing.T) {
	store, now := customGroupExpiryPostgresFixture(t)
	var wait sync.WaitGroup
	results := make(chan *customGroupExpiryMove, 2)
	errors := make(chan error, 2)
	for attempt := 0; attempt < 2; attempt++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			move, err := store.moveExpiredGroup(context.Background(), 1, now)
			results <- move
			errors <- err
		}()
	}
	wait.Wait()
	close(results)
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}
	completed := 0
	for move := range results {
		if move != nil {
			completed++
			require.EqualValues(t, 2, move.targetID)
			require.EqualValues(t, 2, move.keys)
			require.Equal(t, 3, move.users)
			require.ElementsMatch(t, []string{"key-source-a", "key-source-b", "key-other"}, move.authKeys)
		}
	}
	require.Equal(t, 1, completed)
	checks := []struct {
		query string
		want  int
	}{
		{"SELECT count(*) FROM groups WHERE id=1 AND deleted_at IS NOT NULL", 1},
		{"SELECT count(*) FROM api_keys WHERE group_id=1 AND deleted_at IS NULL", 0},
		{"SELECT count(*) FROM api_keys WHERE group_id=2 AND deleted_at IS NULL", 2},
		{"SELECT count(*) FROM api_keys WHERE id=4 AND group_id=1 AND deleted_at IS NOT NULL", 1},
		{"SELECT count(*) FROM api_keys WHERE id IN (1,2) AND quota=100 AND quota_used=7 AND expires_at='2030-01-01'", 2},
		{"SELECT count(*) FROM api_keys WHERE id=2 AND status='disabled'", 1},
		{"SELECT count(*) FROM user_allowed_groups WHERE group_id=1", 0},
		{"SELECT count(*) FROM user_allowed_groups WHERE group_id=2 AND user_id IN (7,8,9)", 3},
		{"SELECT count(*) FROM user_allowed_groups WHERE user_id=10 AND group_id=2", 0},
		{"SELECT count(*) FROM account_groups WHERE group_id=1", 0},
		{"SELECT count(*) FROM account_groups WHERE group_id=2", 1},
		{"SELECT count(*) FROM composite_model_routes WHERE group_id=1 AND deleted_at IS NOT NULL", 1},
		{"SELECT count(*) FROM usage_logs WHERE group_id=1 AND actual_cost=0.1234", 1},
		{"SELECT count(*) FROM user_group_rate_multipliers WHERE user_id=7 AND group_id=2 AND rate_multiplier=0.3", 1},
		{"SELECT count(*) FROM groups WHERE id=4 AND fallback_group_id=2", 1},
		{"SELECT count(*) FROM groups WHERE id=2 AND fallback_group_id_on_invalid_request IS NULL", 1},
		{"SELECT count(*) FROM scheduler_outbox", 2},
	}
	for _, check := range checks {
		var actual int
		require.NoError(t, store.db.QueryRow(check.query).Scan(&actual))
		require.Equal(t, check.want, actual, check.query)
	}
	var balance float64
	require.NoError(t, store.db.QueryRow("SELECT sum(balance) FROM users").Scan(&balance))
	require.Equal(t, float64(270), balance)
	var moderation string
	require.NoError(t, store.db.QueryRow("SELECT value FROM settings WHERE key=$1", SettingKeyContentModerationConfig).Scan(&moderation))
	require.JSONEq(t, `{"group_ids":[2]}`, moderation)
	move, err := store.moveExpiredGroup(context.Background(), 1, now)
	require.NoError(t, err)
	require.Nil(t, move)
	policies, err := store.List(context.Background())
	require.NoError(t, err)
	_, err = store.Save(context.Background(), policies[1])
	require.ErrorIs(t, err, ErrGroupNotFound)
}

func TestCustomGroupExpiryPostgresFailureRetainsSource(t *testing.T) {
	for _, mode := range []string{"inactive", "platform", "subscription", "dedicated", "rollback"} {
		t.Run(mode, func(t *testing.T) {
			store, now := customGroupExpiryPostgresFixture(t)
			var statement string
			switch mode {
			case "inactive":
				statement = "UPDATE groups SET status='disabled' WHERE id=2"
			case "platform":
				statement = "UPDATE groups SET platform='anthropic' WHERE id=2"
			case "subscription":
				statement = "UPDATE groups SET subscription_type='subscription' WHERE id=2"
			case "dedicated":
				statement = "INSERT INTO custom_dedicated_accounts VALUES(1,2)"
			case "rollback":
				statement = `CREATE FUNCTION deny_expiry_delete() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.id=1 AND NEW.deleted_at IS NOT NULL THEN RAISE EXCEPTION 'injected deletion failure'; END IF; RETURN NEW; END; $$; CREATE TRIGGER deny_expiry_delete BEFORE UPDATE ON groups FOR EACH ROW EXECUTE FUNCTION deny_expiry_delete()`
			}
			_, err := store.db.Exec(statement)
			require.NoError(t, err)
			move, err := store.moveExpiredGroup(context.Background(), 1, now)
			require.Error(t, err)
			require.Nil(t, move)
			for _, query := range []string{"SELECT count(*) FROM groups WHERE id=1 AND deleted_at IS NULL", "SELECT count(*) FROM account_groups WHERE group_id=1"} {
				var count int
				require.NoError(t, store.db.QueryRow(query).Scan(&count))
				require.Equal(t, 1, count)
			}
			var count int
			require.NoError(t, store.db.QueryRow("SELECT count(*) FROM api_keys WHERE group_id=1 AND deleted_at IS NULL").Scan(&count))
			require.Equal(t, 2, count)
			require.NoError(t, store.db.QueryRow("SELECT count(*) FROM user_allowed_groups WHERE group_id=1").Scan(&count))
			require.Equal(t, 2, count)
			require.NoError(t, store.db.QueryRow("SELECT count(*) FROM scheduler_outbox").Scan(&count))
			require.Zero(t, count)
		})
	}
}

func TestCustomGroupExpiryPostgresConcurrentBindings(t *testing.T) {
	t.Run("binding_before_expiry_is_transferred", func(t *testing.T) {
		store, now := customGroupExpiryPostgresFixture(t)
		tx, err := store.db.Begin()
		require.NoError(t, err)
		defer func() { _ = tx.Rollback() }()
		_, err = tx.Exec("INSERT INTO api_keys(id,key,user_id,group_id) VALUES(5,'concurrent-key',10,1)")
		require.NoError(t, err)
		done := make(chan error, 1)
		go func() { _, err := store.moveExpiredGroup(context.Background(), 1, now); done <- err }()
		select {
		case err := <-done:
			t.Fatalf("迁移应等待未提交的绑定事务: %v", err)
		case <-time.After(50 * time.Millisecond):
		}
		require.NoError(t, tx.Commit())
		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(3 * time.Second):
			t.Fatal("迁移等待绑定事务后超时")
		}
		var target int64
		require.NoError(t, store.db.QueryRow("SELECT group_id FROM api_keys WHERE id=5").Scan(&target))
		require.EqualValues(t, 2, target)
		var allowed bool
		require.NoError(t, store.db.QueryRow("SELECT EXISTS(SELECT 1 FROM user_allowed_groups WHERE user_id=10 AND group_id=2)").Scan(&allowed))
		require.True(t, allowed)
	})
	for _, statement := range []string{
		"INSERT INTO api_keys(id,key,user_id,group_id) VALUES(5,'stale-key',10,1)",
		"INSERT INTO user_allowed_groups(user_id,group_id) VALUES(10,1)",
	} {
		t.Run(statement, func(t *testing.T) {
			store, now := customGroupExpiryPostgresFixture(t)
			tx, err := store.db.Begin()
			require.NoError(t, err)
			defer func() { _ = tx.Rollback() }()
			_, err = tx.Exec("UPDATE groups SET deleted_at=$1 WHERE id=1", now)
			require.NoError(t, err)
			done := make(chan error, 1)
			go func() { _, err := store.db.Exec(statement); done <- err }()
			select {
			case err := <-done:
				t.Fatalf("旧绑定应等待删组事务: %v", err)
			case <-time.After(50 * time.Millisecond):
			}
			require.NoError(t, tx.Commit())
			select {
			case err := <-done:
				var databaseError *pq.Error
				require.ErrorAs(t, err, &databaseError)
				require.Equal(t, pq.ErrorCode("23503"), databaseError.Code)
			case <-time.After(3 * time.Second):
				t.Fatal("删组后旧绑定未返回")
			}
		})
	}
}

func TestCustomGroupExpiryPostgresFailureDoesNotDelayNextBoundary(t *testing.T) {
	store, now := customGroupExpiryPostgresFixture(t)
	_, err := store.db.Exec("UPDATE groups SET status='disabled' WHERE id=2")
	require.NoError(t, err)
	access := &CustomGroupAccessService{store: store, now: func() time.Time { return now }, cache: newCustomGroupAccessCache()}
	target, expires := int64(3), now.Add(100*time.Millisecond)
	_, err = store.Save(context.Background(), CustomGroupAccessPolicy{GroupID: 4, ExpiresAt: &expires, FallbackGroupID: &target})
	require.NoError(t, err)
	require.Equal(t, 100*time.Millisecond, access.runGroupExpiry(context.Background(), store, nil))
	_, err = store.db.Exec("INSERT INTO custom_dedicated_accounts VALUES(1,2)")
	require.NoError(t, err)
	require.ErrorIs(t, store.checkExpiryDedicatedGroups(context.Background(), 1, 2), ErrGroupAccessExpiryDedicated)
	require.NoError(t, store.checkExpiryDedicatedGroups(context.Background(), 3, 4))
}

func TestCustomGroupExpiryPostgresWorkerBoundaryAndRestart(t *testing.T) {
	for _, mode := range []string{"boundary", "restart", "wake", "chain"} {
		t.Run(mode, func(t *testing.T) {
			store, now := customGroupExpiryPostgresFixture(t)
			if mode == "boundary" || mode == "wake" {
				expires := now.Add(300 * time.Millisecond)
				if mode == "wake" {
					expires = now.Add(time.Hour)
				}
				_, err := store.db.Exec("UPDATE custom_group_access_policies SET expires_at=$1 WHERE group_id=1", expires)
				require.NoError(t, err)
			}
			if mode == "chain" {
				target := int64(3)
				expires := now.Add(-time.Second)
				_, err := store.Save(context.Background(), CustomGroupAccessPolicy{GroupID: 2, ExpiresAt: &expires, FallbackGroupID: &target})
				require.NoError(t, err)
			}
			access := &CustomGroupAccessService{store: store, now: time.Now, cache: newCustomGroupAccessCache(), expiryWake: make(chan struct{}, 1)}
			stop := access.startGroupExpiry(nil)
			t.Cleanup(stop)
			if mode == "wake" {
				time.Sleep(50 * time.Millisecond)
				_, err := store.db.Exec("UPDATE custom_group_access_policies SET expires_at=$1 WHERE group_id=1", now.Add(-time.Second))
				require.NoError(t, err)
				access.wakeGroupExpiry()
			}
			require.Eventually(t, func() bool {
				var count int
				err := store.db.QueryRow("SELECT count(*) FROM groups WHERE id=1 AND deleted_at IS NOT NULL").Scan(&count)
				return err == nil && count == 1
			}, 3*time.Second, 10*time.Millisecond)
			if mode == "chain" {
				require.Eventually(t, func() bool {
					var count int
					err := store.db.QueryRow("SELECT count(*) FROM api_keys WHERE id IN (1,2) AND group_id=3").Scan(&count)
					return err == nil && count == 2
				}, time.Second, 10*time.Millisecond)
			}
			stopped := make(chan struct{})
			go func() { stop(); close(stopped) }()
			select {
			case <-stopped:
			case <-time.After(time.Second):
				t.Fatal("到期任务退出超时")
			}
			stop()
		})
	}
}
