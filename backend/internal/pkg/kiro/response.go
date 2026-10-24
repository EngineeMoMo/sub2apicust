package kiro

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Usage struct {
	InputTokens      int `json:"input_tokens"`
	OutputTokens     int `json:"output_tokens"`
	CacheReadTokens  int `json:"cache_read_input_tokens"`
	CacheWriteTokens int `json:"cache_creation_input_tokens"`
}
type ResponseBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	Thinking  string          `json:"thinking,omitempty"`
	Signature string          `json:"signature,omitempty"`
	Data      string          `json:"data,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
}
type Response struct {
	Blocks      []ResponseBlock
	Usage       Usage
	HasUsage    bool
	UsageSource string
	EventCounts map[string]int
	StopReason  string
}
type toolAccumulator struct {
	ID    string
	Name  string
	Input strings.Builder
}

func DecodeResponse(ctx context.Context, reader io.Reader, emit func(ResponseBlock) error) (*Response, error) {
	response := &Response{StopReason: "end_turn", EventCounts: make(map[string]int)}
	var pending toolAccumulator
	completed := make(map[string]bool)
	appendBlock := func(block ResponseBlock) error {
		if emit != nil {
			if err := emit(block); err != nil {
				return err
			}
		}
		if len(response.Blocks) > 0 {
			last := &response.Blocks[len(response.Blocks)-1]
			if block.Type == last.Type && (block.Type == "text" || block.Type == "thinking") {
				last.Text += block.Text
				last.Thinking += block.Thinking
				last.Signature += block.Signature
				return nil
			}
		}
		response.Blocks = append(response.Blocks, block)
		return nil
	}
	flushTool := func() error {
		if pending.ID == "" {
			return nil
		}
		if pending.Name == "" || completed[pending.ID] {
			return fmt.Errorf("Kiro 工具事件缺少名称或ID重复")
		}
		input := pending.Input.String()
		if input == "" {
			input = "{}"
		}
		if !isJSONObject([]byte(input)) {
			return fmt.Errorf("Kiro 工具输入 JSON 不完整")
		}
		completed[pending.ID] = true
		if pending.Name == "thinking" {
			var thought struct {
				Thought string `json:"thought"`
			}
			if err := json.Unmarshal([]byte(input), &thought); err != nil {
				return fmt.Errorf("Kiro thinking 工具数据无效")
			}
			if err := appendBlock(ResponseBlock{Type: "thinking", Thinking: thought.Thought}); err != nil {
				return err
			}
		} else {
			response.StopReason = "tool_use"
			if err := appendBlock(ResponseBlock{Type: "tool_use", ID: pending.ID, Name: pending.Name, Input: json.RawMessage(input)}); err != nil {
				return err
			}
		}
		pending = toolAccumulator{}
		return nil
	}
	err := ReadEvents(ctx, reader, func(event Event) error {
		if !json.Valid(event.Payload) {
			return fmt.Errorf("Kiro %s 事件 JSON 无效", event.Type)
		}
		switch event.Type {
		case "contextUsageEvent":
			response.EventCounts[event.Type]++
		case "assistantResponseEvent":
			response.EventCounts[event.Type]++
			var payload struct {
				Content string `json:"content"`
			}
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				return fmt.Errorf("Kiro 文本事件无效")
			}
			if payload.Content != "" {
				if err := flushTool(); err != nil {
					return err
				}
				return appendBlock(ResponseBlock{Type: "text", Text: payload.Content})
			}
		case "reasoningContentEvent":
			var payload struct {
				Text            string `json:"text"`
				Signature       string `json:"signature"`
				RedactedContent string `json:"redactedContent"`
			}
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				return fmt.Errorf("Kiro 思考事件无效")
			}
			if err := flushTool(); err != nil {
				return err
			}
			if payload.RedactedContent != "" {
				return appendBlock(ResponseBlock{Type: "redacted_thinking", Data: payload.RedactedContent})
			}
			if payload.Text != "" || payload.Signature != "" {
				return appendBlock(ResponseBlock{Type: "thinking", Thinking: payload.Text, Signature: payload.Signature})
			}
		case "toolUseEvent":
			var payload struct {
				ID    string          `json:"toolUseId"`
				Name  string          `json:"name"`
				Input json.RawMessage `json:"input"`
				Stop  bool            `json:"stop"`
			}
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				return fmt.Errorf("Kiro 工具事件无效")
			}
			if payload.ID != "" && payload.ID != pending.ID {
				if err := flushTool(); err != nil {
					return err
				}
				pending.ID = payload.ID
			}
			if pending.ID == "" {
				return fmt.Errorf("Kiro 工具事件缺少 toolUseId")
			}
			if payload.Name != "" {
				if pending.Name != "" && pending.Name != payload.Name {
					return fmt.Errorf("Kiro 工具名称在分片间变化")
				}
				pending.Name = payload.Name
			}
			if len(payload.Input) > 0 {
				if payload.Input[0] == '"' {
					var fragment string
					if err := json.Unmarshal(payload.Input, &fragment); err != nil {
						return err
					}
					pending.Input.WriteString(fragment)
				} else {
					pending.Input.Reset()
					pending.Input.Write(payload.Input)
				}
				if pending.Input.Len() > 1<<20 {
					return fmt.Errorf("Kiro 工具输入超过1MiB")
				}
			}
			if payload.Stop {
				return flushTool()
			}
		case "metadataEvent":
			response.EventCounts[event.Type]++
			var payload struct {
				TokenUsage *struct {
					Input      *int `json:"uncachedInputTokens"`
					Output     *int `json:"outputTokens"`
					CacheRead  int  `json:"cacheReadInputTokens"`
					CacheWrite int  `json:"cacheWriteInputTokens"`
				} `json:"tokenUsage"`
			}
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				return fmt.Errorf("Kiro 用量事件无效")
			}
			if payload.TokenUsage != nil {
				usage := payload.TokenUsage
				if usage.Input == nil || usage.Output == nil {
					return fmt.Errorf("Kiro 用量缺少 uncachedInputTokens/outputTokens")
				}
				for _, count := range []int{*usage.Input, *usage.Output, usage.CacheRead, usage.CacheWrite} {
					if count < 0 || count > 1000000000 {
						return fmt.Errorf("Kiro 用量计数超出范围")
					}
				}
				if *usage.Input > 0 || *usage.Output > 0 || usage.CacheRead > 0 || usage.CacheWrite > 0 {
					response.Usage = Usage{InputTokens: *usage.Input, OutputTokens: *usage.Output, CacheReadTokens: usage.CacheRead, CacheWriteTokens: usage.CacheWrite}
					response.HasUsage = true
					response.UsageSource = "metadata"
				}
			}
		case "meteringEvent":
			response.EventCounts[event.Type]++
			var payload struct {
				Input  *int `json:"inputTokens"`
				Output *int `json:"outputTokens"`
			}
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				return fmt.Errorf("Kiro metering 用量事件无效")
			}
			for _, count := range []*int{payload.Input, payload.Output} {
				if count != nil && (*count < 0 || *count > 1000000000) {
					return fmt.Errorf("Kiro 用量计数超出范围")
				}
			}
			if response.UsageSource != "metadata" && payload.Input != nil && payload.Output != nil && (*payload.Input > 0 || *payload.Output > 0) {
				response.Usage = Usage{InputTokens: *payload.Input, OutputTokens: *payload.Output}
				response.HasUsage = true
				response.UsageSource = "metering"
			}
		case "invalidStateEvent":
			return fmt.Errorf("Kiro 返回 invalidStateEvent")
		}
		return nil
	})
	if err != nil {
		return response, err
	}
	if err := flushTool(); err != nil {
		return response, err
	}
	if len(response.Blocks) == 0 {
		return response, fmt.Errorf("Kiro 响应内容为空")
	}
	return response, nil
}
