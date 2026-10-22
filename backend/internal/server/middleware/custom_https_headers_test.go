package middleware

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCustomHTTPSHeaders(t *testing.T) {
	for _, scenario := range []struct {
		name           string
		peer           string
		forwardedProto string
		tls            bool
		secure         bool
	}{
		{"direct_tls", "203.0.113.5:443", "", true, true},
		{"plain_local_http", "127.0.0.1:5000", "", false, false},
		{"trusted_ipv4", "127.0.0.1:5000", "https", false, true},
		{"trusted_ipv6", "[::1]:5000", "https", false, true},
		{"untrusted_forwarded_proto", "203.0.113.5:5000", "https", false, false},
		{"ambiguous_forwarded_proto", "127.0.0.1:5000", "http, https", false, false},
		{"forwarded_http", "127.0.0.1:5000", "http", false, false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			router := gin.New()
			router.Use(CustomHTTPSHeaders([]string{"127.0.0.1/32", "::1/128"}))
			router.GET("/", func(request *gin.Context) { request.AbortWithStatus(http.StatusUnauthorized) })
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.RemoteAddr = scenario.peer
			request.Header.Set("X-Forwarded-Proto", scenario.forwardedProto)
			request.Header.Set("X-Forwarded-For", "127.0.0.1")
			if scenario.tls {
				request.TLS = &tls.ConnectionState{}
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.Equal(t, http.StatusUnauthorized, response.Code)
			if scenario.secure {
				require.Equal(t, "max-age=31536000", response.Header().Get("Strict-Transport-Security"))
			} else {
				require.Empty(t, response.Header().Get("Strict-Transport-Security"))
			}
		})
	}
}
