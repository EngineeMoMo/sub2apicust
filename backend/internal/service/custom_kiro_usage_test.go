//go:build unit

package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/kiro"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestCustomKiroUsageModeGatewayProtocols(t *testing.T) {
	for _, format := range []string{"messages", "chat", "responses"} {
		for _, stream := range []bool{false, true} {
			for _, source := range []string{"estimated", "metadata", "metering"} {
				t.Run(format+"/"+map[bool]string{true: "stream", false: "buffered"}[stream]+"/"+source, func(t *testing.T) {
					account := customKiroAccount()
					account.Credentials["usage_mode"] = "estimated"
					upstream := customKiroStream(source == "metadata", true)
					if source == "metering" {
						upstream = append(upstream, customKiroFrame("meteringEvent", map[string]int{"inputTokens": 23, "outputTokens": 8})...)
					}
					service, _ := customKiroTestService(account, customKiroHTTP{send: func(*http.Request, string, int64, int) (*http.Response, error) {
						return customKiroStreamResponse(upstream), nil
					}})
					recorder := httptest.NewRecorder()
					contextGin, _ := gin.CreateTestContext(recorder)
					contextGin.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
					result, err := (&GatewayService{customKiro: service}).forwardCustomKiro(context.Background(), contextGin, account, []byte("{\"messages\":[{\"role\":\"user\",\"content\":\"hi\"}]}"), "claude-sonnet-4-5-20250929", stream, format, true, nil)
					require.NoError(t, err)
					require.Equal(t, http.StatusOK, recorder.Code)
					require.Positive(t, result.Usage.InputTokens)
					require.Positive(t, result.Usage.OutputTokens)
					require.Contains(t, recorder.Body.String(), "\"kiro_usage_source\":\""+source+"\"")
					require.Equal(t, "estimated", recorder.Header().Get("X-Sub2API-Kiro-Usage-Policy"))
					if source == "metadata" {
						require.Equal(t, ClaudeUsage{InputTokens: 13, OutputTokens: 17, CacheReadInputTokens: 3, CacheCreationInputTokens: 2}, result.Usage)
					} else {
						require.Zero(t, result.Usage.CacheReadInputTokens)
						require.Zero(t, result.Usage.CacheCreationInputTokens)
					}
					if !stream {
						require.Equal(t, source, recorder.Header().Get("X-Sub2API-Kiro-Usage-Source"))
						require.True(t, json.Valid(recorder.Body.Bytes()))
					}
					require.NotContains(t, recorder.Body.String(), "access-old")
				})
			}
		}
	}
}

func TestCustomKiroEstimatedModeRejectsDamagedStreams(t *testing.T) {
	valid := customKiroStream(false, false)
	corrupt := append([]byte{}, valid...)
	corrupt[len(corrupt)-1] ^= 1
	malformed := append(customKiroStream(false, false), customKiroFrame("metadataEvent", map[string]any{"tokenUsage": map[string]int{"outputTokens": 2}})...)
	for name, upstream := range map[string][]byte{"truncated": valid[:len(valid)-1], "crc": corrupt, "invalid usage": malformed, "empty": nil} {
		t.Run(name, func(t *testing.T) {
			account := customKiroAccount()
			account.Credentials["usage_mode"] = "estimated"
			service, _ := customKiroTestService(account, customKiroHTTP{send: func(*http.Request, string, int64, int) (*http.Response, error) {
				return customKiroStreamResponse(upstream), nil
			}})
			recorder := httptest.NewRecorder()
			contextGin, _ := gin.CreateTestContext(recorder)
			contextGin.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
			_, err := (&GatewayService{customKiro: service}).forwardCustomKiro(context.Background(), contextGin, account, []byte("{\"messages\":[{\"role\":\"user\",\"content\":\"hi\"}]}"), "claude-sonnet-4-5-20250929", false, "messages", false, nil)
			require.Error(t, err)
			require.Equal(t, http.StatusBadGateway, recorder.Code)
			require.NotContains(t, recorder.Body.String(), "kiro_usage_source")
		})
	}
}

func TestCustomKiroConnectionTestSeparatesBillingReadiness(t *testing.T) {
	for _, source := range []string{"missing", "estimated", "metadata"} {
		t.Run(source, func(t *testing.T) {
			account := customKiroAccount()
			if source == "estimated" {
				account.Credentials["usage_mode"] = "estimated"
			}
			service, _ := customKiroTestService(account, customKiroHTTP{send: func(*http.Request, string, int64, int) (*http.Response, error) {
				return customKiroStreamResponse(customKiroStream(source == "metadata", false)), nil
			}})
			recorder := httptest.NewRecorder()
			contextGin, _ := gin.CreateTestContext(recorder)
			contextGin.Request = httptest.NewRequest(http.MethodPost, "/test", nil)
			require.NoError(t, (&AccountTestService{customKiro: service}).testCustomKiroConnection(contextGin, account, "claude-sonnet-4-5-20250929", "hi"))
			events := strings.Split(recorder.Body.String(), "\n")
			var completion gjson.Result
			for _, line := range events {
				if strings.HasPrefix(line, "data:") {
					payload := gjson.Parse(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
					if payload.Get("type").String() == "test_complete" {
						completion = payload
					}
				}
			}
			require.True(t, completion.Get("success").Bool())
			require.Equal(t, source, completion.Get("data.usage_source").String())
			require.Equal(t, source != "missing", completion.Get("data.billing_ready").Bool())
			require.EqualValues(t, 1, completion.Get("data.event_counts.assistantResponseEvent").Int())
			if source != "metadata" {
				require.Contains(t, recorder.Body.String(), "\"status\":\"warning\"")
			}
		})
	}
}

func TestCustomKiroUsageAnnotationPreservesToolNumbers(t *testing.T) {
	encoded, err := customKiroAnnotateUsage([]byte("{\"tool\":{\"id\":9007199254740993}}"), "estimated")
	require.NoError(t, err)
	require.Contains(t, string(encoded), "9007199254740993")
	response := &kiro.Response{Blocks: []kiro.ResponseBlock{{Type: "redacted_thinking", Data: "opaque"}}}
	account := customKiroAccount()
	account.Credentials["usage_mode"] = "estimated"
	require.Error(t, resolveCustomKiroUsage(context.Background(), account, []byte("{\"messages\":[{\"role\":\"user\",\"content\":\"hi\"}]}"), "claude-sonnet-4-5-20250929", response))
}

func TestCustomKiroDefaultUsagePolicyRejectsAllProtocols(t *testing.T) {
	for _, format := range []string{"messages", "chat", "responses"} {
		for _, stream := range []bool{false, true} {
			t.Run(format+"/"+map[bool]string{true: "stream", false: "buffered"}[stream], func(t *testing.T) {
				account := customKiroAccount()
				service, _ := customKiroTestService(account, customKiroHTTP{send: func(*http.Request, string, int64, int) (*http.Response, error) {
					return customKiroStreamResponse(customKiroStream(false, false)), nil
				}})
				recorder := httptest.NewRecorder()
				contextGin, _ := gin.CreateTestContext(recorder)
				contextGin.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
				result, err := (&GatewayService{customKiro: service}).forwardCustomKiro(context.Background(), contextGin, account, []byte("{\"messages\":[{\"role\":\"user\",\"content\":\"hi\"}]}"), "claude-sonnet-4-5-20250929", stream, format, true, nil)
				require.ErrorContains(t, err, "精确计量模式")
				require.Equal(t, ClaudeUsage{}, result.Usage)
				require.Equal(t, "verified", recorder.Header().Get("X-Sub2API-Kiro-Usage-Policy"))
				require.NotContains(t, recorder.Body.String(), "\"kiro_usage_source\":\"estimated\"")
				require.NotContains(t, recorder.Body.String(), "event: message_stop")
				require.NotContains(t, recorder.Body.String(), "event: response.completed")
				require.NotContains(t, recorder.Body.String(), "[DONE]")
				if !stream {
					require.Equal(t, http.StatusBadGateway, recorder.Code)
				}
			})
		}
	}
}

func TestCustomKiroConnectionFailureReportsHTTPStatus(t *testing.T) {
	account := customKiroAccount()
	service, _ := customKiroTestService(account, customKiroHTTP{send: func(*http.Request, string, int64, int) (*http.Response, error) {
		return customKiroJSONResponse(403, "private-upstream-body"), nil
	}})
	recorder := httptest.NewRecorder()
	contextGin, _ := gin.CreateTestContext(recorder)
	contextGin.Request = httptest.NewRequest(http.MethodPost, "/test", nil)
	require.Error(t, (&AccountTestService{customKiro: service}).testCustomKiroConnection(contextGin, account, "claude-sonnet-4-5-20250929", "hi"))
	require.Contains(t, recorder.Body.String(), "HTTP 403")
	require.NotContains(t, recorder.Body.String(), "private-upstream-body")
}

func TestCustomKiroUsageModeEditPreservesSecrets(t *testing.T) {
	account := customKiroAccount()
	merged := MergePreservingSensitiveCreds(account.Credentials, map[string]any{"usage_mode": "estimated"})
	normalized, err := kiro.Normalize(account.Type, merged)
	require.NoError(t, err)
	require.Equal(t, "estimated", normalized["usage_mode"])
	require.Equal(t, "refresh-old", normalized["refresh_token"])
	require.Equal(t, "access-old", normalized["access_token"])
}
