package handler

import "github.com/Wei-Shaw/sub2api/ent"

func (handler *AuthHandler) AuthIPBlacklistEntClient() *ent.Client {
	if handler == nil || handler.authService == nil {
		return nil
	}
	return handler.authService.EntClient()
}
