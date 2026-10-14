//go:build unit

package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type authIPManualRepo struct {
	AuthIPBlacklistRepository
	input *AuthIPManualBlock
	err   error
}

func (repo *authIPManualRepo) ManualBlock(_ context.Context, input AuthIPManualBlock) (*AuthIPBlacklistEntry, error) {
	repo.input = &input
	return &AuthIPBlacklistEntry{IP: input.IP, Source: "manual", Reason: input.Reason}, repo.err
}

func TestAuthIPManualBlockValidation(t *testing.T) {
	valid := AuthIPManualBlock{IP: " ::ffff:192.0.2.40 ", BlockMode: AuthIPBlockModePermanent, Reason: "  恶意扫描  "}
	repo := &authIPManualRepo{}
	guard := NewAuthIPBlacklistService(repo, nil)
	entry, err := guard.ManualBlock(context.Background(), valid)
	require.NoError(t, err)
	require.Equal(t, "192.0.2.40", entry.IP)
	require.Equal(t, "恶意扫描", repo.input.Reason)
	for _, change := range []func(*AuthIPManualBlock){
		func(input *AuthIPManualBlock) { input.IP = "192.0.2.0/24" },
		func(input *AuthIPManualBlock) { input.IP = "fe80::1%eth0" },
		func(input *AuthIPManualBlock) { input.Reason = " " },
		func(input *AuthIPManualBlock) { input.Reason = strings.Repeat("恶", 501) },
		func(input *AuthIPManualBlock) { input.BlockMode = "" },
		func(input *AuthIPManualBlock) { input.BlockSeconds = 60 },
		func(input *AuthIPManualBlock) { input.BlockMode = AuthIPBlockModeTimed; input.BlockSeconds = 0 },
		func(input *AuthIPManualBlock) { input.BlockMode = AuthIPBlockModeTimed; input.BlockSeconds = 61 },
		func(input *AuthIPManualBlock) { input.BlockMode = AuthIPBlockModeTimed; input.BlockSeconds = 86460 },
	} {
		input := valid
		change(&input)
		repo.input = nil
		_, err = guard.ManualBlock(context.Background(), input)
		require.Error(t, err)
		require.Nil(t, repo.input, "无效参数不能写入数据库")
	}
	for _, seconds := range []int{60, 86400} {
		input := valid
		input.BlockMode = AuthIPBlockModeTimed
		input.BlockSeconds = seconds
		input.Reason = strings.Repeat("恶", 500)
		_, err = guard.ManualBlock(context.Background(), input)
		require.NoError(t, err)
	}
	repo.err = ErrAuthIPAlreadyBlocked
	_, err = guard.ManualBlock(context.Background(), valid)
	require.ErrorIs(t, err, ErrAuthIPAlreadyBlocked)
	repo.err = errors.New("数据库故障")
	_, err = guard.ManualBlock(context.Background(), valid)
	require.ErrorIs(t, err, ErrAuthIPGuardUnavailable)
	_, err = NewAuthIPBlacklistService(nil, nil).ManualBlock(context.Background(), valid)
	require.ErrorIs(t, err, ErrAuthIPGuardUnavailable)
}
