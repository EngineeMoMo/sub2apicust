package service

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrAuthIPBlocked            = infraerrors.Forbidden("IP_BLOCKED", "当前 IP 的敏感认证请求已被封禁，请联系管理员")
	ErrAuthIPGuardUnavailable   = infraerrors.ServiceUnavailable("AUTH_IP_GUARD_UNAVAILABLE", "认证防护暂时不可用，请稍后重试")
	ErrAuthIPCounterUnavailable = infraerrors.TooManyRequests("AUTH_IP_COUNTER_UNAVAILABLE", "认证请求暂时受限，请稍后重试")
	ErrAuthIPBlacklistConflict  = infraerrors.Conflict("IP_BLACKLIST_CONFLICT", "状态已变化，请刷新后重试")
	ErrAuthIPBlacklistNotFound  = infraerrors.NotFound("IP_BLACKLIST_NOT_FOUND", "黑名单记录不存在")
	ErrAuthIPAlreadyBlocked     = infraerrors.Conflict("IP_BLACKLIST_ALREADY_BLOCKED", "该 IP 已被封禁，请刷新名单；如需变更期限，请先解除原封禁")
)

const (
	AuthIPBlockModeTimed     = "timed"
	AuthIPBlockModePermanent = "permanent"
)

type AuthIPBlacklistConfig struct {
	Enabled       bool   `json:"enabled"`
	Threshold     int    `json:"threshold"`
	WindowSeconds int    `json:"window_seconds"`
	Version       int64  `json:"version"`
	BlockSeconds  int    `json:"block_seconds"`
	BlockMode     string `json:"block_mode"`
}

type AuthIPBlacklistEntry struct {
	ID            int64      `json:"id"`
	IP            string     `json:"ip"`
	Blocked       bool       `json:"blocked"`
	Generation    int64      `json:"generation"`
	BlockedAt     time.Time  `json:"blocked_at"`
	TriggerPath   string     `json:"trigger_path"`
	TriggerCount  int        `json:"trigger_count"`
	WindowSeconds int        `json:"window_seconds"`
	FirstEventAt  *time.Time `json:"first_event_at"`
	LastEventAt   *time.Time `json:"last_event_at"`
	RuleVersion   int64      `json:"rule_version"`
	BlockedUntil  *time.Time `json:"blocked_until"`
	Source        string     `json:"source"`
	Reason        string     `json:"reason"`
}

type AuthIPBlacklistSnapshot struct {
	Config AuthIPBlacklistConfig
	Entry  *AuthIPBlacklistEntry
}

type AuthIPWindowCount struct {
	Count int
	Now   time.Time
	First time.Time
	Last  time.Time
}

type AuthIPBlacklistDecision struct {
	Entry        *AuthIPBlacklistEntry
	NewlyBlocked bool
}

type AuthIPBlacklistRepository interface {
	Snapshot(context.Context, string) (*AuthIPBlacklistSnapshot, error)
	Config(context.Context) (*AuthIPBlacklistConfig, error)
	UpdateConfig(context.Context, AuthIPBlacklistConfig) (*AuthIPBlacklistConfig, error)
	List(context.Context, string, int, int) ([]AuthIPBlacklistEntry, int64, error)
	Block(context.Context, string, string, AuthIPBlacklistSnapshot, AuthIPWindowCount) (*AuthIPBlacklistDecision, error)
	Unblock(context.Context, int64, int64) (*AuthIPBlacklistEntry, error)
	ManualBlock(context.Context, AuthIPManualBlock) (*AuthIPBlacklistEntry, error)
}

type AuthIPWindowCounter interface {
	Count(context.Context, string, int64, AuthIPBlacklistConfig) (*AuthIPWindowCount, error)
}

type AuthIPRequestCounter interface {
	Time(context.Context) (time.Time, error)
	CountAt(context.Context, string, int64, AuthIPBlacklistConfig, time.Time) (*AuthIPWindowCount, error)
}

type AuthIPObservation struct {
	IP       string
	Path     string
	Snapshot AuthIPBlacklistSnapshot
	Entered  time.Time
}

func (config AuthIPBlacklistConfig) CooldownSeconds() int {
	if config.BlockSeconds == 0 {
		return 900
	}
	return config.BlockSeconds
}

type AuthIPBlacklistService struct {
	repo    AuthIPBlacklistRepository
	counter AuthIPWindowCounter
}

func NewAuthIPBlacklistService(repo AuthIPBlacklistRepository, counter AuthIPWindowCounter) *AuthIPBlacklistService {
	return &AuthIPBlacklistService{repo: repo, counter: counter}
}

func NormalizeAuthIP(raw string) (string, error) {
	address, err := netip.ParseAddr(raw)
	if err != nil || address.Zone() != "" {
		return "", infraerrors.BadRequest("IP_BLACKLIST_INVALID_IP", "请输入完整有效的 IPv4 或 IPv6 地址")
	}
	return address.Unmap().String(), nil
}

func ValidateAuthIPBlacklistConfig(config AuthIPBlacklistConfig) error {
	if config.BlockMode != AuthIPBlockModeTimed && config.BlockMode != AuthIPBlockModePermanent {
		return infraerrors.BadRequest("IP_BLACKLIST_INVALID_CONFIG", "请选择临时封禁或永久封禁，并提交封禁方式")
	}
	if seconds := config.CooldownSeconds(); seconds < 60 || seconds > 86400 || seconds%60 != 0 {
		return infraerrors.BadRequest("IP_BLACKLIST_INVALID_CONFIG", "封禁时长须为1–1440分钟的整数")
	}
	if config.Threshold < 1 || config.Threshold > 1000 || config.WindowSeconds < 60 || config.WindowSeconds > 86400 || config.WindowSeconds%60 != 0 || config.Version < 1 {
		return infraerrors.BadRequest("IP_BLACKLIST_INVALID_CONFIG", "次数须为1–1000的整数，周期须为1–1440分钟的整数，并提交当前配置版本")
	}
	return nil
}

func (guard *AuthIPBlacklistService) Check(ctx context.Context, rawIP, path string, countRequest bool) (*AuthIPBlacklistDecision, error) {
	if guard == nil || guard.repo == nil {
		return nil, ErrAuthIPGuardUnavailable
	}
	canonicalIP, err := NormalizeAuthIP(rawIP)
	if err != nil {
		return nil, ErrAuthIPGuardUnavailable.WithCause(err)
	}
	snapshot, err := guard.repo.Snapshot(ctx, canonicalIP)
	if err != nil || snapshot == nil {
		return nil, ErrAuthIPGuardUnavailable.WithCause(err)
	}
	if snapshot.Entry != nil && snapshot.Entry.Blocked {
		return &AuthIPBlacklistDecision{Entry: snapshot.Entry}, ErrAuthIPBlocked
	}
	if !snapshot.Config.Enabled || !countRequest {
		return &AuthIPBlacklistDecision{}, nil
	}
	if err = ValidateAuthIPBlacklistConfig(snapshot.Config); err != nil {
		return nil, ErrAuthIPGuardUnavailable.WithCause(err)
	}
	if guard.counter == nil {
		return nil, ErrAuthIPCounterUnavailable
	}
	generation := int64(0)
	if snapshot.Entry != nil {
		generation = snapshot.Entry.Generation
	}
	events, err := guard.counter.Count(ctx, canonicalIP, generation, snapshot.Config)
	if err != nil || events == nil {
		return nil, ErrAuthIPCounterUnavailable.WithCause(err)
	}
	if events.Count < snapshot.Config.Threshold {
		return &AuthIPBlacklistDecision{}, nil
	}
	decision, err := guard.repo.Block(ctx, canonicalIP, path, *snapshot, *events)
	if err != nil || decision == nil {
		return nil, ErrAuthIPGuardUnavailable.WithCause(err)
	}
	if decision.Entry != nil && decision.Entry.Blocked {
		return decision, ErrAuthIPBlocked
	}
	return decision, nil
}

func (guard *AuthIPBlacklistService) Config(ctx context.Context) (*AuthIPBlacklistConfig, error) {
	if guard == nil || guard.repo == nil {
		return nil, ErrAuthIPGuardUnavailable
	}
	result, err := guard.repo.Config(ctx)
	if err != nil {
		return nil, ErrAuthIPGuardUnavailable.WithCause(err)
	}
	return result, nil
}

func (guard *AuthIPBlacklistService) UpdateConfig(ctx context.Context, config AuthIPBlacklistConfig) (*AuthIPBlacklistConfig, error) {
	config.BlockSeconds = config.CooldownSeconds()
	if err := ValidateAuthIPBlacklistConfig(config); err != nil {
		return nil, err
	}
	if guard == nil || guard.repo == nil {
		return nil, ErrAuthIPGuardUnavailable
	}
	result, err := guard.repo.UpdateConfig(ctx, config)
	return result, authIPStorageError(err)
}

func (guard *AuthIPBlacklistService) List(ctx context.Context, rawIP string, page, pageSize int) ([]AuthIPBlacklistEntry, int64, error) {
	if page < 1 || page > 1000000 || pageSize < 1 || pageSize > 100 {
		return nil, 0, infraerrors.BadRequest("IP_BLACKLIST_INVALID_PAGE", "分页参数无效")
	}
	canonicalIP := ""
	if rawIP != "" {
		var err error
		canonicalIP, err = NormalizeAuthIP(rawIP)
		if err != nil {
			return nil, 0, err
		}
	}
	if guard == nil || guard.repo == nil {
		return nil, 0, ErrAuthIPGuardUnavailable
	}
	items, total, err := guard.repo.List(ctx, canonicalIP, page, pageSize)
	return items, total, authIPStorageError(err)
}

func (guard *AuthIPBlacklistService) Unblock(ctx context.Context, id, generation int64) (*AuthIPBlacklistEntry, error) {
	if id < 1 || generation < 1 {
		return nil, infraerrors.BadRequest("IP_BLACKLIST_INVALID_ENTRY", "记录编号和状态版本无效")
	}
	if guard == nil || guard.repo == nil {
		return nil, ErrAuthIPGuardUnavailable
	}
	result, err := guard.repo.Unblock(ctx, id, generation)
	return result, authIPStorageError(err)
}

func authIPStorageError(err error) error {
	if err == nil || errors.Is(err, ErrAuthIPBlacklistConflict) || errors.Is(err, ErrAuthIPBlacklistNotFound) || errors.Is(err, ErrAuthIPAlreadyBlocked) {
		return err
	}
	return ErrAuthIPGuardUnavailable.WithCause(err)
}

func AuthIPCounterKey(ip string, generation, ruleVersion int64) string {
	return fmt.Sprintf("custom_auth_ip_window:%s:%d:%d", ip, generation, ruleVersion)
}

func (guard *AuthIPBlacklistService) Begin(ctx context.Context, rawIP, path string, observe bool) (*AuthIPObservation, *AuthIPBlacklistDecision, error) {
	if guard == nil || guard.repo == nil {
		return nil, nil, ErrAuthIPGuardUnavailable
	}
	ip, err := NormalizeAuthIP(rawIP)
	if err != nil {
		return nil, nil, ErrAuthIPGuardUnavailable.WithCause(err)
	}
	snapshot, err := guard.repo.Snapshot(ctx, ip)
	if err != nil || snapshot == nil {
		return nil, nil, ErrAuthIPGuardUnavailable.WithCause(err)
	}
	if snapshot.Entry != nil && snapshot.Entry.Blocked {
		return nil, &AuthIPBlacklistDecision{Entry: snapshot.Entry}, ErrAuthIPBlocked
	}
	if !observe || !snapshot.Config.Enabled {
		return nil, nil, nil
	}
	if err := ValidateAuthIPBlacklistConfig(snapshot.Config); err != nil {
		return nil, nil, ErrAuthIPGuardUnavailable.WithCause(err)
	}
	counter, ok := guard.counter.(AuthIPRequestCounter)
	if !ok {
		return nil, nil, ErrAuthIPCounterUnavailable
	}
	entered, err := counter.Time(ctx)
	if err != nil {
		return nil, nil, ErrAuthIPCounterUnavailable.WithCause(err)
	}
	return &AuthIPObservation{IP: ip, Path: path, Snapshot: *snapshot, Entered: entered}, nil, nil
}

func (guard *AuthIPBlacklistService) Failed(ctx context.Context, observation *AuthIPObservation) (*AuthIPBlacklistDecision, error) {
	if observation == nil {
		return nil, nil
	}
	counter, ok := guard.counter.(AuthIPRequestCounter)
	if !ok {
		return nil, ErrAuthIPCounterUnavailable
	}
	generation := int64(0)
	if observation.Snapshot.Entry != nil {
		generation = observation.Snapshot.Entry.Generation
	}
	events, err := counter.CountAt(ctx, observation.IP, generation, observation.Snapshot.Config, observation.Entered)
	if err != nil || events == nil {
		return nil, ErrAuthIPCounterUnavailable.WithCause(err)
	}
	if events.Count < observation.Snapshot.Config.Threshold {
		return nil, nil
	}
	decision, err := guard.repo.Block(ctx, observation.IP, observation.Path, observation.Snapshot, *events)
	if err != nil || decision == nil {
		return nil, ErrAuthIPGuardUnavailable.WithCause(err)
	}
	if decision.Entry != nil && decision.Entry.Blocked {
		return decision, ErrAuthIPBlocked
	}
	return decision, nil
}
