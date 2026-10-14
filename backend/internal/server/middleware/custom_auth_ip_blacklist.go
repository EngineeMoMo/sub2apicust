package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func AuthIPBlacklistRoute(method, path string) (bool, bool) {
	const prefix = "/api/v1/auth/"
	if !strings.HasPrefix(path, prefix) {
		return false, false
	}
	relative := strings.TrimPrefix(path, prefix)
	switch relative {
	case "refresh", "logout", "validate-promo-code", "validate-invitation-code":
		return false, false
	}
	if strings.HasPrefix(relative, "oauth/wechat/payment/") {
		return false, false
	}
	if method == http.MethodPost {
		return true, true
	}
	if method == http.MethodGet && strings.HasPrefix(relative, "oauth/") &&
		(strings.HasSuffix(relative, "/start") || strings.HasSuffix(relative, "/callback")) {
		return true, false
	}
	return false, false
}

func NewAuthIPBlacklistGuard(guard *service.AuthIPBlacklistService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		protected, count := AuthIPBlacklistRoute(ctx.Request.Method, ctx.FullPath())
		if !protected {
			ctx.Next()
			return
		}
		observation, decision, err := guard.Begin(ctx.Request.Context(), SecurityClientIP(ctx), ctx.FullPath(), count)
		if err != nil {
			rejectAuthIP(ctx, decision, err)
			return
		}
		if observation == nil {
			ctx.Next()
			return
		}
		original := ctx.Writer
		writer := &authIPResponseBuffer{ResponseWriter: original, status: http.StatusOK}
		ctx.Writer = writer
		defer func() { ctx.Writer = original }()
		ctx.Next()
		if !writer.committed && authIPCountFailure(writer.Status(), writer.body.Bytes()) {
			decision, err = guard.Failed(ctx.Request.Context(), observation)
			if err != nil {
				writer.reset()
				rejectAuthIP(ctx, decision, err)
			}
		}
		writer.commit()
	}
}

func rejectAuthIP(ctx *gin.Context, decision *service.AuthIPBlacklistDecision, err error) {
	if decision != nil && decision.Entry != nil {
		entry := decision.Entry
		action := "auth.ip_blacklist.reject"
		if decision.NewlyBlocked {
			action = "auth.ip_blacklist.block"
		}
		SetAuditAction(ctx, action)
		ctx.Set(auditCtxKeyExtra, map[string]any{
			"blacklist_id": entry.ID, "target_ip": entry.IP, "generation": entry.Generation,
			"trigger_path": entry.TriggerPath, "trigger_count": entry.TriggerCount,
			"window_seconds": entry.WindowSeconds, "rule_version": entry.RuleVersion,
			"first_event_at": entry.FirstEventAt, "last_event_at": entry.LastEventAt,
			"blocked_until": entry.BlockedUntil,
		})
		if entry.BlockedUntil != nil {
			seconds := int(time.Until(*entry.BlockedUntil).Seconds()) + 1
			if seconds < 1 {
				seconds = 1
			}
			ctx.Header("Retry-After", fmt.Sprint(seconds))
		}
	}
	if infraerrors.Code(err) == http.StatusTooManyRequests {
		ctx.Header("Retry-After", "60")
	}
	response.ErrorFrom(ctx, err)
	ctx.Abort()
}

func authIPCountFailure(status int, body []byte) bool {
	if status != http.StatusBadRequest && status != http.StatusUnauthorized {
		return false
	}
	var result response.Response
	if json.Unmarshal(body, &result) != nil {
		return false
	}
	detail := strings.ToLower(result.Reason + " " + result.Message)
	for _, excluded := range []string{"captcha", "turnstile", "verification code", "验证码", "totp", "expired", "过期"} {
		if strings.Contains(detail, excluded) {
			return false
		}
	}
	if status == http.StatusUnauthorized {
		return result.Reason == "INVALID_CREDENTIALS" || result.Reason == "PASSKEY_VERIFICATION_FAILED"
	}
	return result.Reason == "" && (strings.Contains(detail, "invalid request") || strings.Contains(detail, "invalid passkey response"))
}

type authIPResponseBuffer struct {
	gin.ResponseWriter
	body      bytes.Buffer
	status    int
	written   bool
	committed bool
}

func (writer *authIPResponseBuffer) WriteHeader(status int) {
	if !writer.written && status > 0 {
		writer.status = status
	}
}

func (writer *authIPResponseBuffer) WriteHeaderNow() { writer.written = true }
func (writer *authIPResponseBuffer) Status() int     { return writer.status }
func (writer *authIPResponseBuffer) Written() bool   { return writer.written }
func (writer *authIPResponseBuffer) Size() int {
	if writer.committed {
		return writer.ResponseWriter.Size()
	}
	if !writer.written {
		return -1
	}
	return writer.body.Len()
}

func (writer *authIPResponseBuffer) Write(data []byte) (int, error) {
	writer.written = true
	if writer.committed {
		return writer.ResponseWriter.Write(data)
	}
	if writer.body.Len()+len(data) > 256*1024 {
		writer.commit()
		return writer.ResponseWriter.Write(data)
	}
	return writer.body.Write(data)
}

func (writer *authIPResponseBuffer) WriteString(data string) (int, error) {
	return writer.Write([]byte(data))
}

func (writer *authIPResponseBuffer) reset() {
	writer.body.Reset()
	writer.status = http.StatusOK
	writer.written = false
	writer.Header().Del("Content-Length")
}

func (writer *authIPResponseBuffer) commit() {
	if writer.committed {
		return
	}
	writer.committed = true
	writer.ResponseWriter.WriteHeader(writer.status)
	writer.ResponseWriter.WriteHeaderNow()
	if writer.body.Len() > 0 {
		_, _ = writer.ResponseWriter.Write(writer.body.Bytes())
	}
}

func (writer *authIPResponseBuffer) Flush() {
	writer.commit()
	writer.ResponseWriter.Flush()
}
