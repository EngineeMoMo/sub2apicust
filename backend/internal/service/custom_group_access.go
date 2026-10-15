package service

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrGroupAccessRequirements = infraerrors.Forbidden("GROUP_ACCESS_REQUIREMENTS", "当前月净充值或账户余额未达到该分组的开放条件")
	ErrGroupAccessExpired      = infraerrors.Forbidden("GROUP_ACCESS_EXPIRED", "该分组已到期，且没有可用的到期后分组")
	ErrGroupAccessPolicyStale  = infraerrors.Conflict("GROUP_ACCESS_POLICY_STALE", "开放规则已被修改，请重新打开后保存")
	ErrGroupAccessRouteChanged = infraerrors.Forbidden("GROUP_ACCESS_ROUTE_CHANGED", "分组路由已变化，请重新连接；无需修改 Key 或分组绑定")
)

type CustomGroupAccessPolicy struct {
	GroupID                int64                   `json:"group_id"`
	MonthlyRechargeEnabled bool                    `json:"monthly_recharge_enabled"`
	MonthlyRechargeMin     float64                 `json:"monthly_recharge_min"`
	BalanceEnabled         bool                    `json:"balance_enabled"`
	BalanceMin             float64                 `json:"balance_min"`
	ExpiresAt              *time.Time              `json:"expires_at"`
	FallbackGroupID        *int64                  `json:"fallback_group_id"`
	UpdatedAt              *time.Time              `json:"updated_at"`
	RateScheduleEnabled    bool                    `json:"rate_schedule_enabled"`
	RateSchedule           []CustomGroupRatePeriod `json:"rate_schedule"`
}

type customGroupAccessFacts struct {
	Balance         float64
	MonthlyRecharge float64
}

type customGroupAccessStore interface {
	List(context.Context) (map[int64]CustomGroupAccessPolicy, error)
	Facts(context.Context, int64, time.Time, time.Time) (customGroupAccessFacts, error)
	Save(context.Context, CustomGroupAccessPolicy) (*CustomGroupAccessPolicy, error)
}

type CustomGroupAccessService struct {
	store         customGroupAccessStore
	groups        GroupRepository
	users         UserRepository
	subscriptions UserSubscriptionRepository
	now           func() time.Time
	cache         *customGroupAccessCache
	stop          func()
}

func NewCustomGroupAccessService(db *sql.DB, groups GroupRepository, users UserRepository, subscriptions UserSubscriptionRepository, apiKeys *APIKeyService) *CustomGroupAccessService {
	result := &CustomGroupAccessService{store: &customGroupAccessSQLStore{db: db}, groups: groups, users: users, subscriptions: subscriptions, now: time.Now}
	apiKeys.customGroupAccess = result
	return result
}

func (s *CustomGroupAccessService) Get(ctx context.Context, groupID int64) (*CustomGroupAccessPolicy, error) {
	if _, err := s.groups.GetByID(ctx, groupID); err != nil {
		return nil, err
	}
	policies, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}
	policy := policies[groupID]
	policy.GroupID = groupID
	return &policy, nil
}

func validateCustomGroupAccessPolicy(policy CustomGroupAccessPolicy) error {
	for _, threshold := range []struct {
		enabled bool
		amount  float64
	}{
		{policy.MonthlyRechargeEnabled, policy.MonthlyRechargeMin},
		{policy.BalanceEnabled, policy.BalanceMin},
	} {
		if math.IsNaN(threshold.amount) || math.IsInf(threshold.amount, 0) || threshold.amount < 0 || threshold.amount > 1e12 || math.Abs(threshold.amount-math.Round(threshold.amount*100)/100) > 1e-9 || (threshold.enabled && threshold.amount <= 0) {
			return infraerrors.BadRequest("GROUP_ACCESS_INVALID_THRESHOLD", "门槛金额须为非负数，最多两位小数；开启时须大于零")
		}
	}
	if policy.GroupID <= 0 || (policy.FallbackGroupID != nil && (*policy.FallbackGroupID <= 0 || *policy.FallbackGroupID == policy.GroupID)) {
		return infraerrors.BadRequest("GROUP_ACCESS_INVALID_FALLBACK", "到期后分组必须是其他有效分组")
	}
	if (policy.ExpiresAt == nil) != (policy.FallbackGroupID == nil) || (policy.ExpiresAt != nil && policy.ExpiresAt.IsZero()) {
		return infraerrors.BadRequest("GROUP_ACCESS_INVALID_EXPIRY", "启用到期时间时须同时指定到期后分组")
	}
	return validateCustomGroupRateSchedule(policy)
}

func (s *CustomGroupAccessService) Save(ctx context.Context, policy CustomGroupAccessPolicy) (*CustomGroupAccessPolicy, error) {
	if err := validateCustomGroupAccessPolicy(policy); err != nil {
		return nil, err
	}
	source, err := s.groups.GetByID(ctx, policy.GroupID)
	if err != nil {
		return nil, err
	}
	if policy.FallbackGroupID != nil {
		target, err := s.groups.GetByID(ctx, *policy.FallbackGroupID)
		if err != nil {
			return nil, err
		}
		if !target.IsActive() || target.Platform != source.Platform {
			return nil, infraerrors.BadRequest("GROUP_ACCESS_INVALID_FALLBACK", "到期后分组须启用且与原分组属于同一平台")
		}
	}
	saved, err := s.store.Save(ctx, policy)
	if err == nil {
		s.cache.invalidate("policies")
	}
	return saved, err
}

func customGroupAccessMonth(now time.Time) (time.Time, time.Time) {
	local := now.In(time.FixedZone("Asia/Shanghai", 8*60*60))
	start := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, local.Location())
	return start, start.AddDate(0, 1, 0)
}

func (policy CustomGroupAccessPolicy) expired(now time.Time) bool {
	return policy.ExpiresAt != nil && !now.Before(*policy.ExpiresAt)
}

func (policy CustomGroupAccessPolicy) hasThresholds() bool {
	return policy.MonthlyRechargeEnabled || policy.BalanceEnabled
}

func (policy CustomGroupAccessPolicy) qualifies(facts customGroupAccessFacts) bool {
	return (!policy.MonthlyRechargeEnabled || facts.MonthlyRecharge >= policy.MonthlyRechargeMin) && (!policy.BalanceEnabled || facts.Balance >= policy.BalanceMin)
}

func (s *CustomGroupAccessService) DeniedGroupIDs(ctx context.Context, userID int64) (map[int64]bool, error) {
	policies, err := s.loadPolicies(ctx)
	if err != nil {
		return nil, err
	}
	denied := make(map[int64]bool)
	now := s.now()
	monthly := false
	for _, policy := range policies {
		monthly = monthly || (policy.MonthlyRechargeEnabled && !policy.expired(now))
	}
	var facts *customGroupAccessFacts
	for groupID, policy := range policies {
		if policy.expired(now) {
			denied[groupID] = true
			continue
		}
		if !policy.hasThresholds() {
			continue
		}
		if userID <= 0 {
			denied[groupID] = true
			continue
		}
		if facts == nil {
			loaded, err := s.loadFacts(ctx, userID, monthly)
			if err != nil {
				return nil, err
			}
			facts = &loaded
		}
		if !policy.qualifies(*facts) {
			denied[groupID] = true
		}
	}
	return denied, nil
}

func (s *CustomGroupAccessService) canUse(ctx context.Context, user *User, group *Group) bool {
	if group.IsSubscriptionType() {
		_, err := s.loadSubscription(ctx, user.ID, group.ID)
		return err == nil
	}
	return user.CanBindGroup(group.ID, group.IsExclusive)
}

func (s *CustomGroupAccessService) PrepareKey(ctx context.Context, key *APIKey) (*APIKey, error) {
	if key == nil || key.GroupID == nil {
		return key, nil
	}
	policies, err := s.loadPolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("load group access policies: %w", err)
	}
	initial, exists := policies[*key.GroupID]
	if !exists || (!initial.hasThresholds() && initial.ExpiresAt == nil && !initial.RateScheduleEnabled) {
		return key, nil
	}
	user, err := s.loadUser(ctx, key.UserID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive() {
		return nil, ErrUserNotActive
	}
	now := s.now()
	groupID := *key.GroupID
	seen := make(map[int64]bool)
	var facts *customGroupAccessFacts
	for len(seen) < 32 {
		if seen[groupID] {
			return nil, ErrGroupAccessExpired
		}
		seen[groupID] = true
		group, err := s.loadGroup(ctx, groupID)
		if err != nil {
			return nil, err
		}
		if !group.IsActive() || (key.Group != nil && group.Platform != key.Group.Platform) {
			return nil, ErrGroupAccessExpired
		}
		if !group.IsSubscriptionType() && !user.CanBindGroup(group.ID, group.IsExclusive) {
			return nil, ErrGroupNotAllowed
		}
		policy := policies[groupID]
		if policy.expired(now) {
			if policy.FallbackGroupID == nil {
				return nil, ErrGroupAccessExpired
			}
			groupID = *policy.FallbackGroupID
			continue
		}
		if !s.canUse(ctx, user, group) {
			return nil, ErrGroupNotAllowed
		}
		if policy.hasThresholds() {
			if facts == nil {
				loaded, err := s.loadFacts(ctx, user.ID, policy.MonthlyRechargeEnabled)
				if err != nil {
					return nil, err
				}
				facts = &loaded
				user.Balance = facts.Balance
			}
			if !policy.qualifies(*facts) {
				return nil, ErrGroupAccessRequirements
			}
		}
		prepared := *key
		group.RateMultiplier = policy.rateAt(now, group.RateMultiplier)
		prepared.GroupID = &groupID
		prepared.Group = group
		prepared.User = user
		return &prepared, nil
	}
	return nil, ErrGroupAccessExpired
}

func (s *APIKeyService) PrepareCustomGroupAccess(ctx context.Context, key *APIKey) (*APIKey, error) {
	if s == nil || s.customGroupAccess == nil {
		return key, nil
	}
	return s.customGroupAccess.PrepareKey(ctx, key)
}

func (s *APIKeyService) CustomDeniedGroupIDs(ctx context.Context, userID int64) (map[int64]bool, error) {
	if s == nil || s.customGroupAccess == nil {
		return nil, nil
	}
	return s.customGroupAccess.DeniedGroupIDs(ctx, userID)
}

func (s *APIKeyService) checkCustomGroupBinding(ctx context.Context, userID, groupID int64) error {
	denied, err := s.CustomDeniedGroupIDs(ctx, userID)
	if err != nil {
		return err
	}
	if denied[groupID] {
		return ErrGroupAccessRequirements
	}
	return nil
}

func (s *APIKeyService) ValidateCustomGroupAccessContinuation(ctx context.Context, connected *APIKey) error {
	if s == nil || s.customGroupAccess == nil || connected == nil || connected.GroupID == nil {
		return nil
	}
	latest, err := s.GetByKey(ctx, connected.Key)
	if err != nil {
		return err
	}
	prepared, err := s.PrepareCustomGroupAccess(ctx, latest)
	if err != nil {
		return err
	}
	if prepared.GroupID == nil || *prepared.GroupID != *connected.GroupID || (prepared.Group != nil && connected.Group != nil && prepared.Group.RateMultiplier != connected.Group.RateMultiplier) {
		return ErrGroupAccessRouteChanged
	}
	return nil
}
