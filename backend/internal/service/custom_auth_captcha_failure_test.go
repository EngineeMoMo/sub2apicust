//go:build unit

package service

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestCustomAuthCaptchaMissingProofMetadata(t *testing.T) {
	for _, provider := range []string{"turnstile", "tencent", "aliyun"} {
		t.Run(provider, func(t *testing.T) {
			turnstile := &turnstileVerifierSpy{}
			tencent := &tencentCaptchaVerifierStub{}
			aliyun := &aliyunVerifierSpy{}
			settings := map[string]string{SettingKeyTurnstileEnabled: "true", SettingKeyTurnstileSecretKey: "local-test-secret"}
			if provider == "tencent" {
				settings = tencentCaptchaSettings()
			}
			if provider == "aliyun" {
				settings = aliyunEnabledSettings()
			}
			svc := newAuthServiceForCaptchaTest(settings, false, turnstile, tencent)
			svc.SetAliyunCaptchaService(NewAliyunCaptchaService(nil, aliyun))
			methods := []func(context.Context, CaptchaProof, string) error{svc.VerifyCaptcha}
			if provider != "turnstile" {
				methods = append(methods, svc.VerifyActionCaptchaIfEnabled)
			}
			for _, verify := range methods {
				err := verify(context.Background(), CaptchaProof{}, "198.51.100.24")
				require.Equal(t, 400, infraerrors.Code(err))
				require.Equal(t, "missing_captcha_proof", infraerrors.FromError(err).Metadata["auth_ip_failure"])
			}
			require.Zero(t, turnstile.called)
			require.Zero(t, tencent.calls)
			require.Zero(t, aliyun.called)
		})
	}
}

func TestCustomAuthCaptchaDisabledDoesNotRequireProof(t *testing.T) {
	svc := newAuthServiceForCaptchaTest(map[string]string{}, false, nil, nil)
	require.NoError(t, svc.VerifyCaptcha(context.Background(), CaptchaProof{}, "198.51.100.24"))
	svc = newAliyunAuthServiceForTest(&config.Config{}, map[string]string{}, &aliyunVerifierSpy{})
	require.NoError(t, svc.VerifyActionCaptchaIfEnabled(context.Background(), CaptchaProof{}, "198.51.100.24"))
}

func TestCustomAuthCaptchaSubmittedProofIsNotMarked(t *testing.T) {
	for _, provider := range []string{"turnstile", "tencent", "aliyun"} {
		for _, outcome := range []string{"rejected", "network_error", "success", "not_configured", "no_verifier"} {
			t.Run(provider+"/"+outcome, func(t *testing.T) {
				settings := map[string]string{SettingKeyTurnstileEnabled: "true", SettingKeyTurnstileSecretKey: "local-test-secret"}
				secretKey := SettingKeyTurnstileSecretKey
				if provider == "tencent" {
					settings, secretKey = tencentCaptchaSettings(), SettingKeyTencentCaptchaCloudSecretKey
				}
				if provider == "aliyun" {
					settings, secretKey = aliyunEnabledSettings(), SettingKeyAliyunCaptchaAccessKeySecret
				}
				turnstile := &turnstileVerifierSpy{result: &TurnstileVerifyResponse{Success: outcome == "success"}}
				tencent := &tencentCaptchaVerifierStub{response: &TencentCaptchaVerifyResponse{}}
				aliyun := &aliyunVerifierSpy{result: &AliyunCaptchaVerifyResult{VerifyResult: outcome == "success"}}
				if outcome == "success" {
					tencent.response.CaptchaCode = 1
				}
				if outcome == "network_error" {
					turnstile.err, tencent.err, aliyun.err = errors.New("verifier offline"), errors.New("verifier offline"), errors.New("verifier offline")
				}
				if outcome == "not_configured" {
					delete(settings, secretKey)
				}
				svc := newAuthServiceForCaptchaTest(settings, false, turnstile, tencent)
				svc.SetAliyunCaptchaService(NewAliyunCaptchaService(nil, aliyun))
				proof := CaptchaProof{TurnstileToken: "submitted-proof", TencentTicket: "submitted-ticket", TencentRandstr: "submitted-randstr"}
				if outcome == "no_verifier" {
					svc.turnstileService.verifier = nil
					svc.tencentCaptchaService.verifier = nil
					svc.aliyunCaptchaService.verifier = nil
				}
				if outcome == "not_configured" || outcome == "no_verifier" {
					proof = CaptchaProof{}
				}
				methods := []func(context.Context, CaptchaProof, string) error{svc.VerifyCaptcha}
				if provider != "turnstile" {
					methods = append(methods, svc.VerifyActionCaptchaIfEnabled)
				}
				for _, verify := range methods {
					err := verify(context.Background(), proof, "198.51.100.24")
					if outcome == "success" {
						require.NoError(t, err)
						continue
					}
					require.Error(t, err)
					require.Empty(t, infraerrors.FromError(err).Metadata[AuthIPFailureMetadataKey])
					if outcome == "not_configured" || outcome == "no_verifier" {
						require.Equal(t, http.StatusServiceUnavailable, infraerrors.Code(err))
					} else if provider != "turnstile" || outcome == "rejected" {
						require.Equal(t, http.StatusBadRequest, infraerrors.Code(err))
					}
				}
			})
		}
	}
}

func TestCustomAuthCaptchaTencentPartialAndDisasterProof(t *testing.T) {
	for _, test := range []struct {
		name    string
		proof   CaptchaProof
		missing bool
	}{
		{name: "缺少随机串", proof: CaptchaProof{TencentTicket: "ticket"}, missing: true},
		{name: "缺少票据", proof: CaptchaProof{TencentRandstr: "randstr"}, missing: true},
		{name: "空白字段", proof: CaptchaProof{TencentTicket: "  ", TencentRandstr: "  "}, missing: true},
		{name: "灾备票据", proof: CaptchaProof{TencentTicket: "trerror_fallback", TencentRandstr: "randstr"}},
		{name: "灾备票据缺少随机串", proof: CaptchaProof{TencentTicket: " trerror_fallback "}},
	} {
		t.Run(test.name, func(t *testing.T) {
			verifier := &tencentCaptchaVerifierStub{}
			svc := newAuthServiceForCaptchaTest(tencentCaptchaSettings(), false, nil, verifier)
			for _, verify := range []func(context.Context, CaptchaProof, string) error{svc.VerifyCaptcha, svc.VerifyActionCaptchaIfEnabled} {
				err := verify(context.Background(), test.proof, "198.51.100.24")
				require.ErrorIs(t, err, ErrTencentCaptchaVerificationFailed)
				require.Equal(t, test.missing, infraerrors.FromError(err).Metadata[AuthIPFailureMetadataKey] == AuthIPFailureMissingCaptchaProof)
			}
			require.Zero(t, verifier.calls)
		})
	}
}

func TestCustomAuthCaptchaProofErrorPreservesErrors(t *testing.T) {
	original := ErrTurnstileVerificationFailed.WithMetadata(map[string]string{"existing": "value"})
	marked := customAuthCaptchaProofError(original, true)
	require.ErrorIs(t, marked, original)
	require.Equal(t, infraerrors.Message(original), infraerrors.Message(marked))
	require.Equal(t, "value", infraerrors.FromError(marked).Metadata["existing"])
	require.Empty(t, original.Metadata[AuthIPFailureMetadataKey])
	require.Empty(t, ErrTurnstileVerificationFailed.Metadata)
	for _, originalError := range []error{nil, ErrTurnstileNotConfigured, errors.New("offline"), infraerrors.BadRequest("UNRELATED", "captcha failed")} {
		require.True(t, originalError == customAuthCaptchaProofError(originalError, true))
	}
}
