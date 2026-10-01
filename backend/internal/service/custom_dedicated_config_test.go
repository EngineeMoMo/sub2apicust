//go:build unit

package service

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestCustomDedicatedConfigFailureDetails(t *testing.T) {
	for _, issue := range []string{"group_not_exclusive", "account_unsafe_groups", "other_authorized_users", "member_not_authorized", "non_member_keys"} {
		t.Run(issue, func(t *testing.T) {
			database, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func() { _ = database.Close() }()
			cache := &customDedicatedInvalidator{}
			service := &CustomDedicatedService{db: database, apiKeys: cache}
			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta(customDedicatedIntegritySQL)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
			mock.ExpectQuery(regexp.QuoteMeta(customDedicatedConfigSQL)).WithArgs(int64(22), int64(33), "[11]").WillReturnRows(sqlmock.NewRows([]string{"issue", "resource_ids"}).AddRow(issue, "55"))
			mock.ExpectRollback()
			binding, err := service.Save(context.Background(), 0, CustomDedicatedInput{UserID: 11, AccountID: 22, GroupID: 33, Label: "专属账号", ExpiresAt: time.Now().Add(time.Hour)})
			require.Nil(t, binding)
			require.ErrorIs(t, err, ErrDedicatedConfig)
			var status *infraerrors.ApplicationError
			require.ErrorAs(t, err, &status)
			require.Equal(t, issue, status.Metadata["config_issue"])
			require.Equal(t, "55", status.Metadata["resource_ids"])
			require.Contains(t, status.Message, customDedicatedConfigMessages[issue])
			require.Empty(t, cache.groups)
			require.Empty(t, cache.users)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCustomDedicatedConfigDiagnosticFailure(t *testing.T) {
	for _, failure := range []error{sql.ErrNoRows, errors.New("database failure")} {
		database, mock, err := sqlmock.New()
		require.NoError(t, err)
		mock.ExpectQuery(regexp.QuoteMeta(customDedicatedConfigSQL)).WillReturnError(failure)
		err = customDedicatedConfigError(context.Background(), database, CustomDedicatedBinding{AccountID: 22, GroupID: 33, UserID: 11})
		if errors.Is(failure, sql.ErrNoRows) {
			require.ErrorIs(t, err, ErrDedicatedConfig)
		} else {
			require.ErrorIs(t, err, failure)
		}
		mock.ExpectClose()
		require.NoError(t, database.Close())
		require.NoError(t, mock.ExpectationsWereMet())
	}
}
