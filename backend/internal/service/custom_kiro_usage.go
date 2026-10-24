package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/kiro"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/tiktoken-go/tokenizer"
	"go.uber.org/zap"
)

func resolveCustomKiroUsage(ctx context.Context, account *Account, body []byte, model string, response *kiro.Response) error {
	if response == nil {
		return errors.New("Kiro 响应内容为空")
	}
	if response.HasUsage {
		return nil
	}
	if account == nil || account.GetCredential("usage_mode") != "estimated" {
		return errors.New("Kiro 已生成内容但未提供可核实的 Token 用量；当前为精确计量模式。只有显式启用估算计费后才能按估算用量完成请求")
	}
	var request map[string]any
	if err := json.Unmarshal(body, &request); err != nil {
		return errors.New("Kiro 无法估算请求用量")
	}
	request["model"] = model
	normalized, err := json.Marshal(request)
	if err != nil {
		return err
	}
	inputTokens, err := estimateAnthropicCountTokensLocally(normalized)
	if err != nil {
		return errors.New("Kiro 输入 Token 估算失败")
	}
	codec, err := tokenizer.Get(tokenizer.Cl100kBase)
	if err != nil {
		return errors.New("Kiro 输出 Token 估算器不可用")
	}
	var output strings.Builder
	for _, block := range response.Blocks {
		switch block.Type {
		case "text":
			output.WriteString(block.Text)
		case "thinking":
			output.WriteString(block.Thinking)
		case "tool_use":
			output.WriteString(block.Name)
			output.Write(block.Input)
		case "redacted_thinking":
			return errors.New("Kiro 无法估算已隐藏的思考用量，请使用精确计量模式")
		}
		output.WriteByte('\n')
	}
	outputTokens, err := codec.Count(output.String())
	if err != nil {
		return errors.New("Kiro 输出 Token 估算失败")
	}
	response.Usage = kiro.Usage{InputTokens: max(1, inputTokens), OutputTokens: max(1, outputTokens)}
	response.HasUsage = true
	response.UsageSource = "estimated"
	logger.FromContext(ctx).Warn("kiro.estimated_usage", zap.Int64("account_id", account.ID), zap.String("model", model), zap.Int("input_tokens", response.Usage.InputTokens), zap.Int("output_tokens", response.Usage.OutputTokens))
	return nil
}

func customKiroAnnotateUsage(encoded []byte, source string) ([]byte, error) {
	if source == "" {
		return encoded, nil
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &payload); err != nil {
		return nil, err
	}
	encodedSource, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}
	payload["kiro_usage_source"] = encodedSource
	return json.Marshal(payload)
}

func customKiroAnnotateUsageSSE(text, source string) (string, error) {
	if source == "" {
		return text, nil
	}
	payload := strings.TrimSpace(strings.TrimPrefix(text, "data:"))
	if payload == "[DONE]" {
		return text, nil
	}
	encoded, err := customKiroAnnotateUsage([]byte(payload), source)
	if err != nil {
		return "", err
	}
	return "data: " + string(encoded) + "\n\n", nil
}
