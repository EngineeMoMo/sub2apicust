//go:build unit

package service

import (
	"context"
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
	saved, err := service.Save(context.Background(), 1, CustomDedicatedInput{UserIDs: []int64{11, 12}, AccountID: 44, GroupID: 55, Label: updated.Label, ExpiresAt: updated.ExpiresAt})
	require.NoError(t, err)
	require.Equal(t, []int64{11, 12}, saved.UserIDs)
	require.Equal(t, []int64{55, 33}, cache.groups)
	require.NoError(t, mock.ExpectationsWereMet())
}
