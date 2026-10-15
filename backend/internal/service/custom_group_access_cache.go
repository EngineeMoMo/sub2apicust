package service

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

const customGroupAccessCacheTTL = 30 * time.Second
const customGroupAccessCacheLimit = 4096

type customGroupAccessCacheEntry struct {
	value   any
	expires time.Time
}

type customGroupAccessCache struct {
	mu           sync.Mutex
	entries      map[string]*customGroupAccessCacheEntry
	generation   uint64
	healthyUntil time.Time
	loads        singleflight.Group
}

func newCustomGroupAccessCache() *customGroupAccessCache {
	return &customGroupAccessCache{entries: make(map[string]*customGroupAccessCacheEntry)}
}

func (c *customGroupAccessCache) health(connected bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !connected || !time.Now().Before(c.healthyUntil) {
		c.entries = make(map[string]*customGroupAccessCacheEntry)
		c.generation++
	}
	c.healthyUntil = time.Time{}
	if connected {
		c.healthyUntil = time.Now().Add(3 * time.Second)
	}
}

func (c *customGroupAccessCache) invalidate(key string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}

func (c *customGroupAccessCache) get(ctx context.Context, key string, now, expires time.Time, load func(context.Context) (any, error)) (any, error) {
	if c == nil {
		return load(ctx)
	}
	c.mu.Lock()
	healthy := time.Now().Before(c.healthyUntil)
	entry := c.entries[key]
	if healthy && entry != nil && entry.value != nil && now.Before(entry.expires) {
		value := entry.value
		c.mu.Unlock()
		return value, nil
	}
	c.mu.Unlock()
	if !healthy {
		return load(ctx)
	}
	result := c.loads.DoChan(key, func() (any, error) {
		loadCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		for attempt := 0; attempt < 3; attempt++ {
			c.mu.Lock()
			if !time.Now().Before(c.healthyUntil) {
				c.mu.Unlock()
				return load(loadCtx)
			}
			entry := c.entries[key]
			if entry != nil && entry.value != nil && now.Before(entry.expires) {
				value := entry.value
				c.mu.Unlock()
				return value, nil
			}
			if len(c.entries) >= customGroupAccessCacheLimit {
				c.entries = make(map[string]*customGroupAccessCacheEntry)
				c.generation++
			}
			token := &customGroupAccessCacheEntry{}
			c.entries[key] = token
			generation := c.generation
			c.mu.Unlock()
			value, err := load(loadCtx)
			c.mu.Lock()
			current := c.entries[key] == token && c.generation == generation && time.Now().Before(c.healthyUntil)
			if err != nil {
				if c.entries[key] == token {
					delete(c.entries, key)
				}
				c.mu.Unlock()
				return nil, err
			}
			if current {
				token.value, token.expires = value, expires
				c.mu.Unlock()
				return value, nil
			}
			c.mu.Unlock()
		}
		return nil, errors.New("分组资格变更频繁，请重试")
	})
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-result:
		return result.Val, result.Err
	}
}

type customGroupAccessChange struct {
	Kind    string   `json:"kind"`
	ID      int64    `json:"id"`
	Balance *float64 `json:"balance"`
}

func customGroupAccessCacheKey(kind string, id int64) string {
	return kind + ":" + strconv.FormatInt(id, 10)
}

func (c *customGroupAccessCache) changed(payload string) error {
	var change customGroupAccessChange
	if err := json.Unmarshal([]byte(payload), &change); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	switch change.Kind {
	case "reset":
		c.entries = make(map[string]*customGroupAccessCacheEntry)
		c.generation++
	case "policies":
		delete(c.entries, "policies")
	case "monthly":
		key := customGroupAccessCacheKey("monthly", change.ID)
		for existing := range c.entries {
			if existing == key || strings.HasPrefix(existing, key+":") {
				delete(c.entries, existing)
			}
		}
	case "group":
		delete(c.entries, customGroupAccessCacheKey(change.Kind, change.ID))
	case "user", "subscription":
		delete(c.entries, customGroupAccessCacheKey("user", change.ID))
		prefix := customGroupAccessCacheKey("subscription", change.ID) + ":"
		for key := range c.entries {
			if strings.HasPrefix(key, prefix) {
				delete(c.entries, key)
			}
		}
	case "balance":
		if change.Balance == nil {
			return errors.New("余额通知缺少金额")
		}
		key := customGroupAccessCacheKey("user", change.ID)
		entry := c.entries[key]
		if entry == nil {
			return nil
		}
		user, ok := entry.value.(*User)
		if !ok {
			delete(c.entries, key)
			return nil
		}
		copy := *user
		copy.Balance = *change.Balance
		c.entries[key] = &customGroupAccessCacheEntry{value: &copy, expires: entry.expires}
	default:
		return errors.New("未知分组资格通知")
	}
	return nil
}

func (s *CustomGroupAccessService) loadPolicies(ctx context.Context) (map[int64]CustomGroupAccessPolicy, error) {
	now := s.now()
	value, err := s.cache.get(ctx, "policies", now, now.Add(customGroupAccessCacheTTL), func(ctx context.Context) (any, error) { return s.store.List(ctx) })
	if err != nil {
		return nil, err
	}
	policies, ok := value.(map[int64]CustomGroupAccessPolicy)
	if !ok {
		return nil, errors.New("分组规则缓存类型无效")
	}
	return policies, nil
}

func (s *CustomGroupAccessService) loadUser(ctx context.Context, userID int64) (*User, error) {
	now := s.now()
	value, err := s.cache.get(ctx, customGroupAccessCacheKey("user", userID), now, now.Add(customGroupAccessCacheTTL), func(ctx context.Context) (any, error) { return s.users.GetByID(ctx, userID) })
	if err != nil {
		return nil, err
	}
	cached, ok := value.(*User)
	if !ok || cached == nil {
		return nil, errors.New("分组资格用户缓存无效")
	}
	user := *cached
	return &user, nil
}

func (s *CustomGroupAccessService) loadGroup(ctx context.Context, groupID int64) (*Group, error) {
	now := s.now()
	value, err := s.cache.get(ctx, customGroupAccessCacheKey("group", groupID), now, now.Add(customGroupAccessCacheTTL), func(ctx context.Context) (any, error) { return s.groups.GetByIDLite(ctx, groupID) })
	if err != nil {
		return nil, err
	}
	cached, ok := value.(*Group)
	if !ok || cached == nil {
		return nil, errors.New("分组资格分组缓存无效")
	}
	group := *cached
	return &group, nil
}

func (s *CustomGroupAccessService) loadFacts(ctx context.Context, userID int64, monthly bool) (customGroupAccessFacts, error) {
	now := s.now()
	start, end := customGroupAccessMonth(now)
	if s.cache == nil {
		return s.store.Facts(ctx, userID, start, end)
	}
	if !monthly {
		user, err := s.loadUser(ctx, userID)
		if err != nil {
			return customGroupAccessFacts{}, err
		}
		return customGroupAccessFacts{Balance: user.Balance}, nil
	}
	expires := now.Add(customGroupAccessCacheTTL)
	if end.Before(expires) {
		expires = end
	}
	key := customGroupAccessCacheKey("monthly", userID) + ":" + strconv.FormatInt(start.Unix(), 10)
	value, err := s.cache.get(ctx, key, now, expires, func(ctx context.Context) (any, error) {
		facts, err := s.store.Facts(ctx, userID, start, end)
		return facts.MonthlyRecharge, err
	})
	if err != nil {
		return customGroupAccessFacts{}, err
	}
	monthlyRecharge, ok := value.(float64)
	if !ok {
		return customGroupAccessFacts{}, errors.New("分组资格月充值缓存类型无效")
	}
	user, err := s.loadUser(ctx, userID)
	if err != nil {
		return customGroupAccessFacts{}, err
	}
	return customGroupAccessFacts{Balance: user.Balance, MonthlyRecharge: monthlyRecharge}, nil
}

func (s *CustomGroupAccessService) loadSubscription(ctx context.Context, userID, groupID int64) (*UserSubscription, error) {
	now := s.now()
	key := customGroupAccessCacheKey("subscription", userID) + ":" + strconv.FormatInt(groupID, 10)
	value, err := s.cache.get(ctx, key, now, now.Add(customGroupAccessCacheTTL), func(ctx context.Context) (any, error) {
		return s.subscriptions.GetActiveByUserIDAndGroupID(ctx, userID, groupID)
	})
	if err != nil {
		return nil, err
	}
	subscription, ok := value.(*UserSubscription)
	if !ok || subscription == nil {
		return nil, errors.New("分组资格订阅缓存无效")
	}
	if s.cache != nil && !subscription.ExpiresAt.IsZero() && !now.Before(subscription.ExpiresAt) {
		return nil, ErrGroupNotAllowed
	}
	return subscription, nil
}
