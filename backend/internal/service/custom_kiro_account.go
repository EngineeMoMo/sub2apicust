package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/kiro"
	"github.com/gin-gonic/gin"
)

func normalizeCustomKiroCredentials(accountType string, credentials map[string]any) (map[string]any, error) {
	normalized, err := kiro.Normalize(accountType, credentials)
	if err != nil {
		return nil, infraerrors.BadRequest("INVALID_KIRO_CREDENTIALS", err.Error())
	}
	return normalized, nil
}

func mergeCustomKiroCredentials(existing, incoming map[string]any) (map[string]any, error) {
	canonical, err := kiro.CanonicalFields(incoming)
	if err != nil {
		return nil, infraerrors.BadRequest("INVALID_KIRO_CREDENTIALS", err.Error())
	}
	merged := MergeCredentials(existing, shallowCopyMap(canonical))
	if refresh, changed := canonical["refresh_token"].(string); changed && refresh != existing["refresh_token"] {
		for _, key := range []string{"access_token", "expires_at", "profile_arn"} {
			if _, provided := canonical[key]; !provided {
				delete(merged, key)
			}
		}
	}
	return merged, nil
}

func (s *AccountTestService) RefreshCustomKiroAccount(ctx context.Context, account *Account) (*Account, error) {
	if s.customKiro == nil || account == nil || !s.customKiro.CanRefresh(account) {
		return nil, infraerrors.BadRequest("KIRO_NOT_REFRESHABLE", "Kiro 账号缺少可刷新的 OAuth 凭据")
	}
	rejected := account.GetCredential("access_token")
	executor := customKiroRejectedTokenRefresher{CustomKiroService: s.customKiro, rejected: rejected}
	if s.customKiro.refreshAPI == nil {
		return nil, errors.New("kiro 统一刷新服务未配置")
	}
	for attempt := 0; attempt < 40; attempt++ {
		result, err := s.customKiro.refreshAPI.RefreshIfNeeded(ctx, account, executor, time.Minute)
		if err != nil {
			return nil, err
		}
		if result != nil && !result.LockHeld && result.Account != nil {
			return result.Account, nil
		}
		timer := time.NewTimer(250 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	return nil, errors.New("kiro 刷新等待超时")
}

func (s *AccountTestService) testCustomKiroConnection(c *gin.Context, account *Account, model, prompt string) error {
	if s.customKiro == nil {
		return s.sendErrorAndEnd(c, "Kiro 服务未配置")
	}
	if model == "" {
		model = "claude-sonnet-4-6"
	}
	if prompt == "" {
		prompt = "Reply with OK."
	}
	body, err := json.Marshal(map[string]any{"model": model, "messages": []map[string]string{{"role": "user", "content": prompt}}, "max_tokens": 64})
	if err != nil {
		return s.sendErrorAndEnd(c, err.Error())
	}
	ctx, cancel := context.WithTimeout(context.WithValue(c.Request.Context(), customKiroProbeContextKey{}, true), 2*time.Minute)
	defer cancel()
	call, err := s.customKiro.call(ctx, account, body, model)
	if err != nil {
		return s.sendErrorAndEnd(c, err.Error())
	}
	defer func() { _ = call.Response.Body.Close() }()
	if call.Response.StatusCode != http.StatusOK {
		return s.sendErrorAndEnd(c, fmt.Sprintf("Kiro 上游连接测试失败（HTTP %d）", call.Response.StatusCode))
	}
	if !strings.HasPrefix(strings.ToLower(call.Response.Header.Get("Content-Type")), "application/vnd.amazon.eventstream") {
		return s.sendErrorAndEnd(c, "Kiro 上游未返回 EventStream")
	}
	s.sendEvent(c, TestEvent{Type: "test_start", Model: model})
	decoded, err := kiro.DecodeResponse(ctx, call.Response.Body, func(block kiro.ResponseBlock) error {
		if block.Type == "text" {
			s.sendEvent(c, TestEvent{Type: "content", Text: block.Text})
		}
		return nil
	})
	if err != nil {
		return s.sendErrorAndEnd(c, err.Error())
	}
	if decoded == nil {
		return s.sendErrorAndEnd(c, "Kiro 响应内容为空")
	}
	if !decoded.HasUsage && call.Account.GetCredential("usage_mode") == "estimated" {
		if err := resolveCustomKiroUsage(ctx, call.Account, body, model, decoded); err != nil {
			return s.sendErrorAndEnd(c, err.Error())
		}
	}
	if !decoded.HasUsage {
		s.sendEvent(c, TestEvent{Type: "status", Status: "warning", Text: "连接与生成成功；上游未返回可识别的精确 Token 用量。当前精确计量模式会拒绝正式请求；如接受估算计费，请在账号计量方式中显式启用。"})
	} else if decoded.UsageSource == "estimated" {
		s.sendEvent(c, TestEvent{Type: "status", Status: "warning", Text: "连接与生成成功；本次用量为本地估算。正式请求会按本站模型价格和倍率对估算 Token 计费，不能当作 Kiro 精确用量或账单。"})
	}
	source := decoded.UsageSource
	if source == "" {
		source = "missing"
	}
	s.sendEvent(c, TestEvent{Type: "test_complete", Success: true, Data: map[string]any{"usage_source": source, "billing_ready": decoded.HasUsage, "usage": decoded.Usage, "event_counts": decoded.EventCounts}})
	return nil
}
