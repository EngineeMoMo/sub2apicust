package service

import (
	"database/sql"
	"log/slog"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/lib/pq"
)

const customGroupAccessChannel = "custom_group_access_changed"

func ProvideCustomGroupAccessService(db *sql.DB, groups GroupRepository, users UserRepository, subscriptions UserSubscriptionRepository, apiKeys *APIKeyService, cfg *config.Config) *CustomGroupAccessService {
	result := NewCustomGroupAccessService(db, groups, users, subscriptions, apiKeys)
	result.cache = newCustomGroupAccessCache()
	result.stop = result.startChangeListener(cfg.Database.DSNWithTimezone(cfg.Timezone))
	return result
}

func (s *CustomGroupAccessService) Stop() {
	if s != nil && s.stop != nil {
		s.stop()
	}
}

func (s *CustomGroupAccessService) startChangeListener(dsn string) func() {
	stop := make(chan struct{})
	done := make(chan struct{})
	listener := pq.NewListener(dsn+" connect_timeout=3", time.Second, 10*time.Second, func(event pq.ListenerEventType, err error) {
		s.cache.health(false)
		if err != nil {
			slog.Warn("分组资格通知连接变更，暂时回源校验", "event", event, "error", err)
		}
	})
	go func() {
		defer close(done)
		if err := listener.Listen(customGroupAccessChannel); err != nil {
			slog.Error("分组资格通知订阅失败，使用数据库校验", "error", err)
			return
		}
		s.cache.health(true)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case notification, ok := <-listener.Notify:
				if !ok {
					s.cache.health(false)
					return
				}
				if notification == nil {
					s.cache.health(false)
					s.cache.health(true)
					continue
				}
				if err := s.cache.changed(notification.Extra); err != nil {
					s.cache.health(false)
					slog.Error("分组资格通知无效，清空缓存", "error", err)
				}
			case <-ticker.C:
				if err := listener.Ping(); err != nil {
					s.cache.health(false)
				} else {
					s.cache.health(true)
				}
			}
		}
	}()
	var once sync.Once
	return func() {
		once.Do(func() {
			s.cache.health(false)
			close(stop)
			_ = listener.Close()
			<-done
		})
	}
}
