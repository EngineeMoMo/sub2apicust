package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"mime"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

// 租约定期续期；崩溃后90秒回收共享并发槽。完成后的租约继续保存供异步结算取证。
const customDedicatedLeaseTTL = 90 * time.Second

func customDedicatedEndpoint(path string, websocket bool, images bool) bool {
	path = strings.TrimPrefix(path, "/v1")
	if path == "/backend-api/codex/responses" {
		path = "/responses"
	}
	switch path {
	case "/responses", "/messages", "/chat/completions", "/messages/count_tokens", "/responses/input_tokens":
		return true
	case "/images/generations", "/images/edits":
		return !websocket && images
	default:
		return false
	}
}

// 共享计数由绑定行锁串行化；失败事务不消费计数，已准入的请求（含上游失败）不退请求次数。
func (s *CustomDedicatedService) AdmitBillingRequest(ctx context.Context, key *APIKey, path string, payload []byte, websocket bool, existingLease bool, contentType ...string) error {
	if !key.IsCustomDedicatedPrepaid() {
		return nil
	}
	g := key.customDedicatedBilling
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ErrDedicatedAccess
	}
	defer func() { _ = tx.Rollback() }()
	binding, err := scanCustomDedicated(tx.QueryRowContext(ctx, "SELECT "+customDedicatedColumns+" FROM custom_dedicated_accounts WHERE id=$1 AND deleted_at IS NULL FOR UPDATE", g.BindingID))
	if err != nil {
		return ErrDedicatedAccess
	}
	subject := customDedicatedSubject{UserID: g.UserID, GroupID: g.GroupID}
	requestBinding := binding
	requestBinding.GroupID = g.GroupID
	// 在取得共享锁后使用同一个数据库时刻，排队跨分钟不能把计数窗口拨回去。
	var acceptedAt time.Time
	if err := tx.QueryRowContext(ctx, "SELECT clock_timestamp()").Scan(&acceptedAt); err != nil {
		return ErrDedicatedAccess
	}
	if binding.AccountID != g.AccountID || !customDedicatedAllowed(requestBinding, subject, nil, acceptedAt) {
		return ErrDedicatedAccess
	}
	valid, err := customDedicatedRequestAccess(ctx, tx, binding, subject)
	if err != nil || !valid {
		return ErrDedicatedAccess
	}
	p, err := loadCustomDedicatedPolicy(ctx, tx, binding.ID)
	if err != nil || !p.valid() {
		return ErrDedicatedAccess
	}
	if !customDedicatedEndpoint(path, websocket, p.AllowImages) {
		return ErrDedicatedEndpoint
	}
	if int64(len(payload)) > p.MaxBodyBytes {
		return ErrDedicatedPolicy
	}
	if len(payload) > 0 {
		mediaType := ""
		if len(contentType) > 0 {
			mediaType, _, _ = mime.ParseMediaType(contentType[0])
		}
		imageUpload := p.AllowImages && !websocket && strings.TrimPrefix(path, "/v1") == "/images/edits" && mediaType == "multipart/form-data"
		if !imageUpload && !gjson.ValidBytes(payload) {
			return ErrDedicatedPolicy
		}
		if !imageUpload && !p.AllowImages && customDedicatedContainsImageTool(gjson.ParseBytes(payload), 0) {
			return ErrDedicatedEndpoint
		}
	}
	leaseID := g.LeaseID
	if !existingLease {
		var active int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM custom_dedicated_request_leases WHERE binding_id=$1 AND finished_at IS NULL AND expires_at>$2`, binding.ID, acceptedAt).Scan(&active); err != nil {
			return ErrDedicatedAccess
		}
		if active >= p.ConcurrencyLimit {
			return ErrDedicatedLimit
		}
		var random [24]byte
		if _, err := rand.Read(random[:]); err != nil {
			return ErrDedicatedAccess
		}
		leaseID = hex.EncodeToString(random[:])
		if _, err := tx.ExecContext(ctx, `INSERT INTO custom_dedicated_request_leases(id,binding_id,account_id,user_id,api_key_id,group_id,expires_at,accepted_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, leaseID, g.BindingID, g.AccountID, g.UserID, g.APIKeyID, g.GroupID, acceptedAt.Add(customDedicatedLeaseTTL), acceptedAt); err != nil {
			return ErrDedicatedAccess
		}
	} else {
		var active bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM custom_dedicated_request_leases WHERE id=$1 AND binding_id=$2 AND api_key_id=$3 AND finished_at IS NULL AND expires_at>$4)`, g.LeaseID, g.BindingID, g.APIKeyID, acceptedAt).Scan(&active); err != nil || !active {
			return ErrDedicatedAccess
		}
	}
	// WS建连只占并发；每个response.create（含首轮）才消耗共享请求次数。
	if !websocket || existingLease {
		minuteStart := acceptedAt.UTC().Truncate(time.Minute)
		dayStart := time.Date(acceptedAt.UTC().Year(), acceptedAt.UTC().Month(), acceptedAt.UTC().Day(), 0, 0, 0, 0, time.UTC)
		if _, err := tx.ExecContext(ctx, `INSERT INTO custom_dedicated_request_counters(binding_id,minute_start,day_start) VALUES($1,$2,$3) ON CONFLICT(binding_id) DO NOTHING`, g.BindingID, minuteStart, dayStart); err != nil {
			return ErrDedicatedAccess
		}
		var minute, day int
		if err := tx.QueryRowContext(ctx, `SELECT CASE WHEN minute_start=$2 THEN minute_requests ELSE 0 END,
 CASE WHEN day_start=$3 THEN day_requests ELSE 0 END FROM custom_dedicated_request_counters WHERE binding_id=$1 FOR UPDATE`, g.BindingID, minuteStart, dayStart).Scan(&minute, &day); err != nil {
			return ErrDedicatedAccess
		}
		if minute >= p.RPMLimit || (p.DailyRequestLimit > 0 && day >= p.DailyRequestLimit) {
			return ErrDedicatedLimit
		}
		if _, err := tx.ExecContext(ctx, `UPDATE custom_dedicated_request_counters SET minute_start=$2,minute_requests=$3,day_start=$4,day_requests=$5 WHERE binding_id=$1`, g.BindingID, minuteStart, minute+1, dayStart, day+1); err != nil {
			return ErrDedicatedAccess
		}
	}
	if err := tx.Commit(); err != nil {
		return ErrDedicatedAccess
	}
	g.Policy = p
	g.LeaseID = leaseID
	if !existingLease {
		g.WebSocket = websocket
	}
	return nil
}

func (s *APIKeyService) AdmitCustomDedicatedHTTP(ctx context.Context, key *APIKey, path string, payload []byte, websocket bool, contentType ...string) error {
	if !key.IsCustomDedicatedPrepaid() {
		return nil
	}
	if s.customDedicated == nil {
		return ErrDedicatedAccess
	}
	return s.customDedicated.AdmitBillingRequest(ctx, key, path, payload, websocket, false, contentType...)
}

func (s *OpenAIGatewayService) AdmitCustomDedicatedWSTurn(ctx context.Context, key *APIKey, payload []byte) error {
	if !key.IsCustomDedicatedPrepaid() {
		return nil
	}
	if s.customDedicated == nil {
		return ErrDedicatedAccess
	}
	if !customDedicatedWSFrameAllowed(payload) || gjson.GetBytes(payload, "type").String() != "response.create" {
		return ErrDedicatedEndpoint
	}
	return s.customDedicated.AdmitBillingRequest(ctx, key, "/responses", payload, true, true)
}

// 包号只允许已计量的response.create和取消；session.update可持久化生图工具或实时音频，不能旁路逐轮准入。
func (k *APIKey) ValidateCustomDedicatedWSFrame(payload []byte) error {
	if !k.IsCustomDedicatedPrepaid() {
		return nil
	}
	if int64(len(payload)) > k.CustomDedicatedMaxBodyBytes() || !customDedicatedWSFrameAllowed(payload) {
		return ErrDedicatedEndpoint
	}
	return nil
}

func customDedicatedWSFrameAllowed(payload []byte) bool {
	if !gjson.ValidBytes(payload) {
		return false
	}
	node := gjson.ParseBytes(payload)
	if !node.IsObject() {
		return false
	}
	types, allowed := 0, false
	node.ForEach(func(key, value gjson.Result) bool {
		if strings.EqualFold(key.String(), "type") {
			types++
			allowed = key.String() == "type" && value.Type == gjson.String && (value.String() == "response.create" || value.String() == "response.cancel")
		}
		return true
	})
	return types == 1 && allowed
}

// HoldCustomDedicatedLease 返回可靠释放函数；续期失败取消请求，避免继续借用失效并发槽。
func (s *APIKeyService) HoldCustomDedicatedLease(ctx context.Context, key *APIKey, cancel context.CancelFunc) func() {
	if !key.IsCustomDedicatedPrepaid() || s.customDedicated == nil || key.customDedicatedBilling.LeaseID == "" {
		return func() {}
	}
	leaseID := key.customDedicatedBilling.LeaseID
	leaseCtx, stop := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-leaseCtx.Done():
				return
			case <-ticker.C:
				checkCtx, checkCancel := context.WithTimeout(leaseCtx, 5*time.Second)
				result, err := s.customDedicated.db.ExecContext(checkCtx, `UPDATE custom_dedicated_request_leases SET expires_at=NOW()+INTERVAL '90 seconds' WHERE id=$1 AND finished_at IS NULL AND expires_at>NOW()`, leaseID)
				checkCancel()
				var affected int64
				if err == nil {
					affected, err = result.RowsAffected()
				}
				if err != nil || affected != 1 {
					slog.Warn("包号并发租约续期失败，取消当前请求", "lease_id", leaseID, "error", err)
					cancel()
					return
				}
			}
		}
	}()
	return func() {
		stop()
		<-done
		releaseCtx, releaseCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer releaseCancel()
		_, err := s.customDedicated.db.ExecContext(releaseCtx, "UPDATE custom_dedicated_request_leases SET finished_at=COALESCE(finished_at,NOW()) WHERE id=$1", leaseID)
		if err != nil {
			slog.Warn("包号并发租约释放失败，等待到期回收", "lease_id", leaseID, "error", err)
		}
	}
}

func customDedicatedContainsImageTool(node gjson.Result, depth int) bool {
	if depth > 64 {
		return true
	}
	blocked := false
	// ForEach保留重复键；不能被encoding/json末值覆盖而掩盖另一解析器读到的生图工具。
	node.ForEach(func(key, child gjson.Result) bool {
		if strings.EqualFold(key.String(), "type") && (child.String() == "image_generation" || child.String() == "image_generation_call") {
			blocked = true
			return false
		}
		if child.IsObject() || child.IsArray() {
			blocked = customDedicatedContainsImageTool(child, depth+1)
		}
		return !blocked
	})
	return blocked
}
