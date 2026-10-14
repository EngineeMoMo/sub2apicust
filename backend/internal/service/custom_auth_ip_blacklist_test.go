//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type authIPServiceTestRepo struct {
	AuthIPBlacklistRepository
	snapshot *AuthIPBlacklistSnapshot
	failure  error
	blocks   int
	ip       string
}

func (repo *authIPServiceTestRepo) Snapshot(_ context.Context, ip string) (*AuthIPBlacklistSnapshot, error) {
	repo.ip = ip
	return repo.snapshot, repo.failure
}

func (repo *authIPServiceTestRepo) Block(_ context.Context, ip, _ string, _ AuthIPBlacklistSnapshot, events AuthIPWindowCount) (*AuthIPBlacklistDecision, error) {
	repo.blocks++
	if repo.failure != nil {
		return nil, repo.failure
	}
	return &AuthIPBlacklistDecision{NewlyBlocked: true, Entry: &AuthIPBlacklistEntry{IP: ip, Blocked: true, TriggerCount: events.Count}}, nil
}

type authIPServiceTestCounter struct {
	calls   int
	count   int
	failure error
}

func (counter *authIPServiceTestCounter) Count(context.Context, string, int64, AuthIPBlacklistConfig) (*AuthIPWindowCount, error) {
	counter.calls++
	return &AuthIPWindowCount{Count: counter.count}, counter.failure
}

func TestAuthIPBlacklistServiceDecisions(t *testing.T) {
	ctx := context.Background()
	repo := &authIPServiceTestRepo{snapshot: &AuthIPBlacklistSnapshot{Config: AuthIPBlacklistConfig{Enabled: true, Threshold: 10, WindowSeconds: 7200, Version: 1, BlockMode: AuthIPBlockModeTimed}}}
	counter := &authIPServiceTestCounter{count: 9}
	guard := NewAuthIPBlacklistService(repo, counter)
	_, err := guard.Check(ctx, "::ffff:192.0.2.1", "/api/v1/auth/login", true)
	require.NoError(t, err)
	require.Equal(t, "192.0.2.1", repo.ip)
	require.Zero(t, repo.blocks)
	counter.count = 10
	decision, err := guard.Check(ctx, "192.0.2.1", "/api/v1/auth/register", true)
	require.ErrorIs(t, err, ErrAuthIPBlocked)
	require.True(t, decision.NewlyBlocked)
	require.Equal(t, 1, repo.blocks)
	repo.snapshot.Entry = decision.Entry
	repo.snapshot.Config.Enabled = false
	_, err = guard.Check(ctx, "192.0.2.1", "/api/v1/auth/login", true)
	require.ErrorIs(t, err, ErrAuthIPBlocked)
	require.Equal(t, 2, counter.calls)
	repo.snapshot.Entry = nil
	_, err = guard.Check(ctx, "192.0.2.1", "/api/v1/auth/login", true)
	require.NoError(t, err)
	repo.snapshot.Config.Enabled = true
	_, err = guard.Check(ctx, "192.0.2.1", "/api/v1/auth/oauth/google/start", false)
	require.NoError(t, err)
	require.Equal(t, 2, counter.calls)
	counter.failure = errors.New("redis offline")
	_, err = guard.Check(ctx, "192.0.2.1", "/api/v1/auth/login", true)
	require.ErrorIs(t, err, ErrAuthIPCounterUnavailable)
	require.Equal(t, 1, repo.blocks)
	repo.failure = errors.New("database offline")
	_, err = guard.Check(ctx, "192.0.2.1", "/api/v1/auth/login", true)
	require.ErrorIs(t, err, ErrAuthIPGuardUnavailable)
	var unavailable *AuthIPBlacklistService
	_, err = unavailable.Check(ctx, "192.0.2.1", "/api/v1/auth/login", true)
	require.ErrorIs(t, err, ErrAuthIPGuardUnavailable)
}

func TestAuthIPBlacklistValidation(t *testing.T) {
	for _, input := range []string{"192.0.2.1/32", "example.org", "fe80::1%eth0", ""} {
		_, err := NormalizeAuthIP(input)
		require.Error(t, err)
	}
	for _, config := range []AuthIPBlacklistConfig{
		{Threshold: 0, WindowSeconds: 7200, Version: 1},
		{Threshold: 1001, WindowSeconds: 7200, Version: 1},
		{Threshold: 10, WindowSeconds: 61, Version: 1},
		{Threshold: 10, WindowSeconds: 86460, Version: 1},
		{Threshold: 10, WindowSeconds: 7200, Version: 0},
	} {
		config.BlockMode = AuthIPBlockModeTimed
		require.Error(t, ValidateAuthIPBlacklistConfig(config))
	}
	require.NoError(t, ValidateAuthIPBlacklistConfig(AuthIPBlacklistConfig{Threshold: 1, WindowSeconds: 60, Version: 1, BlockMode: AuthIPBlockModeTimed}))
	require.NoError(t, ValidateAuthIPBlacklistConfig(AuthIPBlacklistConfig{Threshold: 1000, WindowSeconds: 86400, Version: 1, BlockMode: AuthIPBlockModePermanent}))
}

func TestAuthIPBlacklistBlockModeValidation(t *testing.T) {
	for _, mode := range []string{"", "forever", "TIMED"} {
		err := ValidateAuthIPBlacklistConfig(AuthIPBlacklistConfig{Threshold: 10, WindowSeconds: 7200, Version: 1, BlockMode: mode})
		require.Error(t, err, "省略或非法模式不能静默改为临时封禁")
	}
	for _, mode := range []string{AuthIPBlockModeTimed, AuthIPBlockModePermanent} {
		for _, seconds := range []int{-60, 1, 61, 86460} {
			require.Error(t, ValidateAuthIPBlacklistConfig(AuthIPBlacklistConfig{Threshold: 10, WindowSeconds: 7200, Version: 1, BlockMode: mode, BlockSeconds: seconds}))
		}
		for _, seconds := range []int{0, 60, 900, 86400} {
			require.NoError(t, ValidateAuthIPBlacklistConfig(AuthIPBlacklistConfig{Threshold: 10, WindowSeconds: 7200, Version: 1, BlockMode: mode, BlockSeconds: seconds}))
		}
	}
}
