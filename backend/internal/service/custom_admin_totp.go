package service

import "context"

func (settings *SettingService) IsTotpAvailableForUser(ctx context.Context, user *User) bool {
	if user != nil && user.Role == RoleAdmin {
		return true
	}
	return settings.IsTotpEnabled(ctx)
}
