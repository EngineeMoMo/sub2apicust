//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegistrationIPErrorsSurviveAuthFlows(t *testing.T) {
	for _, want := range []error{ErrRegistrationIPLimit, ErrRegistrationIPUnavailable} {
		for _, flow := range []string{"email", "oauth-legacy", "oauth-token-pair", "oauth-verified-email", "oauth-auto-email"} {
			t.Run(want.Error()+flow, func(t *testing.T) {
				repo := &userRepoStub{createErr: want}
				svc := newAuthService(repo, map[string]string{SettingKeyRegistrationEnabled: "true"}, nil, nil)
				svc.refreshTokenCache = &refreshTokenCacheStub{}
				ctx := WithRegistrationIP(context.Background(), "192.0.2.1")
				var err error
				switch flow {
				case "email":
					_, _, err = svc.Register(ctx, "first@gmail.com", "password")
				case "oauth-legacy":
					_, _, err = svc.LoginOrRegisterOAuth(ctx, "first@gmail.com", "first")
				case "oauth-token-pair":
					_, _, err = svc.LoginOrRegisterOAuthWithTokenPair(ctx, "first@gmail.com", "first", "", "", "github")
				case "oauth-verified-email":
					_, _, err = svc.RegisterVerifiedOAuthEmailAccount(ctx, "first@gmail.com", "password", "", "github")
				case "oauth-auto-email":
					_, err = svc.createEmailOAuthUser(ctx, "first@gmail.com", "first", "github", "", "")
				}
				require.ErrorIs(t, err, want)
			})
		}
	}
}

func TestRegistrationIPSettingDefaultsAndStoredOverride(t *testing.T) {
	svc := newAuthService(&userRepoStub{}, map[string]string{}, nil, nil)
	for _, scenario := range []struct {
		name   string
		values map[string]string
		want   bool
	}{
		{"missing defaults enabled", map[string]string{}, true},
		{"enabled", map[string]string{SettingKeyRegistrationIPLimitEnabled: "true"}, true},
		{"explicitly disabled", map[string]string{SettingKeyRegistrationIPLimitEnabled: "false"}, false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			settings := svc.settingService.parseSettings(scenario.values)
			require.Equal(t, scenario.want, settings.RegistrationIPLimitEnabled)
		})
	}
}
