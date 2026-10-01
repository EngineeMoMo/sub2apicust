package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const BillingTypeDedicated int8 = 2

var (
	ErrDedicatedLimit    = infraerrors.TooManyRequests("DEDICATED_USAGE_LIMIT", "包号共享请求频率、并发或每日请求上限已达到，请稍后重试")
	ErrDedicatedPolicy   = infraerrors.BadRequest("DEDICATED_BILLING_POLICY", "请检查包号共享使用限制")
	ErrDedicatedEndpoint = infraerrors.Forbidden("DEDICATED_ENDPOINT_NOT_ALLOWED", "此包号不支持该入口或尚未开放生图权益，请联系管理员")
)

// CustomDedicatedBillingPolicy 只限制包号使用；参考金额不充当真实上游账单。
type CustomDedicatedBillingPolicy struct {
	ConcurrencyLimit  int        `json:"concurrency_limit"`
	RPMLimit          int        `json:"rpm_limit"`
	DailyRequestLimit int        `json:"daily_request_limit"`
	MaxBodyBytes      int64      `json:"max_body_bytes"`
	AllowImages       bool       `json:"allow_images"`
	UpdatedAt         *time.Time `json:"updated_at"`
}

func defaultCustomDedicatedPolicy() CustomDedicatedBillingPolicy {
	return CustomDedicatedBillingPolicy{ConcurrencyLimit: 2, RPMLimit: 30, MaxBodyBytes: 2 << 20}
}

func (p CustomDedicatedBillingPolicy) valid() bool {
	return p.ConcurrencyLimit >= 1 && p.ConcurrencyLimit <= 200 && p.RPMLimit >= 1 && p.RPMLimit <= 10000 && p.DailyRequestLimit >= 0 && p.DailyRequestLimit <= 1000000 && p.MaxBodyBytes >= 1024 && p.MaxBodyBytes <= 32<<20
}

// 凭证只附着在单次认证返回的Key副本，不能写入共享认证缓存或由客户端构造。
type customDedicatedBillingGrant struct {
	BindingID, AccountID, UserID, APIKeyID, GroupID int64
	WebSocket                                       bool
	LeaseID                                         string
	Policy                                          CustomDedicatedBillingPolicy
}

func (k *APIKey) IsCustomDedicatedPrepaid() bool {
	if k == nil || k.customDedicatedBilling == nil || k.GroupID == nil {
		return false
	}
	g := k.customDedicatedBilling
	return g.BindingID > 0 && g.AccountID > 0 && g.UserID == k.UserID && g.APIKeyID == k.ID && g.GroupID == *k.GroupID
}

// PrepareCustomDedicatedBilling 正向解析包号。无绑定保持普通计费，不能把Check返回nil当免费资格。
func (s *APIKeyService) PrepareCustomDedicatedBilling(ctx context.Context, key *APIKey) (*APIKey, error) {
	if s.customDedicated == nil || key == nil || key.GroupID == nil {
		return key, nil
	}
	rows, err := s.customDedicated.db.QueryContext(ctx, "SELECT "+customDedicatedColumns+` FROM custom_dedicated_accounts WHERE deleted_at IS NULL
 AND (group_id=$1 OR account_id IN (SELECT account_id FROM account_groups WHERE group_id=$1))`, *key.GroupID)
	if err != nil {
		return nil, ErrDedicatedAccess
	}
	var bindings []CustomDedicatedBinding
	for rows.Next() {
		binding, scanErr := scanCustomDedicated(rows)
		if scanErr != nil {
			_ = rows.Close()
			return nil, ErrDedicatedAccess
		}
		bindings = append(bindings, binding)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, ErrDedicatedAccess
	}
	if err := rows.Close(); err != nil {
		return nil, ErrDedicatedAccess
	}
	if len(bindings) == 0 {
		return key, nil
	}
	if len(bindings) != 1 || s.customDedicated.simpleMode {
		return nil, ErrDedicatedAccess
	}
	binding := bindings[0]
	subject := customDedicatedSubject{UserID: key.UserID, GroupID: *key.GroupID}
	requestBinding := binding
	requestBinding.GroupID = subject.GroupID
	if !customDedicatedAllowed(requestBinding, subject, nil, time.Now()) {
		return nil, ErrDedicatedAccess
	}
	valid, err := customDedicatedRequestAccess(ctx, s.customDedicated.db, binding, subject)
	if err != nil || !valid {
		return nil, ErrDedicatedAccess
	}
	copyKey := *key
	policy, err := loadCustomDedicatedPolicy(ctx, s.customDedicated.db, binding.ID)
	if err != nil || !policy.valid() {
		return nil, ErrDedicatedAccess
	}
	copyKey.customDedicatedBilling = &customDedicatedBillingGrant{BindingID: binding.ID, AccountID: binding.AccountID, UserID: key.UserID, APIKeyID: key.ID, GroupID: *key.GroupID}
	copyKey.customDedicatedBilling.Policy = policy
	return &copyKey, nil
}

func (k *APIKey) CustomDedicatedMaxBodyBytes() int64 {
	if !k.IsCustomDedicatedPrepaid() {
		return 0
	}
	return k.customDedicatedBilling.Policy.MaxBodyBytes
}

const customDedicatedPolicySQL = `SELECT concurrency_limit,rpm_limit,daily_request_limit,max_body_bytes,allow_images,updated_at FROM custom_dedicated_billing_policies WHERE binding_id=$1`

func loadCustomDedicatedPolicy(ctx context.Context, query customDedicatedQuery, id int64) (CustomDedicatedBillingPolicy, error) {
	p := defaultCustomDedicatedPolicy()
	err := query.QueryRowContext(ctx, customDedicatedPolicySQL, id).Scan(&p.ConcurrencyLimit, &p.RPMLimit, &p.DailyRequestLimit, &p.MaxBodyBytes, &p.AllowImages, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return p, nil
	}
	return p, err
}

func (s *CustomDedicatedService) BillingPolicy(ctx context.Context, id int64) (CustomDedicatedBillingPolicy, error) {
	if s.simpleMode {
		return CustomDedicatedBillingPolicy{}, ErrDedicatedMode
	}
	var exists bool
	if err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM custom_dedicated_accounts WHERE id=$1 AND deleted_at IS NULL)", id).Scan(&exists); err != nil {
		return CustomDedicatedBillingPolicy{}, err
	}
	if !exists {
		return CustomDedicatedBillingPolicy{}, ErrDedicatedNotFound
	}
	return loadCustomDedicatedPolicy(ctx, s.db, id)
}

func (s *CustomDedicatedService) UpdateBillingPolicy(ctx context.Context, id int64, policy CustomDedicatedBillingPolicy) (CustomDedicatedBillingPolicy, error) {
	if s.simpleMode {
		return CustomDedicatedBillingPolicy{}, ErrDedicatedMode
	}
	if !policy.valid() {
		return CustomDedicatedBillingPolicy{}, ErrDedicatedPolicy
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return CustomDedicatedBillingPolicy{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var found int64
	if err := tx.QueryRowContext(ctx, "SELECT id FROM custom_dedicated_accounts WHERE id=$1 AND deleted_at IS NULL FOR UPDATE", id).Scan(&found); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CustomDedicatedBillingPolicy{}, ErrDedicatedNotFound
		}
		return CustomDedicatedBillingPolicy{}, err
	}
	current, err := loadCustomDedicatedPolicy(ctx, tx, id)
	if err != nil {
		return CustomDedicatedBillingPolicy{}, err
	}
	if (current.UpdatedAt == nil) != (policy.UpdatedAt == nil) || (current.UpdatedAt != nil && !current.UpdatedAt.Equal(*policy.UpdatedAt)) {
		return CustomDedicatedBillingPolicy{}, ErrDedicatedStale
	}
	if err := tx.QueryRowContext(ctx, `INSERT INTO custom_dedicated_billing_policies(binding_id,concurrency_limit,rpm_limit,daily_request_limit,max_body_bytes,allow_images)
 VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(binding_id) DO UPDATE SET concurrency_limit=$2,rpm_limit=$3,daily_request_limit=$4,max_body_bytes=$5,allow_images=$6,updated_at=clock_timestamp() RETURNING updated_at`, id, policy.ConcurrencyLimit, policy.RPMLimit, policy.DailyRequestLimit, policy.MaxBodyBytes, policy.AllowImages).Scan(&policy.UpdatedAt); err != nil {
		return CustomDedicatedBillingPolicy{}, err
	}
	return policy, tx.Commit()
}

// 包号参考用量固定为原始模型计量，不能被用户或分组倍率0绕过Key限额。
func (p *postUsageBillingParams) customDedicatedMeterCost() float64 {
	if p == nil || p.Cost == nil {
		return 0
	}
	if p.APIKey.IsCustomDedicatedPrepaid() {
		return p.Cost.TotalCost
	}
	return p.Cost.ActualCost
}

func prepareCustomDedicatedSettlement(p *postUsageBillingParams, log *UsageLog) error {
	if p == nil || p.APIKey == nil || p.APIKey.customDedicatedBilling == nil {
		return nil
	}
	if !p.APIKey.IsCustomDedicatedPrepaid() {
		return ErrDedicatedAccess
	}
	g := p.APIKey.customDedicatedBilling
	if p.User == nil || p.Account == nil || p.User.ID != g.UserID || p.Account.ID != g.AccountID || strings.TrimSpace(g.LeaseID) == "" || p.IsSubscriptionBill || p.SimpleModeKeyRateLimitOnly || p.Cost == nil || math.IsNaN(p.Cost.TotalCost) || math.IsInf(p.Cost.TotalCost, 0) || p.Cost.TotalCost < 0 {
		return ErrDedicatedAccess
	}
	if log != nil {
		log.BillingType = BillingTypeDedicated
		log.ActualCost = 0
		log.RateMultiplier = 0
		log.SubscriptionID = nil
	}
	return nil
}

// 每次真实执行使用服务端准入身份；客户端重复X-Request-ID不能抹掉真实包号计量。
func (k *APIKey) CustomDedicatedUsageRequestID(upstreamID string) string {
	if !k.IsCustomDedicatedPrepaid() {
		return ""
	}
	g := k.customDedicatedBilling
	if g.LeaseID == "" {
		return ""
	}
	if !g.WebSocket {
		return "dedicated:" + g.LeaseID
	}
	if strings.TrimSpace(upstreamID) == "" {
		return ""
	}
	hash := sha256.Sum256([]byte(upstreamID))
	return "dedicated:" + g.LeaseID + ":" + fmt.Sprintf("%x", hash[:])
}
