//go:build unit

package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCustomKiroRejectsUnsupportedIngressSemantics(t *testing.T) {
	for _, test := range []struct{ format, body string }{
		{"responses", `{"previous_response_id":"resp_old"}`},
		{"responses", `{"store":true}`},
		{"responses", `{"background":true}`},
		{"responses", `{"conversation":"conv_old"}`},
		{"chat", `{"n":2}`},
		{"chat", `{"response_format":{"type":"json_schema"}}`},
	} {
		require.Error(t, validateCustomKiroIngress([]byte(test.body), test.format))
	}
	require.NoError(t, validateCustomKiroIngress([]byte(`{"store":false,"input":"hi"}`), "responses"))
}

func TestCustomKiroRealGatewayEntrypoints(t *testing.T) {
	for _, protocol := range []string{"messages", "chat", "responses"} {
		for _, stream := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/stream=%t", protocol, stream), func(t *testing.T) {
				account := customKiroAccount()
				calls := 0
				service, _ := customKiroTestService(account, customKiroHTTP{send: func(request *http.Request, proxy string, accountID int64, concurrency int) (*http.Response, error) {
					calls++
					require.Equal(t, "runtime.us-east-1.kiro.dev", request.URL.Host)
					require.Equal(t, "Bearer access-old", request.Header.Get("Authorization"))
					require.Empty(t, request.Header.Get("anthropic-version"))
					return customKiroStreamResponse(customKiroStream(true, true)), nil
				}})
				gateway := &GatewayService{customKiro: service, cfg: &config.Config{}}
				payload := fmt.Sprintf(`{"model":"claude-sonnet-4-6","stream":%t,"max_tokens":100,"messages":[{"role":"user","content":"hi"}]}`, stream)
				if protocol == "responses" {
					payload = fmt.Sprintf(`{"model":"claude-sonnet-4-6","stream":%t,"input":"hi"}`, stream)
				}
				recorder := httptest.NewRecorder()
				contextGin, _ := gin.CreateTestContext(recorder)
				contextGin.Request = httptest.NewRequest(http.MethodPost, "/v1/"+protocol, strings.NewReader(payload))
				var result *ForwardResult
				var err error
				switch protocol {
				case "messages":
					parsed, parseErr := ParseGatewayRequest(NewRequestBodyRef([]byte(payload)), "anthropic")
					require.NoError(t, parseErr)
					result, err = gateway.Forward(context.Background(), contextGin, account, parsed)
				case "chat":
					result, err = gateway.ForwardAsChatCompletions(context.Background(), contextGin, account, []byte(payload), nil)
				case "responses":
					result, err = gateway.ForwardAsResponses(context.Background(), contextGin, account, []byte(payload), nil)
				}
				require.NoError(t, err)
				require.Equal(t, 1, calls)
				require.Equal(t, http.StatusOK, recorder.Code)
				require.Contains(t, recorder.Body.String(), "hello")
				require.Contains(t, recorder.Body.String(), "lookup")
				require.Equal(t, ClaudeUsage{InputTokens: 13, OutputTokens: 17, CacheReadInputTokens: 3, CacheCreationInputTokens: 2}, result.Usage)
				require.Empty(t, result.UpstreamResponseModel)
			})
		}
	}
}

func TestCustomKiroDoesNotUndoUnappliedToolRewrites(t *testing.T) {
	for _, format := range []string{"messages", "chat", "responses"} {
		for _, stream := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/%t", format, stream), func(t *testing.T) {
				account := customKiroAccount()
				service, _ := customKiroTestService(account, customKiroHTTP{send: func(*http.Request, string, int64, int) (*http.Response, error) {
					body := customKiroFrame("toolUseEvent", map[string]any{"toolUseId": "call", "name": "cc_sess_lookup", "input": map[string]any{}, "stop": true})
					body = append(body, customKiroFrame("metadataEvent", map[string]any{"tokenUsage": map[string]int{"uncachedInputTokens": 13, "outputTokens": 17}})...)
					return customKiroStreamResponse(body), nil
				}})
				recorder := httptest.NewRecorder()
				contextGin, _ := gin.CreateTestContext(recorder)
				contextGin.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
				_, err := (&GatewayService{customKiro: service}).forwardCustomKiro(context.Background(), contextGin, account, []byte(`{"messages":[{"role":"user","content":"hi"}]}`), "claude-sonnet-4-6", stream, format, false, nil)
				require.NoError(t, err)
				require.Contains(t, recorder.Body.String(), "cc_sess_lookup")
				require.NotContains(t, recorder.Body.String(), "sessions_lookup")
			})
		}
	}
}

func TestCustomKiroGenericCapabilitiesDoNotSendCredentials(t *testing.T) {
	for _, accountType := range []string{AccountTypeOAuth, AccountTypeAPIKey} {
		account := customKiroAccount()
		account.Type = accountType
		service := &AccountTestService{httpUpstream: customKiroHTTP{send: func(*http.Request, string, int64, int) (*http.Response, error) {
			t.Fatal("通用模型同步不应发送 Kiro 凭据")
			return nil, errors.New("unexpected request")
		}}}
		_, err := service.SyncUpstreamModelCatalog(context.Background(), account)
		require.Error(t, err)
		var unsupported *UpstreamModelSyncError
		require.ErrorAs(t, err, &unsupported)
		require.Equal(t, UpstreamModelSyncErrorUnsupported, unsupported.Kind)
		require.False(t, IsUpstreamBillingProbeIdentity(account.Platform, account.Type))
	}
}

func TestCustomKiroEligibilityRecheckedAfterRotation(t *testing.T) {
	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)
	changes := map[string]func(*Account){
		"disabled":         func(account *Account) { account.Status = "inactive" },
		"unschedulable":    func(account *Account) { account.Schedulable = false },
		"expired":          func(account *Account) { account.ExpiresAt = &past },
		"rate limited":     func(account *Account) { account.RateLimitResetAt = &future },
		"overloaded":       func(account *Account) { account.OverloadUntil = &future },
		"cooldown":         func(account *Account) { account.TempUnschedulableUntil = &future },
		"platform changed": func(account *Account) { account.Platform = PlatformAnthropic },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			account := customKiroAccount()
			account.Credentials["expires_at"] = past.Unix()
			service, repository := customKiroTestService(account, customKiroHTTP{send: func(*http.Request, string, int64, int) (*http.Response, error) {
				return customKiroJSONResponse(200, `{"accessToken":"new","expiresIn":3600}`), nil
			}})
			repository.persistHook = change
			_, credentials, err := service.prepare(context.Background(), account, "")
			require.Error(t, err)
			require.Empty(t, credentials.AccessToken)
		})
	}
}

func TestCustomKiroAdminProbeCanTestDisabledAccount(t *testing.T) {
	account := customKiroAccount()
	account.Status = "inactive"
	account.Schedulable = false
	calls := 0
	service, _ := customKiroTestService(account, customKiroHTTP{send: func(*http.Request, string, int64, int) (*http.Response, error) {
		calls++
		return customKiroStreamResponse(customKiroStream(true, false)), nil
	}})
	body := []byte(`{"messages":[{"role":"user","content":"hi"}]}`)
	_, err := service.call(context.Background(), account, body, "claude-sonnet-4-6")
	require.Error(t, err)
	require.Zero(t, calls)
	ctx := context.WithValue(context.Background(), customKiroProbeContextKey{}, true)
	call, err := service.call(ctx, account, body, "claude-sonnet-4-6")
	require.NoError(t, err)
	require.NoError(t, call.Response.Body.Close())
	require.Equal(t, 1, calls)
}

func TestCustomKiroMergePreservesSettingsAndCanonicalizesReplacement(t *testing.T) {
	account := customKiroAccount()
	merged, err := mergeCustomKiroCredentials(account.Credentials, map[string]any{"runtime_region": "ap-northeast-1"})
	require.NoError(t, err)
	require.Equal(t, "access-old", merged["access_token"])
	require.Equal(t, "eu-west-1", merged["auth_region"])
	merged, err = mergeCustomKiroCredentials(merged, map[string]any{"refreshToken": "admin-rotated"})
	require.NoError(t, err)
	require.Equal(t, "admin-rotated", merged["refresh_token"])
	require.NotContains(t, merged, "access_token")
	require.NotContains(t, merged, "profile_arn")
	require.NotContains(t, merged, "expires_at")
	require.Equal(t, "ap-northeast-1", merged["runtime_region"])
	_, err = normalizeCustomKiroCredentials(AccountTypeOAuth, merged)
	require.NoError(t, err)
}

func TestCustomKiroHTTPFailuresAndInvalidStreams(t *testing.T) {
	cases := []struct {
		name     string
		response func() *http.Response
		failover bool
	}{
		{"rate limit", func() *http.Response { return customKiroJSONResponse(429, "upstream secret") }, true},
		{"server failure", func() *http.Response { return customKiroJSONResponse(503, "upstream secret") }, true},
		{"bad media", func() *http.Response { return customKiroJSONResponse(200, "upstream secret") }, false},
		{"truncated frame", func() *http.Response {
			body := customKiroStream(true, false)
			return customKiroStreamResponse(body[:len(body)-3])
		}, false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			account := customKiroAccount()
			service, _ := customKiroTestService(account, customKiroHTTP{send: func(*http.Request, string, int64, int) (*http.Response, error) { return test.response(), nil }})
			recorder := httptest.NewRecorder()
			contextGin, _ := gin.CreateTestContext(recorder)
			contextGin.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
			_, err := (&GatewayService{customKiro: service}).forwardCustomKiro(context.Background(), contextGin, account, []byte(`{"messages":[{"role":"user","content":"hi"}]}`), "claude-sonnet-4-6", true, "messages", false, nil)
			require.Error(t, err)
			var failover *UpstreamFailoverError
			require.Equal(t, test.failover, errors.As(err, &failover))
			require.NotContains(t, recorder.Body.String(), "upstream secret")
			require.NotContains(t, recorder.Body.String(), "event: message_stop")
		})
	}
}

type customKiroDisconnectWriter struct{ gin.ResponseWriter }

func (writer customKiroDisconnectWriter) Write(data []byte) (int, error) { return 0, io.ErrClosedPipe }
func (writer customKiroDisconnectWriter) WriteString(data string) (int, error) {
	return 0, io.ErrClosedPipe
}

func TestCustomKiroClientDisconnectStillCollectsRealUsage(t *testing.T) {
	account := customKiroAccount()
	service, _ := customKiroTestService(account, customKiroHTTP{send: func(request *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
		require.NoError(t, request.Context().Err())
		return customKiroStreamResponse(customKiroStream(true, false)), nil
	}})
	recorder := httptest.NewRecorder()
	contextGin, _ := gin.CreateTestContext(recorder)
	contextGin.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	contextGin.Writer = customKiroDisconnectWriter{contextGin.Writer}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := (&GatewayService{customKiro: service}).forwardCustomKiro(ctx, contextGin, account, []byte(`{"messages":[{"role":"user","content":"hi"}]}`), "claude-sonnet-4-6", true, "messages", false, nil)
	require.NoError(t, err)
	require.True(t, result.ClientDisconnect)
	require.Equal(t, 17, result.Usage.OutputTokens)
}
