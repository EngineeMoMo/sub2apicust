//go:build unit

package kiro

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"hash/crc32"
	"strings"
	"testing"
)

func testFrame(eventType string, payload any) []byte {
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
func usageFrame() []byte {
	return testFrame("metadataEvent", map[string]any{"tokenUsage": map[string]int{"uncachedInputTokens": 13, "cacheReadInputTokens": 3, "cacheWriteInputTokens": 2, "outputTokens": 17}})
}

func TestKiroCredentials(t *testing.T) {
	cases := []struct {
		name, accountType string
		credentials       map[string]any
		wantError         bool
	}{
		{"social refresh", "oauth", map[string]any{"refreshToken": "refresh", "region": "eu-west-1"}, false},
		{"idc refresh", "oauth", map[string]any{"refresh_token": "refresh", "auth_method": "idc", "client_id": "client", "clientSecret": "secret"}, false},
		{"static access", "oauth", map[string]any{"access_token": "access", "profile_arn": "arn:aws:codewhisperer:us-east-1:123:profile/id", "expiresAt": "2026-12-01T00:00:00Z"}, false},
		{"api key", "apikey", map[string]any{"api_key": "kiro-key"}, false},
		{"missing token", "oauth", map[string]any{}, true},
		{"missing idc secret", "oauth", map[string]any{"refresh_token": "refresh", "auth_method": "idc", "client_id": "client"}, true},
		{"missing access expiry", "oauth", map[string]any{"access_token": "access", "profile_arn": "profile"}, true},
		{"region injection", "oauth", map[string]any{"refresh_token": "refresh", "runtime_region": "us-east-1.evil.test/path"}, true},
		{"bad field type", "oauth", map[string]any{"refresh_token": 123}, true},
		{"conflicting alias", "oauth", map[string]any{"refresh_token": "first", "refreshToken": "second"}, true},
		{"header injection", "apikey", map[string]any{"api_key": "first\r\nX-Header: second"}, true},
		{"unsupported type", "setup-token", map[string]any{"refresh_token": "refresh"}, true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			normalized, err := Normalize(test.accountType, test.credentials)
			if (err != nil) != test.wantError {
				t.Fatalf("error=%v", err)
			}
			if err == nil {
				if _, err := Parse(test.accountType, normalized); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
	credentials, err := Parse("oauth", map[string]any{"refreshToken": "refresh", "region": "eu-west-1"})
	if err != nil {
		t.Fatal(err)
	}
	if credentials.Endpoint() != "https://runtime.us-east-1.kiro.dev/" {
		t.Fatal(credentials.Endpoint())
	}
	endpoint, payload := credentials.RefreshRequest()
	if endpoint != "https://prod.eu-west-1.auth.desktop.kiro.dev/refreshToken" || payload["refreshToken"] != "refresh" {
		t.Fatal(endpoint, payload)
	}
	idc, _ := Parse("oauth", map[string]any{"refresh_token": "refresh", "client_id": "client", "client_secret": "secret", "auth_region": "ap-northeast-1"})
	endpoint, payload = idc.RefreshRequest()
	if endpoint != "https://oidc.ap-northeast-1.amazonaws.com/token" || payload["grantType"] != "refresh_token" {
		t.Fatal(endpoint)
	}
}

func TestKiroExpiryFormats(t *testing.T) {
	for _, input := range []any{int64(1796083200), float64(1796083200000), "1796083200", "2026-12-01T00:00:00Z"} {
		expiry, err := parseExpiry(input)
		if err != nil || expiry.Unix() != 1796083200 {
			t.Fatalf("%v => %v %v", input, expiry, err)
		}
	}
	for _, input := range []any{-1, "never", 1.5} {
		if _, err := parseExpiry(input); err == nil {
			t.Fatalf("accepted %v", input)
		}
	}
}

func TestKiroFramesRejectCorruption(t *testing.T) {
	valid := testFrame("assistantResponseEvent", map[string]string{"content": "hello"})
	cases := map[string][]byte{"valid": valid, "truncated prelude": valid[:7], "truncated body": valid[:len(valid)-2], "bad prelude CRC": append([]byte{}, valid...), "bad message CRC": append([]byte{}, valid...)}
	cases["bad prelude CRC"][8] ^= 1
	cases["bad message CRC"][len(valid)-1] ^= 1
	for name, frame := range cases {
		t.Run(name, func(t *testing.T) {
			events := 0
			err := ReadEvents(context.Background(), bytes.NewReader(frame), func(event Event) error { events++; return nil })
			if name == "valid" {
				if err != nil || events != 1 {
					t.Fatal(err, events)
				}
			} else if err == nil {
				t.Fatal("accepted corrupt frame")
			}
		})
	}
	oversized := append([]byte{}, valid...)
	binary.BigEndian.PutUint32(oversized, uint32(maxFrameSize+1))
	binary.BigEndian.PutUint32(oversized[8:], crc32.ChecksumIEEE(oversized[:8]))
	if err := ReadEvents(context.Background(), bytes.NewReader(oversized), func(Event) error { return nil }); err == nil {
		t.Fatal("oversized frame accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := ReadEvents(ctx, bytes.NewReader(valid), func(Event) error { return nil }); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := decodeHeaders([]byte{20, 1}); err == nil {
		t.Fatal("malformed headers accepted")
	}
}

func TestKiroResponseKeepsExclusiveUsageSnapshots(t *testing.T) {
	stream := append(testFrame("assistantResponseEvent", map[string]string{"content": "hello"}), testFrame("meteringEvent", map[string]any{"usage": 99, "inputTokens": 9999, "outputTokens": 8888})...)
	stream = append(stream, usageFrame()...)
	stream = append(stream, usageFrame()...)
	count := 0
	response, err := DecodeResponse(context.Background(), bytes.NewReader(stream), func(block ResponseBlock) error { count++; return nil })
	if err != nil {
		t.Fatal(err)
	}
	if response.Usage != (Usage{InputTokens: 13, CacheReadTokens: 3, CacheWriteTokens: 2, OutputTokens: 17}) {
		t.Fatalf("double counted or estimated: %+v", response.Usage)
	}
	if count != 1 || response.Blocks[0].Text != "hello" {
		t.Fatal(response.Blocks)
	}
}

func TestKiroResponseRejectsInvalidUsage(t *testing.T) {
	for _, tail := range [][]byte{testFrame("metadataEvent", map[string]any{"tokenUsage": map[string]int{"outputTokens": 2}}), testFrame("metadataEvent", map[string]any{"tokenUsage": map[string]int{"uncachedInputTokens": -1, "outputTokens": 2}})} {
		stream := append(testFrame("assistantResponseEvent", map[string]string{"content": "hello"}), tail...)
		if _, err := DecodeResponse(context.Background(), bytes.NewReader(stream), nil); err == nil {
			t.Fatal("invalid usage accepted")
		}
	}
}

func TestKiroToolFragmentsAndReasoning(t *testing.T) {
	stream := testFrame("reasoningContentEvent", map[string]string{"text": "reason", "signature": "sig"})
	for _, payload := range []map[string]any{{"toolUseId": "tool-1", "name": "lookup", "input": "{\"q\":"}, {"input": "\"hi\"}"}, {"stop": true}} {
		stream = append(stream, testFrame("toolUseEvent", payload)...)
	}
	stream = append(stream, usageFrame()...)
	response, err := DecodeResponse(context.Background(), bytes.NewReader(stream), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Blocks) != 2 || response.Blocks[0].Thinking != "reason" || response.Blocks[0].Signature != "sig" || response.Blocks[1].ID != "tool-1" || string(response.Blocks[1].Input) != `{"q":"hi"}` || response.StopReason != "tool_use" {
		t.Fatalf("%+v", response)
	}
	for _, tool := range []map[string]any{{"toolUseId": "t", "name": "lookup", "input": "{bad"}, {"name": "lookup", "input": map[string]any{}}, {"toolUseId": "t", "input": map[string]any{}}} {
		if _, err := DecodeResponse(context.Background(), bytes.NewReader(append(testFrame("toolUseEvent", tool), usageFrame()...)), nil); err == nil {
			t.Fatal("malformed tool accepted")
		}
	}
}

func TestKiroPayloadToolRoundTripAndImages(t *testing.T) {
	body := []byte(`{"model":"claude-sonnet-4-6","system":[{"type":"text","text":"system"}],"tools":[{"name":"lookup","description":"search","input_schema":{"type":"object"}}],"messages":[{"role":"user","content":[{"type":"text","text":"question"},{"type":"image","source":{"type":"base64","media_type":"image/png","data":"aGk="}}]},{"role":"assistant","content":[{"type":"tool_use","id":"call-1","name":"lookup","input":{"q":"hello"}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"call-1","is_error":true,"content":[{"type":"text","text":"failed"}]}]}]}`)
	payload, err := BuildPayload(body, "claude-sonnet-4-6", "profile", "conversation")
	if err != nil {
		t.Fatal(err)
	}
	history := payload.ConversationState.History
	if len(history) != 2 || history[0].User.Content != "system\n\nquestion" || len(history[0].User.Images) != 1 || history[1].Assistant.Tools[0].ID != "call-1" {
		t.Fatalf("%+v", payload)
	}
	current := payload.ConversationState.Current["userInputMessage"]
	if current.Model != "claude-sonnet-4.6" || current.Context.Results[0].Status != "error" || current.Context.Results[0].Content[0]["text"] != "failed" || len(current.Context.Tools) != 1 {
		t.Fatalf("%+v", current)
	}
	if payload.ProfileARN != "profile" || payload.ConversationState.ID != "conversation" {
		t.Fatal(payload)
	}
}

func TestKiroPayloadRejectsUnsupportedContent(t *testing.T) {
	for _, body := range []string{
		`{"messages":[]}`,
		`{"messages":[{"role":"assistant","content":"prefill"}]}`,
		`{"messages":[{"role":"user","content":[{"type":"document","text":"lost"}]}]}`,
		`{"messages":[{"role":"user","content":[{"type":"tool_result","tool_use_id":"orphan","content":"result"}]}]}`,
		`{"messages":[{"role":"user","content":"hello"}],"tools":[{"type":"web_search_20250305","name":"web_search"}]}`,
		`{"messages":[{"role":"user","content":"hello"}],"tool_choice":{"type":"any"}}`,
		`{"messages":[{"role":"user","content":"hello"}],"stop_sequences":["stop"]}`,
	} {
		if _, err := BuildPayload([]byte(body), "claude-sonnet-4-6", "", "conversation"); err == nil {
			t.Fatal(body)
		}
	}
}

func TestKiroModelsNeverFallbackSilently(t *testing.T) {
	for requested, want := range map[string]string{"claude-sonnet-4-5-20250929": "claude-sonnet-4.5", "claude-sonnet-4-6[1M]": "claude-sonnet-4.6-1m", "custom-model": "custom-model"} {
		actual, err := ResolveModel(requested)
		if err != nil || actual != want {
			t.Fatal(actual, err)
		}
	}
	if _, err := ResolveModel("https://evil.test"); err == nil {
		t.Fatal("invalid model accepted")
	}
	if len(DefaultModels()) == 0 || strings.TrimSpace(DefaultModels()[0]) == "" {
		t.Fatal("no default models")
	}
}
