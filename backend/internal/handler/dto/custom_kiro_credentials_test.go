package dto

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCustomKiroCredentialsRedacted(t *testing.T) {
	credentials := map[string]any{"access_token": "access", "refresh_token": "refresh", "client_secret": "secret", "accessToken": "alias-access", "refreshToken": "alias-refresh", "clientSecret": "alias-secret", "client_id": "public-client"}
	redacted, status := RedactCredentials(credentials)
	for _, key := range []string{"access_token", "refresh_token", "client_secret", "accessToken", "refreshToken", "clientSecret"} {
		require.NotContains(t, redacted, key)
		require.True(t, status["has_"+key])
		require.NotEmpty(t, credentials[key])
	}
	require.Equal(t, "public-client", redacted["client_id"])
}
