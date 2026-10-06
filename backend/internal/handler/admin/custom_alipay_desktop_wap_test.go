//go:build unit

package admin

import (
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestCustomAlipayDesktopWapSettingsRoundTrip(t *testing.T) {
	h, repo := newStepUpSwitchTestHandler(t, map[string]string{service.SettingAlipayForceQRCode: "false", service.SettingAlipayMobilePrecreateDeepLink: "true"})
	h.paymentConfigService = service.NewPaymentConfigService(nil, repo, nil)
	for _, enabled := range []bool{true, false} {
		rec := doUpdateSettings(t, h, map[string]any{"payment_alipay_desktop_wap_qrcode": enabled}, nil)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), `"payment_alipay_desktop_wap_qrcode":`)
		require.Equal(t, enabled, repo.values[service.SettingAlipayDesktopWapQRCode] == "true")
	}
	repo.values[service.SettingAlipayDesktopWapQRCode] = "true"
	rec := doUpdateSettings(t, h, map[string]any{"payment_min_amount": 5}, nil)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "true", repo.values[service.SettingAlipayDesktopWapQRCode])
	require.Equal(t, "false", repo.values[service.SettingAlipayForceQRCode])
	require.Equal(t, "true", repo.values[service.SettingAlipayMobilePrecreateDeepLink])
}
