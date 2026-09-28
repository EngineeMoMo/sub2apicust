//go:build unit

package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func customDedicatedTestBinding(now time.Time) CustomDedicatedBinding {
	return CustomDedicatedBinding{ID: 1, UserID: 11, AccountID: 22, GroupID: 33, Label: "Claude 专属", ExpiresAt: now.Add(time.Hour), UpdatedAt: now.Add(-time.Hour)}
}

func customDedicatedTestRows(binding CustomDedicatedBinding) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "user_id", "account_id", "group_id", "label", "expires_at", "revoked_at", "updated_at"}).AddRow(binding.ID, binding.UserID, binding.AccountID, binding.GroupID, binding.Label, binding.ExpiresAt, binding.RevokedAt, binding.UpdatedAt)
}

func TestCustomDedicatedAllowed(t *testing.T) {
	now := time.Now()
	binding := customDedicatedTestBinding(now)
	owner := customDedicatedSubject{UserID: 11, GroupID: 33}
	require.True(t, customDedicatedAllowed(binding, owner, &Account{ID: 22}, now))
	require.True(t, customDedicatedAllowed(binding, owner, nil, now))
	require.False(t, customDedicatedAllowed(binding, customDedicatedSubject{UserID: 12, GroupID: 33}, &Account{ID: 22}, now))
	require.False(t, customDedicatedAllowed(binding, customDedicatedSubject{UserID: 11, GroupID: 44}, &Account{ID: 22}, now))
	require.False(t, customDedicatedAllowed(binding, owner, &Account{ID: 44}, now))
	require.False(t, customDedicatedAllowed(binding, owner, &Account{ID: 22}, binding.ExpiresAt))
	binding.RevokedAt = &now
	require.False(t, customDedicatedAllowed(binding, owner, nil, now))
}

func TestCustomDedicatedCheck(t *testing.T) {
	now := time.Now()
	for _, test := range []struct {
		name                          string
		subject                       customDedicatedSubject
		authenticated, valid, allowed bool
		accountID                     int64
	}{
		{"owner", customDedicatedSubject{11, 33}, true, true, true, 22},
		{"other user", customDedicatedSubject{12, 33}, true, true, false, 22},
		{"shared pool", customDedicatedSubject{11, 44}, true, true, false, 22},
		{"fallback account", customDedicatedSubject{11, 33}, true, true, false, 44},
		{"missing identity", customDedicatedSubject{}, false, true, false, 22},
		{"configuration changed", customDedicatedSubject{11, 33}, true, false, false, 22},
	} {
		t.Run(test.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func() { _ = db.Close() }()
			binding := customDedicatedTestBinding(now)
			mock.ExpectQuery("SELECT .* FROM custom_dedicated_accounts WHERE group_id").WillReturnRows(customDedicatedTestRows(binding))
			if test.authenticated && test.subject.UserID == 11 && test.subject.GroupID == 33 && test.accountID == 22 {
				mock.ExpectQuery(regexp.QuoteMeta(customDedicatedIntegritySQL)).WithArgs(int64(22), int64(33), int64(11)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(test.valid))
			}
			ctx := context.Background()
			if test.authenticated {
				ctx = context.WithValue(ctx, customDedicatedContextKey{}, test.subject)
			}
			service := &CustomDedicatedService{db: db}
			err = service.Check(ctx, &Account{ID: test.accountID}, nil)
			if test.allowed {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, ErrDedicatedAccess)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCustomDedicatedFailClosedAndRelease(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	service := &CustomDedicatedService{db: db}
	released := 0
	selection := &AccountSelectionResult{Account: &Account{ID: 22}, ReleaseFunc: func() { released++ }}
	mock.ExpectQuery("SELECT .* FROM custom_dedicated_accounts").WillReturnError(errors.New("database unavailable"))
	service.guardSelection(context.Background(), nil, &selection, &err)
	require.ErrorIs(t, err, ErrDedicatedAccess)
	require.Nil(t, selection)
	require.Equal(t, 1, released)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomDedicatedUnreservedUnchanged(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	mock.ExpectQuery("SELECT .* FROM custom_dedicated_accounts").WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "account_id", "group_id", "label", "expires_at", "revoked_at", "updated_at"}))
	service := &CustomDedicatedService{db: db}
	require.NoError(t, service.Check(context.Background(), &Account{ID: 99}, nil))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomDedicatedViewOwnerFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT "+customDedicatedColumns+" FROM custom_dedicated_accounts WHERE id=$1 AND user_id=$2")).WithArgs(int64(1), int64(999)).WillReturnError(sql.ErrNoRows)
	service := &CustomDedicatedService{db: db}
	view, err := service.View(context.Background(), 999, 1, true)
	require.ErrorIs(t, err, ErrDedicatedNotFound)
	require.Nil(t, view)
	require.NoError(t, mock.ExpectationsWereMet())
}

type customDedicatedInvalidator struct{ groups []int64 }

func (cache *customDedicatedInvalidator) InvalidateAuthCacheByKey(context.Context, string)   {}
func (cache *customDedicatedInvalidator) InvalidateAuthCacheByUserID(context.Context, int64) {}
func (cache *customDedicatedInvalidator) InvalidateAuthCacheByGroupID(_ context.Context, id int64) {
	cache.groups = append(cache.groups, id)
}

func TestCustomDedicatedSaveAndRevoke(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	now := time.Now()
	binding := customDedicatedTestBinding(now)
	cache := &customDedicatedInvalidator{}
	service := &CustomDedicatedService{db: db, apiKeys: cache}
	input := CustomDedicatedInput{UserID: 11, AccountID: 22, GroupID: 33, Label: " Claude 专属 ", ExpiresAt: binding.ExpiresAt}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(customDedicatedIntegritySQL)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("SELECT EXISTS.*custom_dedicated_accounts").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("INSERT INTO custom_dedicated_accounts").WithArgs(int64(11), int64(22), int64(33), "Claude 专属", binding.ExpiresAt).WillReturnRows(customDedicatedTestRows(binding))
	mock.ExpectCommit()
	saved, err := service.Save(context.Background(), 0, input)
	require.NoError(t, err)
	require.Equal(t, "Claude 专属", saved.Label)
	mock.ExpectQuery("UPDATE custom_dedicated_accounts SET revoked_at").WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"group_id"}).AddRow(33))
	require.NoError(t, service.Revoke(context.Background(), 1))
	require.Equal(t, []int64{33, 33}, cache.groups)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomDedicatedSaveRejectsUnsafeConfigAndDuplicates(t *testing.T) {
	for _, valid := range []bool{false, true} {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		service := &CustomDedicatedService{db: db}
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(customDedicatedIntegritySQL)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(valid))
		if valid {
			mock.ExpectQuery("SELECT EXISTS.*custom_dedicated_accounts").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
		}
		mock.ExpectRollback()
		_, err = service.Save(context.Background(), 0, CustomDedicatedInput{UserID: 11, AccountID: 22, GroupID: 33, Label: "账号", ExpiresAt: time.Now().Add(time.Hour)})
		if valid {
			require.ErrorIs(t, err, ErrDedicatedConflict)
		} else {
			require.ErrorIs(t, err, ErrDedicatedConfig)
		}
		require.NoError(t, mock.ExpectationsWereMet())
		_ = db.Close()
	}
}

func TestCustomDedicatedQuotaFreshness(t *testing.T) {
	now := time.Now()
	reset := now.Add(time.Hour)
	sampled := now.Add(-time.Minute)
	fresh := customDedicatedQuotaWindow("five_hour", &UsageProgress{Utilization: 35, ResetsAt: &reset}, &sampled, now)
	require.NotNil(t, fresh.RemainingPercent)
	require.Equal(t, 65.0, *fresh.RemainingPercent)
	for _, used := range []float64{math.NaN(), math.Inf(1), -1} {
		require.Nil(t, customDedicatedQuotaWindow("five_hour", &UsageProgress{Utilization: used}, &sampled, now).RemainingPercent)
	}
	require.Nil(t, customDedicatedQuotaWindow("five_hour", nil, &sampled, now).RemainingPercent)
	require.Nil(t, customDedicatedQuotaWindow("five_hour", &UsageProgress{Utilization: 0}, nil, now).RemainingPercent)
	expired := now.Add(-time.Second)
	require.Nil(t, customDedicatedQuotaWindow("five_hour", &UsageProgress{Utilization: 0, ResetsAt: &expired}, &sampled, now).RemainingPercent)
	old := now.Add(-16 * time.Minute)
	require.Nil(t, customDedicatedQuotaWindow("five_hour", &UsageProgress{Utilization: 0}, &old, now).RemainingPercent)
	future := now.Add(2 * time.Minute)
	require.Nil(t, customDedicatedQuotaWindow("five_hour", &UsageProgress{Utilization: 0}, &future, now).RemainingPercent)
	exhausted := customDedicatedQuotaWindow("five_hour", &UsageProgress{Utilization: 110}, &sampled, now)
	require.Equal(t, 0.0, *exhausted.RemainingPercent)
}

func TestCustomDedicatedPassiveQuotaNeverFabricates(t *testing.T) {
	now := time.Now()
	for _, platform := range []string{PlatformAnthropic, PlatformOpenAI} {
		usage := customDedicatedPassiveUsage(&Account{Platform: platform, Extra: map[string]any{}}, now)
		require.Nil(t, usage.FiveHour)
		require.Nil(t, usage.SevenDay)
		require.Nil(t, usage.UpdatedAt)
	}
	for _, raw := range []any{nil, "bad", true, -1.0, math.NaN(), math.Inf(1)} {
		usage := customDedicatedPassiveUsage(&Account{Platform: PlatformOpenAI, Extra: map[string]any{"codex_5h_used_percent": raw}}, now)
		require.Nil(t, usage.FiveHour)
	}
	claude := customDedicatedPassiveUsage(&Account{Platform: PlatformAnthropic, Extra: map[string]any{"session_window_utilization": 0.25, "passive_usage_7d_utilization": 0.0, "passive_usage_sampled_at": now.Format(time.RFC3339)}}, now)
	require.Equal(t, 25.0, claude.FiveHour.Utilization)
	require.Equal(t, 0.0, claude.SevenDay.Utilization)
	require.NotNil(t, claude.UpdatedAt)
	codex := customDedicatedPassiveUsage(&Account{Platform: PlatformOpenAI, Extra: map[string]any{"codex_5h_used_percent": 65.0, "codex_usage_updated_at": now.Format(time.RFC3339)}}, now)
	require.Equal(t, 65.0, codex.FiveHour.Utilization)
}

type customDedicatedAccountReader struct {
	AccountRepository
	account *Account
}

func (reader *customDedicatedAccountReader) GetByID(context.Context, int64) (*Account, error) {
	return reader.account, nil
}

func TestCustomDedicatedViewOnlyReturnsWhitelist(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	now := time.Now()
	binding := customDedicatedTestBinding(now)
	reader := &customDedicatedAccountReader{account: &Account{ID: 22, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Name: "internal-account@example.com", Credentials: map[string]any{"access_token": "private-token"}, ErrorMessage: "secret upstream failure", Extra: map[string]any{"codex_5h_used_percent": 25.0, "codex_usage_updated_at": now.Format(time.RFC3339)}}}
	service := &CustomDedicatedService{db: db, accounts: reader}
	mock.ExpectQuery("SELECT .* WHERE id=").WithArgs(int64(1), int64(11)).WillReturnRows(customDedicatedTestRows(binding))
	mock.ExpectQuery(regexp.QuoteMeta(customDedicatedIntegritySQL)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("SELECT .* WHERE id=").WithArgs(int64(1), int64(11)).WillReturnRows(customDedicatedTestRows(binding))
	view, err := service.View(context.Background(), 11, 1, true)
	require.NoError(t, err)
	require.Equal(t, 75.0, *view.Windows[0].RemainingPercent)
	payload, err := json.Marshal(view)
	require.NoError(t, err)
	for _, secret := range []string{"private-token", "internal-account", "secret upstream", "credentials", "account_id", "user_id"} {
		require.NotContains(t, string(payload), secret)
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomDedicatedViewReassignmentDuringRead(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	binding := customDedicatedTestBinding(time.Now())
	service := &CustomDedicatedService{db: db, accounts: &customDedicatedAccountReader{account: &Account{ID: 22, Platform: PlatformAnthropic}}}
	mock.ExpectQuery("SELECT .* WHERE id=").WillReturnRows(customDedicatedTestRows(binding))
	mock.ExpectQuery(regexp.QuoteMeta(customDedicatedIntegritySQL)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("SELECT .* WHERE id=").WillReturnError(sql.ErrNoRows)
	view, err := service.View(context.Background(), 11, 1, true)
	require.ErrorIs(t, err, ErrDedicatedNotFound)
	require.Nil(t, view)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomDedicatedSimpleModeRejectsAssignment(t *testing.T) {
	service := &CustomDedicatedService{simpleMode: true}
	_, err := service.Save(context.Background(), 0, CustomDedicatedInput{})
	require.ErrorIs(t, err, ErrDedicatedMode)
}
