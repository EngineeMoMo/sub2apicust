package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math"
	"strings"
	"time"
)

const BillingTypeDedicated int8 = 2

// 凭证只附着在单次认证返回的Key副本，不能写入共享认证缓存或由客户端构造。
type customDedicatedBillingGrant struct {
	BindingID, AccountID, UserID, APIKeyID, GroupID int64
	WebSocket                                       bool
	LeaseID                                         string
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
	copyKey.customDedicatedBilling = &customDedicatedBillingGrant{BindingID: binding.ID, AccountID: binding.AccountID, UserID: key.UserID, APIKeyID: key.ID, GroupID: *key.GroupID}
	return &copyKey, nil
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
