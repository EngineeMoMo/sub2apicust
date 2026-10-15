package service

import (
	"context"
	"math"
	"sort"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

type CustomGroupRatePeriod struct {
	StartsAt       time.Time `json:"starts_at"`
	EndsAt         time.Time `json:"ends_at"`
	RateMultiplier float64   `json:"rate_multiplier"`
}

func validateCustomGroupRateSchedule(policy CustomGroupAccessPolicy) error {
	if len(policy.RateSchedule) > 32 || (policy.RateScheduleEnabled && len(policy.RateSchedule) == 0) {
		return infraerrors.BadRequest("GROUP_RATE_INVALID_SCHEDULE", "活动倍率须设置1至32个时段")
	}
	periods := append([]CustomGroupRatePeriod(nil), policy.RateSchedule...)
	sort.Slice(periods, func(first, second int) bool { return periods[first].StartsAt.Before(periods[second].StartsAt) })
	for index, period := range periods {
		if period.StartsAt.IsZero() || !period.EndsAt.After(period.StartsAt) || math.IsNaN(period.RateMultiplier) || math.IsInf(period.RateMultiplier, 0) || period.RateMultiplier < 0 || period.RateMultiplier > 1000 {
			return infraerrors.BadRequest("GROUP_RATE_INVALID_SCHEDULE", "活动时段须有有效起止时间，倍率须在0至1000之间")
		}
		if index > 0 && period.StartsAt.Before(periods[index-1].EndsAt) {
			return infraerrors.BadRequest("GROUP_RATE_INVALID_SCHEDULE", "活动倍率时段不能重叠")
		}
		if policy.RateScheduleEnabled && policy.ExpiresAt != nil && period.EndsAt.After(*policy.ExpiresAt) {
			return infraerrors.BadRequest("GROUP_RATE_INVALID_SCHEDULE", "活动时段不能晚于分组到期时间")
		}
	}
	return nil
}

func (policy CustomGroupAccessPolicy) rateAt(now time.Time, base float64) float64 {
	if policy.RateScheduleEnabled {
		for _, period := range policy.RateSchedule {
			if !now.Before(period.StartsAt) && now.Before(period.EndsAt) {
				return period.RateMultiplier
			}
		}
	}
	return base
}

func (s *APIKeyService) CustomGroupRate(ctx context.Context, groupID int64, base float64) (float64, error) {
	if s == nil || s.customGroupAccess == nil {
		return base, nil
	}
	return s.customGroupAccess.groupRate(ctx, groupID, base)
}

func (s *CustomGroupAccessService) groupRate(ctx context.Context, groupID int64, base float64) (float64, error) {
	policies, err := s.loadPolicies(ctx)
	if err != nil {
		return 0, err
	}
	return policies[groupID].rateAt(s.now(), base), nil
}

type customGroupRateUseDefault struct{}

func customGroupCachedRate(value any, currentDefault float64) (float64, bool) {
	switch value := value.(type) {
	case float64:
		return value, true
	case customGroupRateUseDefault:
		return currentDefault, true
	default:
		return 0, false
	}
}
