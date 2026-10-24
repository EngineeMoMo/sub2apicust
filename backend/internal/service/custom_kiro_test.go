//go:build unit

package service

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"hash/crc32"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type customKiroAccountRepository struct {
	AccountRepository
	mutex        sync.Mutex
	account      *Account
	persistError error
	persistHook  func(*Account)
	rotations    int
}

func (repository *customKiroAccountRepository) GetByID(context.Context, int64) (*Account, error) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()
	snapshot := *repository.account
	snapshot.Credentials = shallowCopyMap(repository.account.Credentials)
	return &snapshot, nil
}
func (repository *customKiroAccountRepository) UpdateKiroOAuthCredentialsIfUnchanged(_ context.Context, _ int64, expected map[string]any, proxyID *int64, credentials map[string]any) (bool, error) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()
	if repository.persistError != nil {
		return false, repository.persistError
	}
	if repository.persistHook != nil {
		repository.persistHook(repository.account)
		repository.persistHook = nil
	}
	if !reflect.DeepEqual(repository.account.Credentials, expected) || !reflect.DeepEqual(repository.account.ProxyID, proxyID) {
		return false, nil
	}
	repository.account.Credentials = shallowCopyMap(credentials)
	repository.rotations++
	return true, nil
}

type customKiroHTTP struct {
	HTTPUpstream
	send func(*http.Request, string, int64, int) (*http.Response, error)
}

func (upstream customKiroHTTP) Do(request *http.Request, proxy string, accountID int64, concurrency int) (*http.Response, error) {
	return upstream.send(request, proxy, accountID, concurrency)
}

type customKiroProxyRepository struct {
	ProxyRepository
	proxy *Proxy
	err   error
}

func (repository customKiroProxyRepository) GetByID(context.Context, int64) (*Proxy, error) {
	return repository.proxy, repository.err
}

func customKiroAccount() *Account {
	return &Account{ID: 88, Platform: PlatformKiro, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Concurrency: 3, Credentials: map[string]any{"access_token": "access-old", "refresh_token": "refresh-old", "profile_arn": "arn:aws:codewhisperer:us-east-1:123:profile/test", "expires_at": time.Now().Add(time.Hour).Unix(), "auth_method": "social", "auth_region": "eu-west-1", "runtime_region": "us-east-1"}}
}
func customKiroJSONResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body))}
}
func customKiroFrame(eventType string, payload any) []byte {
	encoded, _ := json.Marshal(payload)
	headers := []byte{}
	for _, pair := range [][2]string{{":message-type", "event"}, {":event-type", eventType}} {
		headers = append(headers, byte(len(pair[0])))
		headers = append(headers, pair[0]...)
		headers = append(headers, 7, byte(len(pair[1])>>8), byte(len(pair[1])))
		headers = append(headers, pair[1]...)
	}
	frame := make([]byte, 12)
	binary.BigEndian.PutUint32(frame, uint32(16+len(headers)+len(encoded)))
	binary.BigEndian.PutUint32(frame[4:], uint32(len(headers)))
	binary.BigEndian.PutUint32(frame[8:], crc32.ChecksumIEEE(frame[:8]))
	frame = append(frame, headers...)
	frame = append(frame, encoded...)
	checksum := make([]byte, 4)
	binary.BigEndian.PutUint32(checksum, crc32.ChecksumIEEE(frame))
	return append(frame, checksum...)
}
func customKiroStream(includeUsage bool, tool bool) []byte {
	stream := customKiroFrame("assistantResponseEvent", map[string]string{"content": "hello"})
	if tool {
		stream = append(stream, customKiroFrame("toolUseEvent", map[string]any{"toolUseId": "call-1", "name": "lookup", "input": map[string]string{"q": "hi"}, "stop": true})...)
	}
	if includeUsage {
		stream = append(stream, customKiroFrame("metadataEvent", map[string]any{"tokenUsage": map[string]int{"uncachedInputTokens": 13, "cacheReadInputTokens": 3, "cacheWriteInputTokens": 2, "outputTokens": 17}})...)
	}
	return stream
}
func customKiroStreamResponse(body []byte) *http.Response {
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/vnd.amazon.eventstream"}, "X-Amzn-Requestid": []string{"req-kiro"}}, Body: io.NopCloser(bytes.NewReader(body))}
}
func customKiroTestService(account *Account, upstream HTTPUpstream) (*CustomKiroService, *customKiroAccountRepository) {
	repo := &customKiroAccountRepository{account: account}
	api := NewOAuthRefreshAPI(repo, nil)
	return NewCustomKiroService(repo, nil, upstream, api), repo
}

func TestCustomKiroRefreshUsesIDCRegionAndProxy(t *testing.T) {
	account := customKiroAccount()
	proxyID := int64(9)
	account.ProxyID = &proxyID
	account.Credentials["auth_method"] = "idc"
	account.Credentials["client_id"] = "client"
	account.Credentials["client_secret"] = "secret"
	upstream := customKiroHTTP{send: func(request *http.Request, proxy string, accountID int64, concurrency int) (*http.Response, error) {
		require.Equal(t, "https://oidc.eu-west-1.amazonaws.com/token", request.URL.String())
		require.Equal(t, "http://127.0.0.1:18080", proxy)
		require.Equal(t, int64(88), accountID)
		require.Equal(t, 3, concurrency)
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		require.Equal(t, "refresh_token", gjson.GetBytes(body, "grantType").String())
		require.Equal(t, "secret", gjson.GetBytes(body, "clientSecret").String())
		require.Empty(t, request.Header.Get("Authorization"))
		return customKiroJSONResponse(200, `{"accessToken":"new-access","refreshToken":"new-refresh","expiresIn":3600}`), nil
	}}
	service, _ := customKiroTestService(account, upstream)
	service.proxies = customKiroProxyRepository{proxy: &Proxy{ID: proxyID, Protocol: "http", Host: "127.0.0.1", Port: 18080, Status: StatusActive}}
	credentials, err := service.Refresh(context.Background(), account)
	require.NoError(t, err)
	require.Equal(t, "new-refresh", credentials["refresh_token"])
	require.Equal(t, "us-east-1", credentials["runtime_region"])
	require.Equal(t, "secret", credentials["client_secret"])
	require.Equal(t, account.GetCredential("profile_arn"), credentials["profile_arn"])
}

func TestCustomKiroConcurrentRefreshPersistsOnce(t *testing.T) {
	account := customKiroAccount()
	account.Credentials["expires_at"] = time.Now().Add(-time.Hour).Unix()
	var refreshes atomic.Int32
	upstream := customKiroHTTP{send: func(request *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
		refreshes.Add(1)
		return customKiroJSONResponse(200, `{"accessToken":"fresh-access","refreshToken":"rotated-refresh","expiresIn":3600,"profileArn":"profile"}`), nil
	}}
	service, repository := customKiroTestService(account, upstream)
	var workers sync.WaitGroup
	failures := make(chan error, 12)
	for worker := 0; worker < 12; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			_, credentials, err := service.prepare(context.Background(), account, "")
			if err == nil && credentials.AccessToken != "fresh-access" {
				err = errors.New("stale credentials")
			}
			failures <- err
		}()
	}
	workers.Wait()
	close(failures)
	for err := range failures {
		require.NoError(t, err)
	}
	require.Equal(t, int32(1), refreshes.Load())
	require.Equal(t, 1, repository.rotations)
	require.Equal(t, "rotated-refresh", repository.account.GetCredential("refresh_token"))
}

func TestCustomKiroRefreshPersistenceFailureDoesNotReturnTokens(t *testing.T) {
	account := customKiroAccount()
	account.Credentials["expires_at"] = time.Now().Add(-time.Hour).Unix()
	service, repository := customKiroTestService(account, customKiroHTTP{send: func(*http.Request, string, int64, int) (*http.Response, error) {
		return customKiroJSONResponse(200, `{"accessToken":"fresh-access","refreshToken":"rotated-refresh","expiresIn":3600}`), nil
	}})
	repository.persistError = errors.New("database unavailable")
	_, credentials, err := service.prepare(context.Background(), account, "")
	require.Error(t, err)
	require.Empty(t, credentials.AccessToken)
	require.Equal(t, "refresh-old", repository.account.GetCredential("refresh_token"))
	require.Zero(t, repository.rotations)
	var contained *providerCycleContainmentRefreshError
	require.ErrorAs(t, err, &contained)
	require.True(t, isProviderScopedTerminalRefreshError(err))
}

func TestCustomKiroRefreshDoesNotOverwriteConcurrentAdminCredentials(t *testing.T) {
	account := customKiroAccount()
	account.Credentials["expires_at"] = time.Now().Add(-time.Hour).Unix()
	service, repository := customKiroTestService(account, customKiroHTTP{send: func(*http.Request, string, int64, int) (*http.Response, error) {
		return customKiroJSONResponse(200, `{"accessToken":"stale-rotated","refreshToken":"stale-refresh","expiresIn":3600}`), nil
	}})
	repository.persistHook = func(current *Account) {
		current.Credentials = map[string]any{"access_token": "admin-access", "refresh_token": "admin-refresh", "profile_arn": "admin-profile", "expires_at": time.Now().Add(time.Hour).Unix()}
	}
	_, credentials, err := service.prepare(context.Background(), account, "")
	require.NoError(t, err)
	require.Equal(t, "admin-access", credentials.AccessToken)
	require.Zero(t, repository.rotations)
}

func TestCustomKiroRejectedTokenRefreshesOnlyOnce(t *testing.T) {
	account := customKiroAccount()
	var refreshes, calls int
	service, _ := customKiroTestService(account, customKiroHTTP{send: func(request *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
		if strings.Contains(request.URL.Host, "auth.desktop.kiro.dev") {
			refreshes++
			return customKiroJSONResponse(200, `{"accessToken":"fresh-access","refreshToken":"fresh-refresh","expiresIn":3600}`), nil
		}
		calls++
		if calls == 1 {
			return customKiroJSONResponse(401, "{}"), nil
		}
		require.Equal(t, "Bearer fresh-access", request.Header.Get("Authorization"))
		return customKiroStreamResponse(customKiroStream(true, false)), nil
	}})
	call, err := service.call(context.Background(), account, []byte(`{"messages":[{"role":"user","content":"hi"}]}`), "claude-sonnet-4-6")
	require.NoError(t, err)
	defer call.Response.Body.Close()
	require.Equal(t, 1, refreshes)
	require.Equal(t, 2, calls)
}

func TestCustomKiroUnavailableProxyNeverFallsBackToDirect(t *testing.T) {
	account := customKiroAccount()
	proxyID := int64(9)
	account.ProxyID = &proxyID
	var calls int
	service, _ := customKiroTestService(account, customKiroHTTP{send: func(*http.Request, string, int64, int) (*http.Response, error) {
		calls++
		return nil, errors.New("must not send")
	}})
	service.proxies = customKiroProxyRepository{err: errors.New("missing proxy")}
	_, err := service.call(context.Background(), account, []byte(`{"messages":[{"role":"user","content":"hi"}]}`), "claude-sonnet-4-6")
	require.Error(t, err)
	require.Zero(t, calls)
}

func TestCustomKiroGatewayProtocols(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, format := range []string{"messages", "chat", "responses"} {
		for _, stream := range []bool{false, true} {
			t.Run(format+map[bool]string{true: " stream", false: " buffered"}[stream], func(t *testing.T) {
				account := customKiroAccount()
				account.Type = AccountTypeAPIKey
				account.Credentials = map[string]any{"api_key": "kiro-api-key", "runtime_region": "us-east-1"}
				service, _ := customKiroTestService(account, customKiroHTTP{send: func(request *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
					require.Equal(t, "https://runtime.us-east-1.kiro.dev/", request.URL.String())
					require.Equal(t, "Bearer kiro-api-key", request.Header.Get("Authorization"))
					require.Equal(t, "API_KEY", request.Header.Get("TokenType"))
					require.Equal(t, "AmazonCodeWhispererStreamingService.GenerateAssistantResponse", request.Header.Get("X-Amz-Target"))
					body, err := io.ReadAll(request.Body)
					require.NoError(t, err)
					require.Equal(t, "hi", gjson.GetBytes(body, "conversationState.currentMessage.userInputMessage.content").String())
					require.False(t, gjson.GetBytes(body, "profileArn").Exists())
					return customKiroStreamResponse(customKiroStream(true, true)), nil
				}})
				gateway := &GatewayService{customKiro: service}
				recorder := httptest.NewRecorder()
				contextGin, _ := gin.CreateTestContext(recorder)
				contextGin.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
				result, err := gateway.forwardCustomKiro(context.Background(), contextGin, account, []byte(`{"messages":[{"role":"user","content":"hi"}]}`), "claude-sonnet-4-6", stream, format, true, nil)
				require.NoError(t, err)
				require.Equal(t, 200, recorder.Code)
				require.Equal(t, ClaudeUsage{InputTokens: 13, OutputTokens: 17, CacheReadInputTokens: 3, CacheCreationInputTokens: 2}, result.Usage)
				require.Equal(t, "claude-sonnet-4.6", result.UpstreamModel)
				require.Contains(t, recorder.Body.String(), "hello")
				require.Contains(t, recorder.Body.String(), "lookup")
				require.NotContains(t, recorder.Body.String(), "kiro-api-key")
				if stream {
					require.Contains(t, recorder.Header().Get("Content-Type"), "text/event-stream")
					if format == "messages" {
						require.Contains(t, recorder.Body.String(), "event: message_stop")
						require.Contains(t, recorder.Body.String(), "\"input_tokens\":13")
					} else if format == "chat" {
						require.Contains(t, recorder.Body.String(), "[DONE]")
					} else {
						require.Contains(t, recorder.Body.String(), "response.completed")
					}
				}
				if !stream {
					require.True(t, json.Valid(recorder.Body.Bytes()))
					if format == "messages" {
						require.Equal(t, int64(13), gjson.Get(recorder.Body.String(), "usage.input_tokens").Int())
					}
				}
			})
		}
	}
}

func TestCustomKiroGatewayMissingUsageAndTruncation(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(map[bool]string{true: "stream", false: "buffered"}[stream], func(t *testing.T) {
			account := customKiroAccount()
			service, _ := customKiroTestService(account, customKiroHTTP{send: func(*http.Request, string, int64, int) (*http.Response, error) {
				return customKiroStreamResponse(customKiroStream(false, false)), nil
			}})
			recorder := httptest.NewRecorder()
			contextGin, _ := gin.CreateTestContext(recorder)
			contextGin.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
			result, err := (&GatewayService{customKiro: service}).forwardCustomKiro(context.Background(), contextGin, account, []byte(`{"messages":[{"role":"user","content":"hi"}]}`), "claude-sonnet-4-6", stream, "messages", false, nil)
			require.Error(t, err)
			require.Equal(t, ClaudeUsage{}, result.Usage)
			require.NotContains(t, recorder.Body.String(), "event: message_stop")
			if !stream {
				require.Equal(t, http.StatusBadGateway, recorder.Code)
			}
		})
	}
}

func TestCustomKiroCredentialsNeverReachAnthropicRequestBuilder(t *testing.T) {
	for _, accountType := range []string{AccountTypeOAuth, AccountTypeAPIKey} {
		account := customKiroAccount()
		account.Type = accountType
		require.False(t, anthropicMessagesCredentialAllowed(account))
		require.Error(t, checkAnthropicMessagesCredential(account))
	}
}

func TestCustomKiroAdminNormalizationAndSecretMerge(t *testing.T) {
	account, err := buildAccountForCreate(&CreateAccountInput{Platform: PlatformKiro, Type: AccountTypeOAuth, Credentials: map[string]any{"refreshToken": "refresh", "clientId": "client", "clientSecret": "secret"}}, nil)
	require.NoError(t, err)
	require.Equal(t, "secret", account.GetCredential("client_secret"))
	require.Empty(t, account.GetCredential("clientSecret"))
	merged := MergePreservingSensitiveCreds(account.Credentials, map[string]any{"runtime_region": "eu-west-1", "auth_method": "idc", "client_id": "client"})
	require.Equal(t, "secret", merged["client_secret"])
	require.Equal(t, "refresh", merged["refresh_token"])
	for _, accountType := range []string{AccountTypeSetupToken, AccountTypeUpstream, AccountTypeBedrock} {
		_, err := buildAccountForCreate(&CreateAccountInput{Platform: PlatformKiro, Type: accountType}, nil)
		require.Error(t, err)
	}
}

var _ OAuthRefreshExecutor = (*CustomKiroService)(nil)
var _ CustomKiroRefreshRepository = (*customKiroAccountRepository)(nil)
