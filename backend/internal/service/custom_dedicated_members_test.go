//go:build unit

package service

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestCustomDedicatedMembers(t *testing.T) {
	for _, members := range [][]int64{nil, {}, {0}, {-1}, {11, 11}, make([]int64, 101)} {
		require.ErrorIs(t, validateCustomDedicatedMembers(members), ErrDedicatedInput)
	}
	require.NoError(t, validateCustomDedicatedMembers([]int64{11, 12}))
	now := time.Now()
	binding := customDedicatedTestBinding(now)
	binding.UserIDs = []int64{11, 12}
	require.True(t, customDedicatedAllowed(binding, customDedicatedSubject{12, 33}, &Account{ID: 22}, now))
	require.False(t, customDedicatedAllowed(binding, customDedicatedSubject{13, 33}, &Account{ID: 22}, now))
}

func TestCustomDedicatedMemberRemovalFailureRollsBack(t *testing.T) {
	for _, failAt := range []string{"grant", "key", "config"} {
		t.Run(failAt, func(t *testing.T) {
			database, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func() { _ = database.Close() }()
			previous := customDedicatedTestBinding(time.Now())
			previous.UserIDs = []int64{11, 12}
			cache := &customDedicatedInvalidator{}
			service := &CustomDedicatedService{db: database, apiKeys: cache}
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT .* FOR UPDATE").WillReturnRows(customDedicatedTestRows(previous))
			grant := mock.ExpectExec("DELETE FROM user_allowed_groups WHERE .*group_id=.*user_id=").WithArgs(int64(33), int64(12), int64(22))
			if failAt == "grant" {
				grant.WillReturnError(errors.New("grant failure"))
			} else {
				grant.WillReturnResult(sqlmock.NewResult(0, 1))
				key := mock.ExpectExec("UPDATE api_keys SET status='inactive'.*group_id=.*user_id=").WithArgs(int64(33), int64(12), int64(22))
				if failAt == "key" {
					key.WillReturnError(errors.New("key failure"))
				} else {
					key.WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectQuery(regexp.QuoteMeta(customDedicatedIntegritySQL)).WithArgs(int64(44), int64(55), "[11]").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
					mock.ExpectQuery(regexp.QuoteMeta(customDedicatedConfigSQL)).WithArgs(int64(44), int64(55), "[11]").WillReturnRows(sqlmock.NewRows([]string{"issue", "resource_ids"}).AddRow("member_not_authorized", "11"))
				}
			}
			mock.ExpectRollback()
			_, err = service.Save(context.Background(), 1, CustomDedicatedInput{ExpectedUpdatedAt: &previous.UpdatedAt, UserIDs: []int64{11}, AccountID: 44, GroupID: 55, Label: previous.Label, ExpiresAt: previous.ExpiresAt})
			require.Error(t, err)
			require.Empty(t, cache.groups)
			require.Empty(t, cache.users)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCustomDedicatedDeleteRequiresRevoke(t *testing.T) {
	for _, revoked := range []bool{false, true} {
		database, mock, err := sqlmock.New()
		require.NoError(t, err)
		now := time.Now()
		binding := customDedicatedTestBinding(now)
		cache := &customDedicatedInvalidator{}
		service := &CustomDedicatedService{db: database, apiKeys: cache}
		if revoked {
			binding.RevokedAt = &now
		}
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT .* WHERE id=.*deleted_at IS NULL FOR UPDATE").WithArgs(int64(1)).WillReturnRows(customDedicatedTestRows(binding))
		if revoked {
			mock.ExpectExec("UPDATE custom_dedicated_accounts SET deleted_at").WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()
			require.NoError(t, service.Delete(context.Background(), 1))
			require.Equal(t, []int64{33}, cache.groups)
		} else {
			mock.ExpectRollback()
			require.ErrorIs(t, service.Delete(context.Background(), 1), ErrDedicatedDelete)
			require.Empty(t, cache.groups)
		}
		require.NoError(t, mock.ExpectationsWereMet())
		_ = database.Close()
	}
}

func TestCustomDedicatedSaveRequiresFreshVersionAndRestoreConfirmation(t *testing.T) {
	for _, scenario := range []string{"missing", "stale", "unconfirmed"} {
		t.Run(scenario, func(t *testing.T) {
			database, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func() { _ = database.Close() }()
			binding := customDedicatedTestBinding(time.Now())
			input := CustomDedicatedInput{UserID: 11, AccountID: 22, GroupID: 33, Label: binding.Label, ExpiresAt: binding.ExpiresAt}
			expectedError := ErrDedicatedStale
			if scenario == "stale" {
				older := binding.UpdatedAt.Add(-time.Second)
				input.ExpectedUpdatedAt = &older
				input.Reactivate = true
			} else if scenario == "unconfirmed" {
				input.ExpectedUpdatedAt = &binding.UpdatedAt
				binding.RevokedAt = &binding.UpdatedAt
				expectedError = ErrDedicatedRestore
			}
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT .* FOR UPDATE").WillReturnRows(customDedicatedTestRows(binding))
			mock.ExpectRollback()
			service := &CustomDedicatedService{db: database}
			_, err = service.Save(context.Background(), binding.ID, input)
			require.ErrorIs(t, err, expectedError)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCustomDedicatedReassignKeepsOldIsolation(t *testing.T) {
	database, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = database.Close() }()
	previous := customDedicatedTestBinding(time.Now())
	updated := previous
	updated.UserIDs = []int64{11, 12}
	updated.AccountID, updated.GroupID = 44, 55
	cache := &customDedicatedInvalidator{}
	service := &CustomDedicatedService{db: database, apiKeys: cache}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FOR UPDATE").WithArgs(int64(1)).WillReturnRows(customDedicatedTestRows(previous))
	mock.ExpectQuery(regexp.QuoteMeta(customDedicatedIntegritySQL)).WithArgs(int64(44), int64(55), "[11,12]").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("SELECT EXISTS.*deleted_at IS NULL").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("UPDATE custom_dedicated_accounts SET user_id").WithArgs(int64(11), updated.Label, updated.ExpiresAt, int64(1), int64(44), int64(55), "[11,12]").WillReturnRows(customDedicatedTestRows(updated))
	mock.ExpectExec("INSERT INTO custom_dedicated_accounts .*revoked_at, deleted_at").WithArgs(int64(11), int64(22), int64(33), previous.Label, previous.ExpiresAt, "[11]").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()
	saved, err := service.Save(context.Background(), 1, CustomDedicatedInput{ExpectedUpdatedAt: &previous.UpdatedAt, UserIDs: []int64{11, 12}, AccountID: 44, GroupID: 55, Label: updated.Label, ExpiresAt: updated.ExpiresAt})
	require.NoError(t, err)
	require.Equal(t, []int64{11, 12}, saved.UserIDs)
	require.Equal(t, []int64{55, 33}, cache.groups)
	require.NoError(t, mock.ExpectationsWereMet())
}
