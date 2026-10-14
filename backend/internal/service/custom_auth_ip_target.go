package service

import (
	"net/netip"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

func (config AuthIPBlacklistConfig) WithAggregationDefaults() AuthIPBlacklistConfig {
	if config.AggregationMode == "" {
		config.AggregationMode = AuthIPAggregationSingle
	}
	if config.NetworkIPv4Prefix == 0 {
		config.NetworkIPv4Prefix = 24
	}
	if config.NetworkIPv6Prefix == 0 {
		config.NetworkIPv6Prefix = 64
	}
	return config
}

func AuthIPAggregationTarget(ip string, config AuthIPBlacklistConfig) (string, error) {
	canonical, err := NormalizeAuthIP(ip)
	if err != nil {
		return "", err
	}
	config = config.WithAggregationDefaults()
	if config.AggregationMode == AuthIPAggregationSingle {
		return canonical, nil
	}
	address := netip.MustParseAddr(canonical)
	bits := config.NetworkIPv6Prefix
	if address.Is4() {
		bits = config.NetworkIPv4Prefix
	}
	if config.AggregationMode != AuthIPAggregationNetwork || bits < 1 || bits > address.BitLen() {
		return "", infraerrors.BadRequest("IP_BLACKLIST_INVALID_CONFIG", "自动封禁网段配置无效")
	}
	if bits == address.BitLen() {
		return canonical, nil
	}
	return netip.PrefixFrom(address, bits).Masked().String(), nil
}

func NormalizeAuthIPTarget(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	invalid := infraerrors.BadRequest("IP_BLACKLIST_INVALID_IP", "请输入完整 IP、末尾连续通配符（如210.230.206.*）或有效CIDR网段，不支持全网封禁")
	if strings.Contains(raw, "*") {
		parts := strings.Split(raw, ".")
		if len(parts) != 4 {
			return "", invalid
		}
		firstWildcard := -1
		for index, part := range parts {
			if part == "*" {
				if firstWildcard < 0 {
					firstWildcard = index
				}
				parts[index] = "0"
			} else if firstWildcard >= 0 {
				return "", invalid
			}
		}
		if firstWildcard < 1 {
			return "", invalid
		}
		address, err := netip.ParseAddr(strings.Join(parts, "."))
		if err != nil || !address.Is4() {
			return "", invalid
		}
		return netip.PrefixFrom(address, firstWildcard*8).String(), nil
	}
	if !strings.Contains(raw, "/") {
		return NormalizeAuthIP(raw)
	}
	prefix, err := netip.ParsePrefix(raw)
	if err != nil {
		return "", invalid
	}
	if prefix.Addr().Is4In6() {
		if prefix.Bits() <= 96 {
			return "", invalid
		}
		prefix = netip.PrefixFrom(prefix.Addr().Unmap(), prefix.Bits()-96)
	}
	if prefix.Bits() == 0 {
		return "", invalid
	}
	if prefix.Bits() == prefix.Addr().BitLen() {
		return prefix.Addr().String(), nil
	}
	return prefix.Masked().String(), nil
}
