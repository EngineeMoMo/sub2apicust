package service

import (
	"context"
	"net/netip"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const SettingKeyRegistrationIPLimitEnabled = "registration_ip_limit_enabled"

var (
	ErrRegistrationIPLimit       = infraerrors.Forbidden("REGISTRATION_IP_LIMIT", "当前 IP 已注册过账号，请登录已有账号或联系管理员")
	ErrRegistrationIPUnavailable = infraerrors.ServiceUnavailable("REGISTRATION_IP_UNAVAILABLE", "无法核验注册来源，请稍后重试或联系管理员")
)

type registrationIPContextKey struct{}

func WithRegistrationIP(ctx context.Context, address string) context.Context {
	ip, err := netip.ParseAddr(strings.TrimSpace(address))
	canonical := ""
	if err == nil && ip.Zone() == "" && !ip.IsUnspecified() && !ip.IsMulticast() {
		canonical = ip.Unmap().String()
	}
	return context.WithValue(ctx, registrationIPContextKey{}, canonical)
}

func RegistrationIPFromContext(ctx context.Context) (string, bool) {
	address, present := ctx.Value(registrationIPContextKey{}).(string)
	return address, present
}
