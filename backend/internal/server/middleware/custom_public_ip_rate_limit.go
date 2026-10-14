package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// PublicIPScoped 与 PublicIP 相同，但按 scope 单独计数。
// [CUSTOM] 匿名可访问的定制接口（工坊公开画廊／媒体）用它限流：Global/Heavy 按用户计数，
// 没有登录主体时会直接放行，不能保护匿名接口；单独分桶避免画廊浏览挤占登录等公开接口的额度。
func (p *PanelRateLimiter) PublicIPScoped(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if p == nil || p.limiter == nil || p.settingService == nil {
			c.Next()
			return
		}
		settings := p.settingService.GetPanelRateLimitSettingsCached(c.Request.Context())
		if !settings.Enabled || settings.PublicIPRPM <= 0 {
			c.Next()
			return
		}
		clientIP := SecurityClientIP(c)
		if !isPubliclyRoutableClientIP(clientIP) {
			c.Next()
			return
		}
		result, err := p.limiter.Allow(c.Request.Context(), "panel:public:"+scope+":ip:"+clientIP, settings.PublicIPRPM, panelRateLimitWindow)
		if err != nil {
			slog.Warn("panel scoped public rate limit check failed, allowing request", "scope", scope, "error", err)
			c.Next()
			return
		}
		if !result.Allowed {
			abortPanelRateLimited(c, result.RetryAfter)
			return
		}
		c.Next()
	}
}
