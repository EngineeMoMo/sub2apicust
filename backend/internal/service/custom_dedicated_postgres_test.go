//go:build unit

package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestCustomDedicatedPostgres(t *testing.T) {
	dsn := os.Getenv("DEDICATED_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("DEDICATED_TEST_POSTGRES_DSN is not set; requires an isolated test database")
	}
	database, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func() { _ = database.Close() }()
	database.SetMaxOpenConns(1)
	schema := fmt.Sprintf("dedicated_test_%d", time.Now().UnixNano())
	_, err = database.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	defer func() { _, _ = database.Exec("DROP SCHEMA " + schema + " CASCADE") }()
	_, err = database.Exec("SET search_path TO " + schema)
	require.NoError(t, err)
	_, err = database.Exec(`
CREATE TABLE users(id bigint PRIMARY KEY, username text, email text, status text DEFAULT 'active', deleted_at timestamptz);
CREATE TABLE accounts(id bigint PRIMARY KEY, name text, platform text DEFAULT 'openai', type text DEFAULT 'oauth', parent_account_id bigint, deleted_at timestamptz);
CREATE TABLE groups(id bigint PRIMARY KEY, name text, status text DEFAULT 'active', is_exclusive boolean DEFAULT true, platform text DEFAULT 'openai', subscription_type text DEFAULT 'standard', fallback_group_id bigint, fallback_group_id_on_invalid_request bigint, deleted_at timestamptz);
CREATE TABLE account_groups(account_id bigint, group_id bigint);
CREATE TABLE user_allowed_groups(user_id bigint, group_id bigint);
CREATE TABLE api_keys(user_id bigint, group_id bigint, deleted_at timestamptz);
CREATE TABLE user_subscriptions(group_id bigint, deleted_at timestamptz);
INSERT INTO users(id,username) VALUES(11,'甲'),(12,'乙'),(13,'外部用户');
INSERT INTO accounts(id,name) VALUES(22,'账号A'),(44,'账号B');
INSERT INTO groups(id,name) VALUES(33,'组A'),(55,'组B');
INSERT INTO account_groups VALUES(22,33),(44,55);
INSERT INTO user_allowed_groups VALUES(11,33),(12,33),(11,55),(12,55);
INSERT INTO api_keys(user_id,group_id) VALUES(11,33),(12,33),(11,55),(12,55);`)
	require.NoError(t, err)
	for _, filename := range []string{"241_custom_dedicated_accounts.sql", "242_custom_dedicated_members.sql"} {
		migration, err := os.ReadFile("../../migrations/" + filename)
		require.NoError(t, err)
		_, err = database.Exec(string(migration))
		require.NoError(t, err)
		if filename == "241_custom_dedicated_accounts.sql" {
			_, err = database.Exec("INSERT INTO custom_dedicated_accounts(user_id,account_id,group_id,label,expires_at) VALUES(11,22,33,'旧单人绑定',NOW()+INTERVAL '1 day')")
			require.NoError(t, err)
		}
	}
	legacy, err := scanCustomDedicated(database.QueryRow(customDedicatedViewSQL, 1, 11))
	require.NoError(t, err)
	require.Equal(t, []int64{11}, legacy.UserIDs)
	_, err = scanCustomDedicated(database.QueryRow(customDedicatedViewSQL, 1, 12))
	require.ErrorIs(t, err, sql.ErrNoRows)
	_, err = database.Exec("DELETE FROM custom_dedicated_accounts WHERE id=1")
	require.NoError(t, err)
	cache := &customDedicatedInvalidator{}
	service := &CustomDedicatedService{db: database, apiKeys: cache}
	input := CustomDedicatedInput{UserIDs: []int64{11, 12}, AccountID: 22, GroupID: 33, Label: "多人包号", ExpiresAt: time.Now().Add(time.Hour)}
	binding, err := service.Save(context.Background(), 0, input)
	require.NoError(t, err)
	memberContext := context.WithValue(context.Background(), customDedicatedContextKey{}, customDedicatedSubject{12, 33})
	require.NoError(t, service.Check(memberContext, &Account{ID: 22}, nil))
	require.ErrorIs(t, service.Check(context.WithValue(context.Background(), customDedicatedContextKey{}, customDedicatedSubject{13, 33}), &Account{ID: 22}, nil), ErrDedicatedAccess)
	list, err := service.List(context.Background(), 12, 1)
	require.NoError(t, err)
	require.Len(t, list, 1)
	admin, err := service.AdminList(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, "乙", admin[0].Users[1].Name)
	views := []CustomDedicatedView{{ID: binding.ID}}
	require.NoError(t, service.FillGroupNames(context.Background(), 12, views))
	require.Equal(t, "组A", views[0].GroupName)
	_, err = scanCustomDedicated(database.QueryRow(customDedicatedViewSQL, binding.ID, 12))
	require.NoError(t, err)
	_, err = database.Exec("INSERT INTO api_keys(user_id,group_id) VALUES(13,33)")
	require.NoError(t, err)
	require.ErrorIs(t, service.Check(memberContext, &Account{ID: 22}, nil), ErrDedicatedAccess)
	_, err = database.Exec("DELETE FROM api_keys WHERE user_id=13")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM api_keys WHERE user_id=12 AND group_id=33; DELETE FROM user_allowed_groups WHERE user_id=12 AND group_id=33")
	require.NoError(t, err)
	input.UserIDs = []int64{11}
	_, err = service.Save(context.Background(), binding.ID, input)
	require.NoError(t, err)
	require.ErrorIs(t, service.Check(memberContext, &Account{ID: 22}, nil), ErrDedicatedAccess)
	_, err = database.Exec("INSERT INTO user_allowed_groups VALUES(12,33); INSERT INTO api_keys(user_id,group_id) VALUES(12,33)")
	require.NoError(t, err)
	input.UserIDs = []int64{11, 12}
	_, err = service.Save(context.Background(), binding.ID, input)
	require.NoError(t, err)
	require.ErrorIs(t, service.Delete(context.Background(), binding.ID), ErrDedicatedDelete)
	input.AccountID, input.GroupID = 44, 55
	_, err = service.Save(context.Background(), binding.ID, input)
	require.NoError(t, err)
	require.ErrorIs(t, service.Check(memberContext, &Account{ID: 22}, nil), ErrDedicatedAccess)
	newContext := context.WithValue(context.Background(), customDedicatedContextKey{}, customDedicatedSubject{12, 55})
	require.NoError(t, service.Check(newContext, &Account{ID: 44}, nil))
	require.NoError(t, service.Revoke(context.Background(), binding.ID))
	require.NoError(t, service.Delete(context.Background(), binding.ID))
	admin, err = service.AdminList(context.Background(), 1)
	require.NoError(t, err)
	require.Empty(t, admin)
	list, err = service.List(context.Background(), 12, 1)
	require.NoError(t, err)
	require.Empty(t, list)
	require.ErrorIs(t, service.Check(newContext, &Account{ID: 44}, nil), ErrDedicatedAccess)
	_, err = service.Save(context.Background(), 0, input)
	require.NoError(t, err)
	require.NoError(t, service.Check(newContext, nil, nil))
	require.NoError(t, service.Check(newContext, &Account{ID: 44}, nil))
	require.ErrorIs(t, service.Check(memberContext, &Account{ID: 22}, nil), ErrDedicatedAccess)
}
