//go:build unit

package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestCustomDedicatedConfigPostgres(t *testing.T) {
	dsn := os.Getenv("DEDICATED_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("DEDICATED_TEST_POSTGRES_DSN is not set; requires an isolated test database")
	}
	database, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func() { _ = database.Close() }()
	database.SetMaxOpenConns(1)
	schema := fmt.Sprintf("dedicated_config_test_%d", time.Now().UnixNano())
	_, err = database.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	defer func() { _, _ = database.Exec("DROP SCHEMA " + schema + " CASCADE") }()
	_, err = database.Exec("SET search_path TO " + schema)
	require.NoError(t, err)
	_, err = database.Exec(`
CREATE TABLE users(id bigint PRIMARY KEY, status text DEFAULT 'active', deleted_at timestamptz);
CREATE TABLE accounts(id bigint PRIMARY KEY, platform text DEFAULT 'openai', type text DEFAULT 'oauth', parent_account_id bigint, deleted_at timestamptz);
CREATE TABLE groups(id bigint PRIMARY KEY, name text, status text DEFAULT 'active', is_exclusive boolean DEFAULT true, platform text DEFAULT 'openai', subscription_type text DEFAULT 'standard', fallback_group_id bigint, fallback_group_id_on_invalid_request bigint, deleted_at timestamptz);
CREATE TABLE account_groups(account_id bigint, group_id bigint);
CREATE TABLE user_allowed_groups(user_id bigint, group_id bigint);
CREATE TABLE api_keys(user_id bigint, group_id bigint, status text DEFAULT 'active', updated_at timestamptz DEFAULT NOW(), deleted_at timestamptz);
CREATE TABLE user_subscriptions(group_id bigint, deleted_at timestamptz);`)
	require.NoError(t, err)
	for _, filename := range []string{"241_custom_dedicated_accounts.sql", "242_custom_dedicated_members.sql"} {
		migration, err := os.ReadFile("../../migrations/" + filename)
		require.NoError(t, err)
		_, err = database.Exec(string(migration))
		require.NoError(t, err)
	}
	for _, scenario := range []struct {
		name   string
		change string
		issue  string
		ids    string
	}{
		{"valid_members_and_permissions", "", "", ""},
		{"members_can_use_unrelated_public_groups", "INSERT INTO groups(id,name,is_exclusive) VALUES(77,'不定期福利',false),(88,'GPT_PRO按量',false); INSERT INTO user_allowed_groups VALUES(11,77),(12,77),(11,88),(12,88)", "", ""},
		{"two_authorized_users_only_first_selected", "", "other_authorized_users", "12"},
		{"two_authorized_users_only_second_selected", "", "other_authorized_users", "11"},
		{"account_missing", "UPDATE accounts SET deleted_at=NOW() WHERE id=22", "account_missing", "22"},
		{"group_missing", "UPDATE groups SET deleted_at=NOW() WHERE id=33", "group_missing", "33"},
		{"group_disabled", "UPDATE groups SET status='disabled' WHERE id=33", "group_inactive", "33"},
		{"name_does_not_enable_exclusive", "UPDATE groups SET is_exclusive=false WHERE id=33", "group_not_exclusive", "33"},
		{"platform_mismatch", "UPDATE groups SET platform='anthropic' WHERE id=33", "platform_mismatch", "33"},
		{"subscription_group", "UPDATE groups SET subscription_type='subscription' WHERE id=33", "group_not_standard", "33"},
		{"fallback", "UPDATE groups SET fallback_group_id=55 WHERE id=33", "fallback_route", "33"},
		{"invalid_request_fallback", "UPDATE groups SET fallback_group_id_on_invalid_request=55 WHERE id=33", "fallback_route", "33"},
		{"unsupported_account_type", "UPDATE accounts SET type='apikey' WHERE id=22", "account_type", "22"},
		{"shadow_account", "UPDATE accounts SET parent_account_id=44 WHERE id=22", "shadow_account", "22"},
		{"shadow_children", "INSERT INTO accounts(id,parent_account_id) VALUES(66,22)", "account_has_shadows", "66"},
		{"account_not_joined", "DELETE FROM account_groups WHERE account_id=22 AND group_id=33", "account_not_in_group", "22"},
		{"account_in_multiple_exclusive_groups", "DELETE FROM account_groups WHERE account_id=44; INSERT INTO account_groups VALUES(22,55); INSERT INTO user_allowed_groups VALUES(11,55),(12,55)", "", ""},
		{"account_in_public_group", "UPDATE groups SET is_exclusive=false WHERE id=55; INSERT INTO account_groups VALUES(22,55)", "account_unsafe_groups", "55"},
		{"related_group_has_other_account", "INSERT INTO account_groups VALUES(22,55)", "related_group_other_accounts", "55"},
		{"related_group_outside_user", "DELETE FROM account_groups WHERE account_id=44; INSERT INTO account_groups VALUES(22,55); INSERT INTO user_allowed_groups VALUES(13,55)", "other_authorized_users", "13"},
		{"related_group_outside_key", "DELETE FROM account_groups WHERE account_id=44; INSERT INTO account_groups VALUES(22,55); INSERT INTO api_keys(user_id,group_id) VALUES(13,55)", "non_member_keys", "13"},
		{"group_contains_another_account", "INSERT INTO account_groups VALUES(44,33)", "group_other_accounts", "44"},
		{"outsider_also_authorized", "INSERT INTO user_allowed_groups VALUES(13,33)", "other_authorized_users", "13"},
		{"subscription_record", "INSERT INTO user_subscriptions VALUES(33,NULL)", "group_subscriptions", "33"},
		{"member_disabled", "UPDATE users SET status='disabled' WHERE id=12", "member_unavailable", "12"},
		{"member_missing_permission", "DELETE FROM user_allowed_groups WHERE user_id=12 AND group_id=33", "member_not_authorized", "12"},
		{"outsider_enabled_key", "INSERT INTO api_keys(user_id,group_id) VALUES(13,33)", "non_member_keys", "13"},
		{"outsider_inactive_key_allowed", "INSERT INTO api_keys(user_id,group_id,status) VALUES(13,33,'inactive')", "", ""},
		{"outsider_legacy_disabled_key_allowed", "INSERT INTO api_keys(user_id,group_id,status) VALUES(13,33,'disabled')", "", ""},
		{"bounded_conflicting_user_list", "INSERT INTO users(id) SELECT generate_series(100,115); INSERT INTO user_allowed_groups SELECT generate_series(100,115),33", "other_authorized_users", "100, 101, 102, 103, 104, 105, 106, 107, 108, 109"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			_, err := database.Exec(`TRUNCATE custom_dedicated_accounts, users, accounts, groups, account_groups, user_allowed_groups, api_keys, user_subscriptions CASCADE;
INSERT INTO users(id) VALUES(11),(12),(13);
INSERT INTO accounts(id) VALUES(22),(44);
INSERT INTO groups(id,name) VALUES(33,'用户专属分组'),(55,'另一个分组');
INSERT INTO account_groups VALUES(22,33),(44,55);
INSERT INTO user_allowed_groups VALUES(11,33),(12,33);`)
			require.NoError(t, err)
			if scenario.change != "" {
				_, err = database.Exec(scenario.change)
				require.NoError(t, err)
			}
			cache := &customDedicatedInvalidator{}
			service := &CustomDedicatedService{db: database, apiKeys: cache}
			input := CustomDedicatedInput{UserIDs: []int64{11, 12}, AccountID: 22, GroupID: 33, Label: "包号验证", ExpiresAt: time.Now().Add(time.Hour)}
			if scenario.name == "two_authorized_users_only_first_selected" {
				input.UserIDs = []int64{11}
			} else if scenario.name == "two_authorized_users_only_second_selected" {
				input.UserIDs = []int64{12}
			}
			binding, err := service.Save(context.Background(), 0, input)
			if scenario.issue == "" {
				require.NoError(t, err)
				require.NotNil(t, binding)
				if scenario.name == "account_in_multiple_exclusive_groups" {
					testCustomDedicatedMultipleGroups(t, database, service, binding, input)
				}
			} else {
				require.Nil(t, binding)
				require.ErrorIs(t, err, ErrDedicatedConfig)
				var status *infraerrors.ApplicationError
				require.ErrorAs(t, err, &status)
				require.Equal(t, scenario.issue, status.Metadata["config_issue"])
				require.Equal(t, scenario.ids, status.Metadata["resource_ids"])
				var count int
				require.NoError(t, database.QueryRow("SELECT COUNT(*) FROM custom_dedicated_accounts").Scan(&count))
				require.Zero(t, count)
				require.Empty(t, cache.groups)
				require.Empty(t, cache.users)
			}
		})
	}
}

func testCustomDedicatedMultipleGroups(t *testing.T, database *sql.DB, service *CustomDedicatedService, binding *CustomDedicatedBinding, input CustomDedicatedInput) {
	ctx := context.Background()
	member := func(userID, groupID int64) context.Context {
		return context.WithValue(ctx, customDedicatedContextKey{}, customDedicatedSubject{UserID: userID, GroupID: groupID})
	}
	for _, groupID := range []int64{33, 55} {
		for _, userID := range []int64{11, 12} {
			require.NoError(t, service.Check(member(userID, groupID), nil, &groupID))
			require.NoError(t, service.Check(member(userID, groupID), &Account{ID: 22}, nil))
			require.ErrorIs(t, service.Check(member(userID, groupID), &Account{ID: 44}, nil), ErrDedicatedAccess)
		}
		require.ErrorIs(t, service.Check(member(13, groupID), &Account{ID: 22}, nil), ErrDedicatedAccess)
	}
	require.ErrorIs(t, service.Check(member(11, 77), &Account{ID: 22}, nil), ErrDedicatedAccess)
	_, err := database.Exec("DELETE FROM user_allowed_groups WHERE user_id=12 AND group_id=55")
	require.NoError(t, err)
	require.ErrorIs(t, service.Check(member(12, 55), &Account{ID: 22}, nil), ErrDedicatedAccess)
	require.NoError(t, service.Check(member(12, 33), &Account{ID: 22}, nil))
	require.NoError(t, service.Check(member(11, 55), &Account{ID: 22}, nil))
	_, err = database.Exec("INSERT INTO user_allowed_groups VALUES(12,55); UPDATE groups SET is_exclusive=false WHERE id=55")
	require.NoError(t, err)
	require.ErrorIs(t, service.Check(member(11, 33), &Account{ID: 22}, nil), ErrDedicatedAccess)
	_, err = database.Exec("UPDATE groups SET is_exclusive=true WHERE id=55; INSERT INTO user_allowed_groups VALUES(13,55)")
	require.NoError(t, err)
	require.ErrorIs(t, service.Check(member(11, 33), &Account{ID: 22}, nil), ErrDedicatedAccess)
	_, err = database.Exec("DELETE FROM user_allowed_groups WHERE user_id=13; INSERT INTO groups(id,name) VALUES(77,'无关分组'); INSERT INTO user_allowed_groups VALUES(12,77); INSERT INTO api_keys(user_id,group_id) VALUES(12,33),(12,55),(12,77)")
	require.NoError(t, err)
	input.ExpectedUpdatedAt = &binding.UpdatedAt
	input.UserIDs = []int64{11}
	binding, err = service.Save(ctx, binding.ID, input)
	require.NoError(t, err)
	for _, groupID := range []int64{33, 55} {
		var granted bool
		var status string
		require.NoError(t, database.QueryRow("SELECT EXISTS(SELECT 1 FROM user_allowed_groups WHERE user_id=12 AND group_id=$1)", groupID).Scan(&granted))
		require.False(t, granted)
		require.NoError(t, database.QueryRow("SELECT status FROM api_keys WHERE user_id=12 AND group_id=$1", groupID).Scan(&status))
		require.Equal(t, "inactive", status)
		require.NoError(t, service.Check(member(11, groupID), &Account{ID: 22}, nil))
		require.ErrorIs(t, service.Check(member(12, groupID), &Account{ID: 22}, nil), ErrDedicatedAccess)
	}
	var unrelatedStatus string
	require.NoError(t, database.QueryRow("SELECT status FROM api_keys WHERE user_id=12 AND group_id=77").Scan(&unrelatedStatus))
	require.Equal(t, "active", unrelatedStatus)
	input.ExpectedUpdatedAt = &binding.UpdatedAt
	input.GroupID = 55
	binding, err = service.Save(ctx, binding.ID, input)
	require.NoError(t, err)
	for _, groupID := range []int64{33, 55} {
		require.NoError(t, service.Check(member(11, groupID), &Account{ID: 22}, nil))
	}
	require.NoError(t, service.Revoke(ctx, binding.ID))
	for _, groupID := range []int64{33, 55} {
		require.ErrorIs(t, service.Check(member(11, groupID), nil, &groupID), ErrDedicatedAccess)
		require.ErrorIs(t, service.Check(member(11, groupID), &Account{ID: 22}, nil), ErrDedicatedAccess)
	}
	require.NoError(t, service.Delete(ctx, binding.ID))
	for _, groupID := range []int64{33, 55} {
		require.ErrorIs(t, service.Check(member(11, groupID), &Account{ID: 22}, nil), ErrDedicatedAccess)
	}
	_, err = service.Save(ctx, 0, input)
	require.NoError(t, err)
	for _, groupID := range []int64{33, 55} {
		require.NoError(t, service.Check(member(11, groupID), &Account{ID: 22}, nil))
		require.ErrorIs(t, service.Check(member(12, groupID), &Account{ID: 22}, nil), ErrDedicatedAccess)
	}
}
