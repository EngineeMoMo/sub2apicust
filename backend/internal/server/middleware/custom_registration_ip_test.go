package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRegistrationIPContextTrustedProxyOnly(t *testing.T) {
	for _, scenario := range []struct {
		name      string
		trusted   []string
		forwarded string
		want      string
	}{
		{"untrusted forwarding header ignored", nil, "192.0.2.9", "198.51.100.1"},
		{"trusted proxy resolves client", []string{"198.51.100.1"}, "192.0.2.9", "192.0.2.9"},
		{"forged leftmost IP ignored", []string{"198.51.100.1"}, "203.0.113.8, 192.0.2.9", "192.0.2.9"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			router := gin.New()
			require.NoError(t, router.SetTrustedProxies(scenario.trusted))
			router.Use(RegistrationIPContext())
			router.POST("/register", func(ctx *gin.Context) {
				address, present := service.RegistrationIPFromContext(ctx.Request.Context())
				require.True(t, present)
				require.Equal(t, scenario.want, address)
				ctx.Status(http.StatusNoContent)
			})
			request := httptest.NewRequest(http.MethodPost, "/register", nil)
			request.RemoteAddr = "198.51.100.1:1234"
			request.Header.Set("X-Forwarded-For", scenario.forwarded)
			request.Header.Set("X-Real-IP", "203.0.113.99")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.Equal(t, http.StatusNoContent, response.Code)
		})
	}
}
