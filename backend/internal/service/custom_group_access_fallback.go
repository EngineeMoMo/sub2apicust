package service

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
)

func (s *CustomGroupAccessService) PrepareFallback(ctx context.Context, key *APIKey, target *Group) (*APIKey, error) {
	if key == nil || key.UserID <= 0 || target == nil || !target.IsActive() {
		return nil, ErrGroupNotAllowed
	}
	user, err := s.loadUser(ctx, key.UserID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive() {
		return nil, ErrUserNotActive
	}
	if !s.canUse(ctx, user, target) {
		return nil, ErrGroupNotAllowed
	}
	if err := s.prepareRPM(ctx, key, user, target.ID); err != nil {
		return nil, err
	}
	prepared := *key
	group := *target
	prepared.GroupID, prepared.Group, prepared.User = &group.ID, &group, user
	return s.PrepareKey(ctx, &prepared)
}

func (s *APIKeyService) PrepareCustomGroupFallback(ctx context.Context, key *APIKey, target *Group) (*APIKey, error) {
	if s == nil || s.customGroupAccess == nil {
		prepared := *key
		prepared.GroupID, prepared.Group = &target.ID, target
		return &prepared, nil
	}
	return s.customGroupAccess.PrepareFallback(ctx, key, target)
}

func (s *APIKeyService) CustomGroupFallbackSubscription(ctx context.Context, key *APIKey) (*UserSubscription, error) {
	if key == nil || key.Group == nil || !key.Group.IsSubscriptionType() || s == nil || s.customGroupAccess == nil {
		return nil, nil
	}
	return s.customGroupAccess.loadSubscription(ctx, key.UserID, key.Group.ID)
}

func (s *GatewayService) PrepareCustomGatewayGroup(ctx context.Context, key *APIKey) (*APIKey, error) {
	if s == nil || s.customGroupAccess == nil || key == nil || key.GroupID == nil {
		return key, nil
	}
	return s.customGroupAccess.prepareClientFallback(ctx, key)
}

func (s *APIKeyService) PrepareCustomGroupClientFallback(ctx context.Context, key *APIKey) (*APIKey, error) {
	if s == nil || s.customGroupAccess == nil || key == nil {
		return key, nil
	}
	return s.customGroupAccess.prepareClientFallback(ctx, key)
}

func (s *CustomGroupAccessService) prepareClientFallback(ctx context.Context, key *APIKey) (*APIKey, error) {
	if key.Group == nil || !key.Group.ClaudeCodeOnly || IsClaudeCodeClient(ctx) {
		return key, nil
	}
	if platform, _ := ctx.Value(ctxkey.ForcePlatform).(string); platform != "" {
		return key, nil
	}
	visited := make(map[int64]bool)
	prepared := key
	for len(visited) < 32 {
		group := prepared.Group
		if visited[group.ID] {
			return nil, ErrGroupAccessExpired
		}
		visited[group.ID] = true
		if !group.ClaudeCodeOnly {
			return prepared, nil
		}
		if group.FallbackGroupID == nil {
			return nil, ErrClaudeCodeOnly
		}
		target, err := s.loadGroup(ctx, *group.FallbackGroupID)
		if err != nil {
			return nil, err
		}
		prepared, err = s.PrepareFallback(ctx, prepared, target)
		if err != nil {
			return nil, err
		}
	}
	return nil, ErrGroupAccessExpired
}
