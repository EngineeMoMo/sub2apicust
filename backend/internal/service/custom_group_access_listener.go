package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/lib/pq"
)

const customGroupAccessChannel = "custom_group_access_changed"

func ProvideCustomGroupAccessService(db *sql.DB, groups GroupRepository, users UserRepository, subscriptions UserSubscriptionRepository, apiKeys *APIKeyService, gateway *GatewayService, cfg *config.Config) *CustomGroupAccessService {
	result := NewCustomGroupAccessService(db, groups, users, subscriptions, apiKeys)
	gateway.customGroupAccess = result
	result.cache = newCustomGroupAccessCache()
	result.expiryWake = make(chan struct{}, 1)
	stopListener := result.startChangeListener(cfg.Database.DSNWithTimezone(cfg.Timezone))
	stopExpiry := func() {}
	if cfg.RunMode != config.RunModeSimple {
		stopExpiry = result.startGroupExpiry(apiKeys)
	}
	result.stop = func() { stopExpiry(); stopListener() }
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
	ready := make(chan struct{})
	dialCtx, cancel := context.WithCancel(context.Background())
	dialer := &customGroupAccessDialer{ctx: dialCtx}
	listener := pq.NewDialListener(dialer, dsn+" connect_timeout=3", time.Second, 10*time.Second, func(event pq.ListenerEventType, err error) {
		s.cache.health(false)
		if err != nil {
			slog.Warn("分组资格通知连接变更，暂时回源校验", "event", event, "error", err)
		}
	})
	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		if err := listener.Listen(customGroupAccessChannel); err != nil {
			slog.Error("分组资格通知订阅失败，使用数据库校验", "error", err)
			return
		}
		s.cache.health(true)
		close(ready)
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
					s.wakeGroupExpiry()
					continue
				}
				var change customGroupAccessChange
				err := json.Unmarshal([]byte(notification.Extra), &change)
				if err == nil {
					err = s.cache.applyChange(change)
				}
				if err != nil {
					s.cache.health(false)
					slog.Error("分组资格通知无效，清空缓存", "error", err)
				}
				if change.Kind == "policies" || change.Kind == "reset" {
					s.wakeGroupExpiry()
				}
			}
		}
	}()
	go func() {
		defer workers.Done()
		select {
		case <-ready:
		case <-stop:
			return
		}
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				dialer.deadline(time.Now().Add(2 * time.Second))
				err := listener.Ping()
				dialer.deadline(time.Time{})
				s.cache.health(err == nil)
			}
		}
	}()
	go func() {
		workers.Wait()
		close(done)
	}()
	var once sync.Once
	return func() {
		once.Do(func() {
			close(stop)
			cancel()
			dialer.close()
			_ = listener.Close()
			<-done
			s.cache.health(false)
		})
	}
}

type customGroupAccessDialer struct {
	ctx     context.Context
	mu      sync.Mutex
	conn    net.Conn
	stopped bool
}

func (d *customGroupAccessDialer) Dial(network, address string) (net.Conn, error) {
	return d.DialTimeout(network, address, 3*time.Second)
}

func (d *customGroupAccessDialer) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	connection, err := (&net.Dialer{Timeout: timeout}).DialContext(d.ctx, network, address)
	if err != nil {
		return nil, err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stopped {
		_ = connection.Close()
		return nil, net.ErrClosed
	}
	d.conn = connection
	return connection, nil
}

func (d *customGroupAccessDialer) deadline(until time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.conn != nil {
		_ = d.conn.SetDeadline(until)
	}
}

func (d *customGroupAccessDialer) close() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stopped = true
	if d.conn != nil {
		_ = d.conn.Close()
	}
}
