package service

import (
	"context"
	"time"
)

type CustomGroupRateSchedule struct {
	RateScheduleEnabled bool                    `json:"rate_schedule_enabled"`
	RateSchedule        []CustomGroupRatePeriod `json:"rate_schedule"`
}

type CustomGroupRateSchedules struct {
	ServerTime time.Time                         `json:"server_time"`
	Schedules  map[int64]CustomGroupRateSchedule `json:"schedules"`
}

func (s *CustomGroupAccessService) AdminRateSchedules(ctx context.Context) (*CustomGroupRateSchedules, error) {
	policies, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}
	schedules := make(map[int64]CustomGroupRateSchedule)
	for groupID, policy := range policies {
		if policy.RateScheduleEnabled {
			schedules[groupID] = CustomGroupRateSchedule{RateScheduleEnabled: true, RateSchedule: policy.RateSchedule}
		}
	}
	return &CustomGroupRateSchedules{ServerTime: s.now(), Schedules: schedules}, nil
}
