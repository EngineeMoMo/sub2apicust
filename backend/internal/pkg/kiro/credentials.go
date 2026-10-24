package kiro

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const DefaultRegion = "us-east-1"

var regionPattern = regexp.MustCompile("^[a-z]{2}(-[a-z0-9]+)+-[0-9]+$")

type Credentials struct {
	AuthMethod    string
	AuthRegion    string
	RuntimeRegion string
	ProfileARN    string
	AccessToken   string
	RefreshToken  string
	ClientID      string
	ClientSecret  string
	APIKey        string
	ExpiresAt     time.Time
}

func CanonicalFields(values map[string]any) (map[string]any, error) {
	out := make(map[string]any, len(values)+3)
	for key, value := range values {
		out[key] = value
	}
	aliases := map[string][]string{
		"access_token": {"accessToken"}, "refresh_token": {"refreshToken"},
		"client_id": {"clientId"}, "client_secret": {"clientSecret"},
		"profile_arn": {"profileArn", "profileARN"}, "expires_at": {"expiresAt"},
		"auth_method": {"authMethod"}, "auth_region": {"sso_region", "ssoRegion", "region"},
		"runtime_region": {"api_region", "apiRegion"},
	}
	for key, names := range aliases {
		for _, name := range names {
			if value, exists := out[name]; exists {
				if existing, present := out[key]; present && fmt.Sprint(existing) != fmt.Sprint(value) {
					return nil, fmt.Errorf("kiro 字段 %s 与 %s 冲突", key, name)
				}
				out[key] = value
				delete(out, name)
			}
		}
	}
	return out, nil
}

func Normalize(accountType string, values map[string]any) (map[string]any, error) {
	if accountType != "oauth" && accountType != "apikey" {
		return nil, fmt.Errorf("kiro 仅支持 OAuth 凭据导入或 API Key")
	}
	out, err := CanonicalFields(values)
	if err != nil {
		return nil, err
	}
	if value, exists := out["usage_mode"]; exists {
		mode, valid := value.(string)
		if !valid || (mode != "verified" && mode != "estimated") {
			return nil, fmt.Errorf("kiro usage_mode 仅支持 verified 或 estimated；estimated 会按估算 Token 计费")
		}
	}
	for _, key := range []string{"access_token", "refresh_token", "api_key", "client_id", "client_secret", "profile_arn", "auth_region", "runtime_region", "auth_method"} {
		if input, exists := out[key]; exists {
			value, ok := input.(string)
			if !ok {
				return nil, fmt.Errorf("kiro 字段 %s 必须为字符串", key)
			}
			value = strings.TrimSpace(value)
			if len(value) > 65536 || strings.ContainsAny(value, "\r\n") {
				return nil, fmt.Errorf("kiro 字段 %s 格式无效", key)
			}
			out[key] = value
		}
	}
	if stringValue(out, "auth_region") == "" {
		out["auth_region"] = DefaultRegion
	}
	if stringValue(out, "runtime_region") == "" {
		out["runtime_region"] = DefaultRegion
	}
	for _, key := range []string{"auth_region", "runtime_region"} {
		if !regionPattern.MatchString(stringValue(out, key)) {
			return nil, fmt.Errorf("kiro 字段 %s 不是有效 AWS 地区", key)
		}
	}
	method := strings.ToLower(stringValue(out, "auth_method"))
	switch method {
	case "", "social":
		if method == "" && (stringValue(out, "client_id") != "" || stringValue(out, "client_secret") != "") {
			method = "idc"
		} else {
			method = "social"
		}
	case "idc", "oidc", "builderid", "builder-id", "iam":
		method = "idc"
	default:
		return nil, fmt.Errorf("kiro auth_method 必须为 social 或 idc")
	}
	out["auth_method"] = method
	if input, exists := out["expires_at"]; exists && input != nil && fmt.Sprint(input) != "" {
		expires, err := parseExpiry(input)
		if err != nil {
			return nil, err
		}
		out["expires_at"] = strconv.FormatInt(expires.Unix(), 10)
	}
	if accountType == "apikey" {
		if stringValue(out, "api_key") == "" {
			return nil, fmt.Errorf("kiro API Key 不能为空")
		}
		for _, key := range []string{"access_token", "refresh_token", "client_id", "client_secret"} {
			delete(out, key)
		}
	} else {
		delete(out, "api_key")
		if stringValue(out, "refresh_token") == "" && stringValue(out, "access_token") == "" {
			return nil, fmt.Errorf("请导入 Kiro refresh_token 或尚未过期的 access_token")
		}
		if stringValue(out, "refresh_token") == "" && (stringValue(out, "expires_at") == "" || stringValue(out, "profile_arn") == "") {
			return nil, fmt.Errorf("仅导入 access_token 时必须提供 expires_at 和 profile_arn")
		}
		if method == "idc" && stringValue(out, "refresh_token") != "" && (stringValue(out, "client_id") == "" || stringValue(out, "client_secret") == "") {
			return nil, fmt.Errorf("IDC 刷新需要 client_id 和 client_secret")
		}
	}
	return out, nil
}

func Parse(accountType string, values map[string]any) (Credentials, error) {
	normalized, err := Normalize(accountType, values)
	if err != nil {
		return Credentials{}, err
	}
	credentials := Credentials{
		AuthMethod: stringValue(normalized, "auth_method"), AuthRegion: stringValue(normalized, "auth_region"),
		RuntimeRegion: stringValue(normalized, "runtime_region"), ProfileARN: stringValue(normalized, "profile_arn"),
		AccessToken: stringValue(normalized, "access_token"), RefreshToken: stringValue(normalized, "refresh_token"),
		ClientID: stringValue(normalized, "client_id"), ClientSecret: stringValue(normalized, "client_secret"), APIKey: stringValue(normalized, "api_key"),
	}
	if input, exists := normalized["expires_at"]; exists {
		credentials.ExpiresAt, _ = parseExpiry(input)
	}
	return credentials, nil
}

func (credentials Credentials) Endpoint() string {
	return "https://runtime." + credentials.RuntimeRegion + ".kiro.dev/"
}

func (credentials Credentials) RefreshRequest() (string, map[string]string) {
	if credentials.AuthMethod == "idc" {
		return "https://oidc." + credentials.AuthRegion + ".amazonaws.com/token", map[string]string{"grantType": "refresh_token", "clientId": credentials.ClientID, "clientSecret": credentials.ClientSecret, "refreshToken": credentials.RefreshToken}
	}
	return "https://prod." + credentials.AuthRegion + ".auth.desktop.kiro.dev/refreshToken", map[string]string{"refreshToken": credentials.RefreshToken}
}

func stringValue(values map[string]any, key string) string {
	value, _ := values[key].(string)
	return value
}

func parseExpiry(input any) (time.Time, error) {
	value := fmt.Sprint(input)
	if number, ok := input.(float64); ok && number > 0 && number < 1e15 && number == float64(int64(number)) {
		value = strconv.FormatInt(int64(number), 10)
	}
	if timestamp, err := strconv.ParseInt(value, 10, 64); err == nil && timestamp > 0 {
		if timestamp > 100000000000 {
			return time.UnixMilli(timestamp), nil
		}
		return time.Unix(timestamp, 0), nil
	}
	if timestamp, err := time.Parse(time.RFC3339, value); err == nil {
		return timestamp, nil
	}
	return time.Time{}, fmt.Errorf("kiro expires_at 必须为 Unix 秒/毫秒或 RFC3339 时间")
}
