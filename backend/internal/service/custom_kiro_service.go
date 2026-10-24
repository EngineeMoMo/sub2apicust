package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/kiro"
	"github.com/google/uuid"
)

type CustomKiroService struct {
	accounts   AccountRepository
	proxies    ProxyRepository
	upstream   HTTPUpstream
	refreshAPI *OAuthRefreshAPI
}

func NewCustomKiroService(accounts AccountRepository, proxies ProxyRepository, upstream HTTPUpstream, refreshAPI *OAuthRefreshAPI) *CustomKiroService {
	return &CustomKiroService{accounts: accounts, proxies: proxies, upstream: upstream, refreshAPI: refreshAPI}
}

func (s *CustomKiroService) CacheKey(account *Account) string {
	return fmt.Sprintf("kiro:oauth:%d", account.ID)
}
func (s *CustomKiroService) CanRefresh(account *Account) bool {
	return account != nil && account.Platform == PlatformKiro && account.Type == AccountTypeOAuth && account.GetCredential("refresh_token") != ""
}
func (s *CustomKiroService) NeedsRefresh(account *Account, window time.Duration) bool {
	if !s.CanRefresh(account) {
		return false
	}
	credentials, err := kiro.Parse(account.Type, account.Credentials)
	return err != nil || credentials.AccessToken == "" || credentials.ProfileARN == "" || credentials.ExpiresAt.IsZero() || time.Until(credentials.ExpiresAt) <= window
}

func (s *CustomKiroService) proxyURL(ctx context.Context, account *Account) (string, error) {
	if account.ProxyID == nil {
		return "", nil
	}
	if s.proxies == nil {
		return "", errors.New("kiro 账号代理仓库不可用")
	}
	proxy, err := s.proxies.GetByID(ctx, *account.ProxyID)
	if err != nil || proxy == nil || !proxy.IsActive() || proxy.IsExpired(time.Now()) {
		return "", errors.New("kiro 账号代理不可用，请检查代理配置")
	}
	return proxy.URL(), nil
}

func (s *CustomKiroService) Refresh(ctx context.Context, account *Account) (map[string]any, error) {
	credentials, err := kiro.Parse(account.Type, account.Credentials)
	if err != nil {
		return nil, err
	}
	if credentials.RefreshToken == "" {
		return nil, errors.New("kiro 缺少 refresh_token")
	}
	proxy, err := s.proxyURL(ctx, account)
	if err != nil {
		return nil, err
	}
	endpoint, payload := credentials.RefreshRequest()
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	refreshCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(refreshCtx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	if s.upstream == nil {
		return nil, errors.New("kiro HTTP 服务不可用")
	}
	response, err := s.upstream.Do(request, proxy, account.ID, account.Concurrency)
	if err != nil {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
		return nil, errors.New("kiro 刷新请求失败，请检查账号代理及认证地区")
	}
	defer func() { _ = response.Body.Close() }()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, (1<<20)+1))
	if err != nil || len(responseBody) > 1<<20 {
		return nil, errors.New("kiro 刷新响应读取失败")
	}
	if response.StatusCode != http.StatusOK {
		var failure struct {
			Code string `json:"error"`
		}
		_ = json.Unmarshal(responseBody, &failure)
		if failure.Code == "invalid_grant" {
			return nil, errors.New("invalid_grant: Kiro 刷新凭据已失效")
		}
		return nil, fmt.Errorf("kiro 刷新失败，HTTP %d", response.StatusCode)
	}
	var tokens struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		ExpiresIn    int64  `json:"expiresIn"`
		ProfileARN   string `json:"profileArn"`
	}
	if json.Unmarshal(responseBody, &tokens) != nil || strings.TrimSpace(tokens.AccessToken) == "" || tokens.ExpiresIn <= 0 || tokens.ExpiresIn > 366*24*3600 {
		return nil, errors.New("kiro 刷新响应缺少有效 token/有效期")
	}
	updated := MergeCredentials(account.Credentials, map[string]any{"access_token": tokens.AccessToken, "expires_at": strconv.FormatInt(time.Now().Unix()+tokens.ExpiresIn, 10)})
	if tokens.RefreshToken != "" {
		updated["refresh_token"] = tokens.RefreshToken
	}
	if tokens.ProfileARN != "" {
		updated["profile_arn"] = tokens.ProfileARN
	}
	return kiro.Normalize(AccountTypeOAuth, updated)
}

type customKiroRejectedTokenRefresher struct {
	*CustomKiroService
	rejected string
}

type customKiroProbeContextKey struct{}

func customKiroAccountEligibility(ctx context.Context, account *Account) error {
	if account == nil || account.Platform != PlatformKiro || (account.Type != AccountTypeOAuth && account.Type != AccountTypeAPIKey) {
		return errors.New("kiro 账号平台或凭据类型已改变")
	}
	probe, _ := ctx.Value(customKiroProbeContextKey{}).(bool)
	if !probe && (!account.IsSchedulable() || (account.ExpiresAt != nil && !account.ExpiresAt.After(time.Now()))) {
		return errors.New("kiro 账号已停用、到期或处于调度冷却期")
	}
	return nil
}

func (s customKiroRejectedTokenRefresher) NeedsRefresh(account *Account, window time.Duration) bool {
	return account.GetCredential("access_token") == s.rejected || s.CustomKiroService.NeedsRefresh(account, window)
}

func (s *CustomKiroService) prepare(ctx context.Context, account *Account, rejected string) (*Account, kiro.Credentials, error) {
	if s == nil || s.accounts == nil || account == nil {
		return nil, kiro.Credentials{}, errors.New("kiro 服务未配置")
	}
	for attempt := 0; attempt < 40; attempt++ {
		latest, err := s.accounts.GetByID(ctx, account.ID)
		if err != nil || latest == nil {
			return nil, kiro.Credentials{}, errors.New("kiro 账号读取失败")
		}
		if err := customKiroAccountEligibility(ctx, latest); err != nil {
			return nil, kiro.Credentials{}, err
		}
		credentials, err := kiro.Parse(latest.Type, latest.Credentials)
		if err != nil {
			return nil, kiro.Credentials{}, err
		}
		if latest.Type == AccountTypeAPIKey {
			return latest, credentials, nil
		}
		needsRefresh := s.NeedsRefresh(latest, time.Minute) || (rejected != "" && credentials.AccessToken == rejected)
		if needsRefresh && s.CanRefresh(latest) {
			if s.refreshAPI == nil {
				return nil, kiro.Credentials{}, errors.New("kiro 统一刷新服务未配置")
			}
			var executor OAuthRefreshExecutor = s
			if rejected != "" {
				executor = customKiroRejectedTokenRefresher{CustomKiroService: s, rejected: rejected}
			}
			result, err := s.refreshAPI.RefreshIfNeeded(withOAuthRefreshRequestPath(ctx), latest, executor, time.Minute)
			if err != nil {
				return nil, kiro.Credentials{}, err
			}
			if result == nil {
				return nil, kiro.Credentials{}, errors.New("kiro 刷新未返回账号")
			}
			if result.LockHeld {
				timer := time.NewTimer(250 * time.Millisecond)
				select {
				case <-ctx.Done():
					timer.Stop()
					return nil, kiro.Credentials{}, ctx.Err()
				case <-timer.C:
				}
				continue
			}
			if result.Account != nil {
				latest = result.Account
				credentials, err = kiro.Parse(latest.Type, latest.Credentials)
				if err != nil {
					return nil, kiro.Credentials{}, err
				}
			}
		}
		if latest.Type != AccountTypeOAuth {
			return nil, kiro.Credentials{}, errors.New("kiro 账号在刷新期间发生变化")
		}
		if err := customKiroAccountEligibility(ctx, latest); err != nil {
			return nil, kiro.Credentials{}, err
		}
		if credentials.AccessToken == "" || credentials.ExpiresAt.IsZero() || time.Until(credentials.ExpiresAt) <= 0 || credentials.ProfileARN == "" || (rejected != "" && credentials.AccessToken == rejected) {
			return nil, kiro.Credentials{}, errors.New("kiro 访问凭据无效，请重新导入包含 profile_arn 的凭据")
		}
		return latest, credentials, nil
	}
	return nil, kiro.Credentials{}, errors.New("kiro 正在刷新，请稍后重试")
}

type customKiroRequestError struct{ err error }

func (err *customKiroRequestError) Error() string { return err.err.Error() }
func (err *customKiroRequestError) Unwrap() error { return err.err }

type customKiroCall struct {
	Response *http.Response
	Account  *Account
	Model    string
}

func (s *CustomKiroService) call(ctx context.Context, account *Account, body []byte, model string) (*customKiroCall, error) {
	conversationID := uuid.NewString()
	if _, err := kiro.BuildPayload(body, account.GetMappedModel(model), "", conversationID); err != nil {
		return nil, &customKiroRequestError{err: err}
	}
	latest, credentials, err := s.prepare(ctx, account, "")
	if err != nil {
		return nil, err
	}
	for attempt := 0; attempt < 2; attempt++ {
		mappedModel := latest.GetMappedModel(model)
		payload, err := kiro.BuildPayload(body, mappedModel, credentials.ProfileARN, conversationID)
		if err != nil {
			return nil, &customKiroRequestError{err: err}
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		proxy, err := s.proxyURL(ctx, latest)
		if err != nil {
			return nil, err
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, credentials.Endpoint(), bytes.NewReader(encoded))
		if err != nil {
			return nil, err
		}
		token := credentials.AccessToken
		if latest.Type == AccountTypeAPIKey {
			token = credentials.APIKey
			request.Header.Set("TokenType", "API_KEY")
		}
		request.Header.Set("Authorization", "Bearer "+token)
		request.Header.Set("Content-Type", "application/x-amz-json-1.0")
		request.Header.Set("Accept", "application/vnd.amazon.eventstream")
		request.Header.Set("X-Amz-Target", "AmazonCodeWhispererStreamingService.GenerateAssistantResponse")
		request.Header.Set("User-Agent", "aws-sdk-rust/1.3.15 ua/2.1 api/codewhispererstreaming/0.1.17593 os/linux lang/rust/1.92.0 md/appVersion-2.10.0 app/AmazonQ-For-CLI")
		request.Header.Set("x-amz-user-agent", "aws-sdk-rust/1.3.15 app/AmazonQ-For-CLI")
		request.Header.Set("x-amzn-codewhisperer-optout", "true")
		request.Header.Set("amz-sdk-invocation-id", conversationID)
		request.Header.Set("amz-sdk-request", fmt.Sprintf("attempt=%d; max=2", attempt+1))
		if s.upstream == nil {
			return nil, errors.New("kiro HTTP 服务未配置")
		}
		response, err := s.upstream.Do(request, proxy, latest.ID, latest.Concurrency)
		if err != nil {
			if response != nil && response.Body != nil {
				_ = response.Body.Close()
			}
			return nil, errors.New("kiro 上游连接失败，请检查账号代理及运行地区")
		}
		if attempt == 0 && (response.StatusCode == 401 || response.StatusCode == 403) && s.CanRefresh(latest) {
			_ = response.Body.Close()
			latest, credentials, err = s.prepare(ctx, latest, credentials.AccessToken)
			if err != nil {
				return nil, err
			}
			continue
		}
		nativeModel, _ := kiro.ResolveModel(mappedModel)
		return &customKiroCall{Response: response, Account: latest, Model: nativeModel}, nil
	}
	return nil, errors.New("kiro 重试耗尽")
}
