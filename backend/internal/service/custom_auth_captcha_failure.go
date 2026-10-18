package service

import (
	"net/http"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	AuthIPFailureMetadataKey         = "auth_ip_failure"
	AuthIPFailureMissingCaptchaProof = "missing_captcha_proof"
)

func customAuthCaptchaProofError(err error, missing bool) error {
	if !missing || infraerrors.Code(err) != http.StatusBadRequest {
		return err
	}
	switch infraerrors.Reason(err) {
	case "TURNSTILE_VERIFICATION_FAILED", "TENCENT_CAPTCHA_VERIFICATION_FAILED", "ALIYUN_CAPTCHA_VERIFICATION_FAILED":
	default:
		return err
	}
	appError := infraerrors.FromError(err)
	metadata := make(map[string]string, len(appError.Metadata)+1)
	for key, value := range appError.Metadata {
		metadata[key] = value
	}
	metadata[AuthIPFailureMetadataKey] = AuthIPFailureMissingCaptchaProof
	return appError.WithMetadata(metadata)
}

func customAuthTencentCaptchaProofMissing(proof CaptchaProof) bool {
	ticket := strings.TrimSpace(proof.TencentTicket)
	return !strings.HasPrefix(ticket, "trerror_") && (ticket == "" || strings.TrimSpace(proof.TencentRandstr) == "")
}
