package middleware

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// 匿名请求没有登录主体：Global 放行，PublicIPScoped 按公网 IP 与 scope 计数。
func TestPanelRateLimiterPublicIPScoped(t *testing.T) {
	allower := &fakePanelAllower{}
	p := &PanelRateLimiter{
		limiter:        allower,
		settingService: newPanelRateLimitTestService(t, `{"enabled":true,"user_rpm":1,"heavy_rpm":1,"exempt_admin":true,"public_ip_rpm":1}`),
	}
	global := newPanelTestRouter(p.Global(), nil)
	for range 3 {
		require.Equal(t, http.StatusOK, performPanelRequest(global, "203.0.113.9:1000").Code, "按用户计数的限流不保护匿名接口")
	}
	gallery := newPanelTestRouter(p.PublicIPScoped("studio-gallery"), nil)
	media := newPanelTestRouter(p.PublicIPScoped("studio-media"), nil)
	require.Equal(t, http.StatusOK, performPanelRequest(gallery, "203.0.113.9:1000").Code)
	require.Equal(t, http.StatusTooManyRequests, performPanelRequest(gallery, "203.0.113.9:1000").Code)
	require.Equal(t, http.StatusOK, performPanelRequest(media, "203.0.113.9:1000").Code, "不同 scope 分开计数")
	require.Equal(t, http.StatusOK, performPanelRequest(gallery, "198.51.100.7:1000").Code, "不同 IP 分开计数")
	for range 3 {
		require.Equal(t, http.StatusOK, performPanelRequest(gallery, "10.0.0.8:1000").Code, "内网转发地址不计数")
	}
	allower.mu.Lock()
	defer allower.mu.Unlock()
	require.Contains(t, allower.counts, "panel:public:studio-gallery:ip:203.0.113.9")
	require.Contains(t, allower.counts, "panel:public:studio-media:ip:203.0.113.9")
	require.NotContains(t, allower.counts, "panel:public:ip:203.0.113.9", "不挤占共用的公开接口额度")
}
