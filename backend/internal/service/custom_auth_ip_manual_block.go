package service

import (
	"context"
	"strings"
	"unicode/utf8"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

type AuthIPManualBlock struct {
	IP           string `json:"ip"`
	BlockMode    string `json:"block_mode"`
	BlockSeconds int    `json:"block_seconds"`
	Reason       string `json:"reason"`
}

func (guard *AuthIPBlacklistService) ManualBlock(ctx context.Context, input AuthIPManualBlock) (*AuthIPBlacklistEntry, error) {
	ip, err := NormalizeAuthIP(strings.TrimSpace(input.IP))
	if err != nil {
		return nil, err
	}
	input.IP = ip
	input.Reason = strings.TrimSpace(input.Reason)
	if input.Reason == "" || utf8.RuneCountInString(input.Reason) > 500 {
		return nil, infraerrors.BadRequest("IP_BLACKLIST_INVALID_REASON", "请填写1–500字的封禁原因")
	}
	if input.BlockMode != AuthIPBlockModeTimed && input.BlockMode != AuthIPBlockModePermanent {
		return nil, infraerrors.BadRequest("IP_BLACKLIST_INVALID_CONFIG", "请选择临时封禁或永久封禁")
	}
	if (input.BlockMode == AuthIPBlockModeTimed && (input.BlockSeconds < 60 || input.BlockSeconds > 86400 || input.BlockSeconds%60 != 0)) ||
		(input.BlockMode == AuthIPBlockModePermanent && input.BlockSeconds != 0) {
		return nil, infraerrors.BadRequest("IP_BLACKLIST_INVALID_CONFIG", "临时封禁须为1–1440分钟的整数，永久封禁不设置时长")
	}
	if guard == nil || guard.repo == nil {
		return nil, ErrAuthIPGuardUnavailable
	}
	entry, err := guard.repo.ManualBlock(ctx, input)
	return entry, authIPStorageError(err)
}
