package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const authIPWindowLua = `
local clock = redis.call('TIME')
local now = tonumber(clock[1]) * 1000 + math.floor(tonumber(clock[2]) / 1000)
local window = tonumber(ARGV[1])
redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', now - window)
local entered = tonumber(ARGV[3]) or now
if entered > now - window then
    redis.call('ZADD', KEYS[1], entered, ARGV[2])
end
local count = redis.call('ZCOUNT', KEYS[1], '-inf', now)
redis.call('PEXPIRE', KEYS[1], window)
local first = redis.call('ZRANGEBYSCORE', KEYS[1], '-inf', now, 'WITHSCORES', 'LIMIT', 0, 1)
local last = redis.call('ZREVRANGEBYSCORE', KEYS[1], now, '-inf', 'WITHSCORES', 'LIMIT', 0, 1)
return {count, now, tonumber(first[2]) or now, tonumber(last[2]) or now}
`

var authIPWindowScript = redis.NewScript(authIPWindowLua)

type authIPWindowCounter struct {
	redis *redis.Client
}

func NewAuthIPWindowCounter(client *redis.Client) service.AuthIPWindowCounter {
	return &authIPWindowCounter{redis: client}
}

func (counter *authIPWindowCounter) Count(ctx context.Context, ip string, generation int64, config service.AuthIPBlacklistConfig) (*service.AuthIPWindowCount, error) {
	return counter.CountAt(ctx, ip, generation, config, time.Time{})
}

func (counter *authIPWindowCounter) Time(ctx context.Context) (time.Time, error) {
	if counter.redis == nil {
		return time.Time{}, fmt.Errorf("认证计数 Redis 未配置")
	}
	instant, err := counter.redis.Time(ctx).Result()
	return instant.UTC(), err
}

func (counter *authIPWindowCounter) CountAt(ctx context.Context, ip string, generation int64, config service.AuthIPBlacklistConfig, entered time.Time) (*service.AuthIPWindowCount, error) {
	if counter.redis == nil {
		return nil, fmt.Errorf("认证计数 Redis 未配置")
	}
	target, err := service.AuthIPAggregationTarget(ip, config)
	if err != nil {
		return nil, err
	}
	score := ""
	if !entered.IsZero() {
		score = fmt.Sprint(entered.UnixMilli())
	}
	result, err := authIPWindowScript.Run(ctx, counter.redis,
		[]string{service.AuthIPCounterKey(target, generation, config.Version, config.NetworkState)},
		int64(config.WindowSeconds)*1000, uuid.NewString(), score).Int64Slice()
	if err != nil {
		return nil, err
	}
	if len(result) != 4 {
		return nil, fmt.Errorf("认证计数结果长度无效: %d", len(result))
	}
	return &service.AuthIPWindowCount{
		Count: int(result[0]),
		Now:   time.UnixMilli(result[1]).UTC(),
		First: time.UnixMilli(result[2]).UTC(),
		Last:  time.UnixMilli(result[3]).UTC(),
	}, nil
}
