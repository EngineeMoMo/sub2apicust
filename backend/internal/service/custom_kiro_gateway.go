package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/Wei-Shaw/sub2api/internal/pkg/kiro"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

type customKiroOutput struct {
	context      *gin.Context
	format       string
	model        string
	messageID    string
	started      bool
	disconnected bool
	blockType    string
	blockIndex   int
	responses    *apicompat.AnthropicEventToResponsesState
	chat         *apicompat.ResponsesEventToChatState
	restorer     *apicompat.ResponsesClientToolStreamRestorer
	usageSource  string
}

func validateCustomKiroIngress(body []byte, format string) error {
	if format == "responses" && (gjson.GetBytes(body, "previous_response_id").String() != "" || gjson.GetBytes(body, "conversation").Exists() || gjson.GetBytes(body, "background").Bool() || gjson.GetBytes(body, "store").Bool()) {
		return errors.New("Kiro 仅支持无状态 Responses，请提交完整历史并关闭 store/background")
	}
	if format == "responses" {
		if responseFormat := gjson.GetBytes(body, "text.format.type").String(); responseFormat != "" && responseFormat != "text" {
			return errors.New("Kiro 尚不支持结构化输出约束")
		}
	}
	if format == "chat" {
		if gjson.GetBytes(body, "n").Int() > 1 {
			return errors.New("Kiro 每次仅支持一个候选回复")
		}
		if responseFormat := gjson.GetBytes(body, "response_format.type").String(); responseFormat != "" && responseFormat != "text" {
			return errors.New("Kiro 尚不支持结构化输出约束")
		}
	}
	return nil
}

func newCustomKiroOutput(c *gin.Context, format, model string, includeUsage bool, mapping *apicompat.ResponsesClientToolMapping) *customKiroOutput {
	output := &customKiroOutput{context: c, format: format, model: model, messageID: "msg_" + strings.ReplaceAll(uuid.NewString(), "-", ""), blockIndex: -1}
	if format != "messages" {
		output.responses = apicompat.NewAnthropicEventToResponsesState()
		output.responses.Model = model
	}
	if format == "chat" {
		output.chat = apicompat.NewResponsesEventToChatState()
		output.chat.Model = model
		output.chat.IncludeUsage = includeUsage
	}
	if mapping != nil {
		output.restorer = apicompat.NewResponsesClientToolStreamRestorer(*mapping)
	}
	return output
}

func (output *customKiroOutput) write(value string) {
	if output.disconnected {
		return
	}
	if _, err := io.WriteString(output.context.Writer, value); err != nil {
		output.disconnected = true
		return
	}
	output.context.Writer.Flush()
}

func (output *customKiroOutput) sendResponse(event apicompat.ResponsesStreamEvent) error {
	encoded, err := json.Marshal(event)
	if err != nil {
		return err
	}
	encoded, err = customKiroAnnotateUsage(encoded, output.usageSource)
	if err != nil {
		return err
	}
	if output.format == "chat" {
		for _, chunk := range apicompat.ResponsesEventToChatChunks(&event, output.chat) {
			text, err := apicompat.ChatChunkToSSE(chunk)
			if err != nil {
				return err
			}
			text, err = customKiroAnnotateUsageSSE(text, output.usageSource)
			if err != nil {
				return err
			}
			output.write(text)
		}
		return nil
	}
	payloads := [][]byte{encoded}
	if output.restorer != nil {
		payloads, _, err = output.restorer.RestoreEvent(encoded)
		if err != nil {
			return err
		}
	}
	for _, payload := range payloads {
		output.write("event: " + gjson.GetBytes(payload, "type").String() + "\ndata: " + string(payload) + "\n\n")
	}
	return nil
}

func (output *customKiroOutput) event(name string, payload map[string]any) error {
	payload["type"] = name
	if output.usageSource != "" {
		payload["kiro_usage_source"] = output.usageSource
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if !output.started {
		output.started = true
		output.context.Header("Content-Type", "text/event-stream")
		output.context.Header("Cache-Control", "no-cache")
		output.context.Header("X-Accel-Buffering", "no")
		MarkResponseCommitted(output.context)
	}
	if output.format == "messages" {
		output.write("event: " + name + "\ndata: " + string(encoded) + "\n\n")
		return nil
	}
	var event apicompat.AnthropicStreamEvent
	if err := json.Unmarshal(encoded, &event); err != nil {
		return err
	}
	for _, responseEvent := range apicompat.AnthropicEventToResponsesEvents(&event, output.responses) {
		if err := output.sendResponse(responseEvent); err != nil {
			return err
		}
	}
	return nil
}

func (output *customKiroOutput) closeBlock() error {
	if output.blockType == "" {
		return nil
	}
	output.blockType = ""
	return output.event("content_block_stop", map[string]any{"index": output.blockIndex})
}

func (output *customKiroOutput) consume(block kiro.ResponseBlock) error {
	if !output.started {
		if err := output.event("message_start", map[string]any{"message": map[string]any{"id": output.messageID, "type": "message", "role": "assistant", "model": output.model, "content": []any{}, "stop_reason": nil, "stop_sequence": nil, "usage": kiro.Usage{}}}); err != nil {
			return err
		}
	}
	if block.Type == "tool_use" || block.Type == "redacted_thinking" {
		if err := output.closeBlock(); err != nil {
			return err
		}
		output.blockIndex++
		content := map[string]any{"type": block.Type}
		if block.Type == "tool_use" {
			content["id"] = block.ID
			content["name"] = block.Name
			content["input"] = map[string]any{}
		} else {
			content["data"] = block.Data
		}
		if err := output.event("content_block_start", map[string]any{"index": output.blockIndex, "content_block": content}); err != nil {
			return err
		}
		if block.Type == "tool_use" {
			if err := output.event("content_block_delta", map[string]any{"index": output.blockIndex, "delta": map[string]any{"type": "input_json_delta", "partial_json": string(block.Input)}}); err != nil {
				return err
			}
		}
		return output.event("content_block_stop", map[string]any{"index": output.blockIndex})
	}
	if output.blockType != block.Type {
		if err := output.closeBlock(); err != nil {
			return err
		}
		output.blockIndex++
		output.blockType = block.Type
		content := map[string]any{"type": block.Type}
		if block.Type == "text" {
			content["text"] = ""
		} else {
			content["thinking"] = ""
		}
		if err := output.event("content_block_start", map[string]any{"index": output.blockIndex, "content_block": content}); err != nil {
			return err
		}
	}
	if block.Type == "text" {
		return output.event("content_block_delta", map[string]any{"index": output.blockIndex, "delta": map[string]any{"type": "text_delta", "text": block.Text}})
	}
	if block.Thinking != "" {
		if err := output.event("content_block_delta", map[string]any{"index": output.blockIndex, "delta": map[string]any{"type": "thinking_delta", "thinking": block.Thinking}}); err != nil {
			return err
		}
	}
	if block.Signature != "" {
		return output.event("content_block_delta", map[string]any{"index": output.blockIndex, "delta": map[string]any{"type": "signature_delta", "signature": block.Signature}})
	}
	return nil
}

func (output *customKiroOutput) finish(response *kiro.Response) error {
	output.usageSource = response.UsageSource
	if err := output.closeBlock(); err != nil {
		return err
	}
	if err := output.event("message_delta", map[string]any{"delta": map[string]any{"stop_reason": response.StopReason, "stop_sequence": nil}, "usage": response.Usage}); err != nil {
		return err
	}
	if err := output.event("message_stop", map[string]any{}); err != nil {
		return err
	}
	if output.responses != nil {
		for _, event := range apicompat.FinalizeAnthropicResponsesStream(output.responses) {
			if err := output.sendResponse(event); err != nil {
				return err
			}
		}
	}
	if output.chat != nil {
		for _, chunk := range apicompat.FinalizeResponsesChatStream(output.chat) {
			text, err := apicompat.ChatChunkToSSE(chunk)
			if err != nil {
				return err
			}
			text, err = customKiroAnnotateUsageSSE(text, output.usageSource)
			if err != nil {
				return err
			}
			output.write(text)
		}
		output.write("data: [DONE]\n\n")
	}
	return nil
}

func (output *customKiroOutput) failure(message string) {
	encoded, _ := json.Marshal(map[string]any{"type": "error", "error": map[string]string{"type": "upstream_error", "message": message}})
	if output.format == "chat" {
		output.write("data: " + string(encoded) + "\n\n")
	} else {
		output.write("event: error\ndata: " + string(encoded) + "\n\n")
	}
}

func (s *GatewayService) forwardCustomKiro(ctx context.Context, c *gin.Context, account *Account, body []byte, model string, stream bool, format string, includeUsage bool, mapping *apicompat.ResponsesClientToolMapping) (*ForwardResult, error) {
	started := time.Now()
	if s.customKiro == nil {
		return nil, errors.New("Kiro 服务未配置")
	}
	upstreamCtx, release := detachStreamUpstreamContext(ctx, stream)
	defer release()
	upstreamCtx, cancel := context.WithTimeout(upstreamCtx, 5*time.Minute)
	defer cancel()
	call, err := s.customKiro.call(upstreamCtx, account, body, model)
	if err != nil {
		var requestErr *customKiroRequestError
		if errors.As(err, &requestErr) {
			customKiroWriteError(c, format, http.StatusBadRequest, err.Error())
			return nil, err
		}
		return nil, &UpstreamFailoverError{StatusCode: http.StatusBadGateway, ResponseBody: []byte(`{"error":{"type":"upstream_error","message":"Kiro 认证或连接失败"}}`)}
	}
	response := call.Response
	defer func() { _ = response.Body.Close() }()
	requestID := response.Header.Get("x-amzn-requestid")
	if requestID == "" {
		requestID = response.Header.Get("x-amzn-request-id")
	}
	if requestID == "" {
		requestID = response.Header.Get("x-request-id")
	}
	if response.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
		safeBody, _ := json.Marshal(map[string]any{"error": map[string]string{"type": "upstream_error", "message": fmt.Sprintf("Kiro 上游返回 HTTP %d", response.StatusCode)}})
		if s.shouldFailoverUpstreamError(response.StatusCode) {
			if s.rateLimitService != nil {
				s.rateLimitService.HandleUpstreamError(ctx, call.Account, response.StatusCode, response.Header, safeBody)
			}
			return nil, &UpstreamFailoverError{StatusCode: response.StatusCode, ResponseBody: safeBody}
		}
		customKiroWriteError(c, format, http.StatusBadGateway, fmt.Sprintf("Kiro 上游返回 HTTP %d", response.StatusCode))
		return nil, fmt.Errorf("Kiro HTTP %d", response.StatusCode)
	}
	if !strings.HasPrefix(strings.ToLower(response.Header.Get("Content-Type")), "application/vnd.amazon.eventstream") {
		customKiroWriteError(c, format, http.StatusBadGateway, "Kiro 上游未返回 EventStream")
		return nil, errors.New("Kiro 上游响应类型无效")
	}
	output := newCustomKiroOutput(c, format, model, includeUsage, mapping)
	usageMode := call.Account.GetCredential("usage_mode")
	if usageMode == "" {
		usageMode = "verified"
	}
	c.Header("X-Sub2API-Kiro-Usage-Policy", usageMode)
	var firstToken *int
	consume := func(block kiro.ResponseBlock) error {
		if firstToken == nil {
			elapsed := int(time.Since(started).Milliseconds())
			firstToken = &elapsed
		}
		if stream {
			return output.consume(block)
		}
		return nil
	}
	decoded, decodeErr := kiro.DecodeResponse(upstreamCtx, response.Body, consume)
	if decodeErr == nil {
		decodeErr = resolveCustomKiroUsage(ctx, call.Account, body, model, decoded)
	}
	result := &ForwardResult{RequestID: requestID, UpstreamHeaders: response.Header, Model: model, UpstreamModel: call.Model, Stream: stream, Duration: time.Since(started), FirstTokenMs: firstToken, ClientDisconnect: output.disconnected}
	if decoded != nil && decoded.HasUsage {
		result.Usage = ClaudeUsage{InputTokens: decoded.Usage.InputTokens, OutputTokens: decoded.Usage.OutputTokens, CacheReadInputTokens: decoded.Usage.CacheReadTokens, CacheCreationInputTokens: decoded.Usage.CacheWriteTokens}
	}
	if decodeErr != nil {
		if output.started {
			output.failure(decodeErr.Error())
		} else {
			customKiroWriteError(c, format, http.StatusBadGateway, decodeErr.Error())
		}
		result.ClientDisconnect = output.disconnected
		return result, decodeErr
	}
	if stream {
		err = output.finish(decoded)
		result.ClientDisconnect = output.disconnected
		return result, err
	}
	message := map[string]any{"id": output.messageID, "type": "message", "role": "assistant", "model": model, "content": decoded.Blocks, "stop_reason": decoded.StopReason, "stop_sequence": nil, "usage": decoded.Usage}
	encoded, err := json.Marshal(message)
	if err != nil {
		return result, err
	}
	if format != "messages" {
		var anthropic apicompat.AnthropicResponse
		if err := json.Unmarshal(encoded, &anthropic); err != nil {
			return result, err
		}
		responses := apicompat.AnthropicToResponsesResponse(&anthropic)
		responses.Model = model
		if format == "chat" {
			encoded, err = json.Marshal(apicompat.ResponsesToChatCompletions(responses, model))
		} else {
			encoded, err = json.Marshal(responses)
		}
		if err != nil {
			return result, err
		}
		if mapping != nil {
			encoded, _, err = apicompat.RestoreResponsesClientToolPayload(encoded, *mapping)
			if err != nil {
				return result, err
			}
		}
	}
	encoded, err = customKiroAnnotateUsage(encoded, decoded.UsageSource)
	if err != nil {
		return result, err
	}
	c.Header("X-Sub2API-Kiro-Usage-Source", decoded.UsageSource)
	c.Data(http.StatusOK, "application/json; charset=utf-8", encoded)
	return result, nil
}

func customKiroWriteError(c *gin.Context, format string, status int, message string) {
	payload := gin.H{"error": gin.H{"type": "upstream_error", "message": message}}
	if status == http.StatusBadRequest {
		payload["error"] = gin.H{"type": "invalid_request_error", "message": message}
	}
	if format == "messages" {
		payload["type"] = "error"
	}
	c.JSON(status, payload)
}
