//go:build unit

package kiro

import (
	"bytes"
	"context"
	"testing"
)

func TestKiroResponseUsageSources(t *testing.T) {
	text := testFrame("assistantResponseEvent", map[string]string{"content": "OK."})
	metering := testFrame("meteringEvent", map[string]any{"inputTokens": 19, "outputTokens": 7, "usage": 0.01})
	cases := []struct {
		name   string
		tail   []byte
		source string
		usage  Usage
	}{
		{"missing", nil, "", Usage{}},
		{"credits only", testFrame("meteringEvent", map[string]any{"usage": 0.1}), "", Usage{}},
		{"zero metering", testFrame("meteringEvent", map[string]int{"inputTokens": 0, "outputTokens": 0}), "", Usage{}},
		{"zero metadata", testFrame("metadataEvent", map[string]any{"tokenUsage": map[string]int{"uncachedInputTokens": 0, "outputTokens": 0}}), "", Usage{}},
		{"metering", metering, "metering", Usage{InputTokens: 19, OutputTokens: 7}},
		{"metering snapshot", append(append([]byte{}, metering...), metering...), "metering", Usage{InputTokens: 19, OutputTokens: 7}},
		{"metadata wins before", append(usageFrame(), metering...), "metadata", Usage{InputTokens: 13, OutputTokens: 17, CacheReadTokens: 3, CacheWriteTokens: 2}},
		{"metadata wins after", append(append([]byte{}, metering...), usageFrame()...), "metadata", Usage{InputTokens: 13, OutputTokens: 17, CacheReadTokens: 3, CacheWriteTokens: 2}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			stream := append(append([]byte{}, text...), test.tail...)
			response, err := DecodeResponse(context.Background(), bytes.NewReader(stream), nil)
			if err != nil {
				t.Fatal(err)
			}
			if response.HasUsage != (test.source != "") || response.UsageSource != test.source || response.Usage != test.usage {
				t.Fatalf("unexpected usage: %+v", response)
			}
			if response.EventCounts["assistantResponseEvent"] != 1 {
				t.Fatal(response.EventCounts)
			}
		})
	}
}

func TestKiroResponseRejectsMalformedMetering(t *testing.T) {
	for _, payload := range []map[string]any{{"inputTokens": -1, "outputTokens": 2}, {"inputTokens": 2, "outputTokens": 1000000001}, {"inputTokens": 1.5, "outputTokens": 2}, {"inputTokens": "2", "outputTokens": 2}} {
		stream := append(testFrame("assistantResponseEvent", map[string]string{"content": "hello"}), testFrame("meteringEvent", payload)...)
		if _, err := DecodeResponse(context.Background(), bytes.NewReader(stream), nil); err == nil {
			t.Fatal("invalid token count accepted")
		}
	}
}

func TestKiroCredentialsUsageModeRequiresOptIn(t *testing.T) {
	for _, mode := range []any{"verified", "estimated", "auto", "", true, 1, nil} {
		_, err := Normalize("apikey", map[string]any{"api_key": "test-key", "usage_mode": mode})
		valid := mode == "verified" || mode == "estimated"
		if (err == nil) != valid {
			t.Fatalf("mode %v: %v", mode, err)
		}
	}
	normalized, err := Normalize("apikey", map[string]any{"api_key": "test-key"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := normalized["usage_mode"]; ok {
		t.Fatal("unexpected automatic opt-in")
	}
}
