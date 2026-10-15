package service

import (
	"context"
	"errors"
	"strconv"
)

type customGroupRPMValue struct {
	override *int
}

func (s *CustomGroupAccessService) prepareRPM(ctx context.Context, key *APIKey, user *User, groupID int64) error {
	user.UserGroupRPMOverride = nil
	user.UserGroupRPMOverrideLoaded = false
	if key.GroupID != nil && *key.GroupID == groupID && key.User != nil && (key.User.UserGroupRPMOverride != nil || key.User.UserGroupRPMOverrideLoaded) {
		user.UserGroupRPMOverride = key.User.UserGroupRPMOverride
		user.UserGroupRPMOverrideLoaded = true
		return nil
	}
	if s.rates == nil {
		return nil
	}
	now := s.now()
	cacheKey := customGroupAccessCacheKey("rpm", user.ID) + ":" + strconv.FormatInt(groupID, 10)
	value, err := s.cache.get(ctx, cacheKey, now, now.Add(customGroupAccessCacheTTL), func(ctx context.Context) (any, error) {
		override, err := s.rates.GetRPMOverrideByUserAndGroup(ctx, user.ID, groupID)
		return customGroupRPMValue{override: override}, err
	})
	if err != nil {
		return err
	}
	cached, ok := value.(customGroupRPMValue)
	if !ok {
		return errors.New("分组RPM缓存类型无效")
	}
	if cached.override != nil {
		override := *cached.override
		user.UserGroupRPMOverride = &override
	}
	user.UserGroupRPMOverrideLoaded = true
	return nil
}
