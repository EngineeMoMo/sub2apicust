//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type customAdminTotpCache struct {
	TotpCache
	session *TotpSetupSession
}

func (cache *customAdminTotpCache) SetSetupSession(_ context.Context, _ int64, session *TotpSetupSession, _ time.Duration) error {
	cache.session = session
	return nil
}

func (cache *customAdminTotpCache) GetSetupSession(context.Context, int64) (*TotpSetupSession, error) {
	return cache.session, nil
}

func TestCustomAdminTotpAvailableWhenGlobalDisabled(t *testing.T) {
	for _, role := range []string{RoleAdmin, RoleUser} {
		t.Run(role, func(t *testing.T) {
			user := &User{ID: 1, Email: "totp@example.com", Role: role}
			require.NoError(t, user.SetPassword("correct-password"))
			service, _ := newTotpVMService(t, user, false)
			cache := &customAdminTotpCache{}
			service.cache = cache
			status, err := service.GetStatus(context.Background(), user.ID)
			require.NoError(t, err)
			require.Equal(t, role == RoleAdmin, status.FeatureEnabled)
			setup, err := service.InitiateSetup(context.Background(), user.ID, "", "correct-password")
			if role == RoleAdmin {
				require.NoError(t, err)
				require.NotEmpty(t, setup.Secret)
				require.NotEmpty(t, cache.session.SetupToken)
				require.ErrorIs(t, service.CompleteSetup(context.Background(), user.ID, "000000", "wrong-token"), ErrTotpSetupExpired)
			} else {
				require.ErrorIs(t, err, ErrTotpNotEnabled)
				require.ErrorIs(t, service.CompleteSetup(context.Background(), user.ID, "000000", "wrong-token"), ErrTotpNotEnabled)
			}
		})
	}
}

type customStepUpSettingRepo struct {
	SettingRepository
	value string
	err   error
}

func (repo customStepUpSettingRepo) GetValue(context.Context, string) (string, error) {
	return repo.value, repo.err
}

func TestCustomStepUpSettingFailureDoesNotDisableProtection(t *testing.T) {
	for _, scenario := range []struct {
		name    string
		value   string
		err     error
		enabled bool
	}{
		{"enabled", "true", nil, true},
		{"disabled", "false", nil, false},
		{"missing", "", ErrSettingNotFound, false},
		{"database_error", "", errors.New("database unavailable"), true},
		{"corrupt_value", "invalid", nil, true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			settings := &SettingService{settingRepo: customStepUpSettingRepo{value: scenario.value, err: scenario.err}}
			require.Equal(t, scenario.enabled, settings.IsStepUpEnabled(context.Background()))
		})
	}
}
