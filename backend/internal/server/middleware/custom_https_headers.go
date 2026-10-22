package middleware

import (
	"net"
	"net/netip"
	"strings"

	"github.com/gin-gonic/gin"
)

func CustomHTTPSHeaders(trustedProxies []string) gin.HandlerFunc {
	prefixes := make([]netip.Prefix, 0, len(trustedProxies))
	for _, raw := range trustedProxies {
		value := strings.TrimSpace(raw)
		if prefix, err := netip.ParsePrefix(value); err == nil {
			prefixes = append(prefixes, prefix)
		} else if address, addressError := netip.ParseAddr(value); addressError == nil {
			prefixes = append(prefixes, netip.PrefixFrom(address, address.BitLen()))
		}
	}
	return func(request *gin.Context) {
		secure := request.Request.TLS != nil
		if !secure && strings.EqualFold(strings.TrimSpace(request.GetHeader("X-Forwarded-Proto")), "https") {
			peer, _, splitError := net.SplitHostPort(request.Request.RemoteAddr)
			if splitError != nil {
				peer = request.Request.RemoteAddr
			}
			if address, err := netip.ParseAddr(peer); err == nil {
				for _, prefix := range prefixes {
					if prefix.Contains(address.Unmap()) {
						secure = true
						break
					}
				}
			}
		}
		if secure {
			request.Header("Strict-Transport-Security", "max-age=31536000")
		}
		request.Next()
	}
}
