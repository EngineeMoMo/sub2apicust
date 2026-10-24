package kiro

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type Message struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}
type Tool struct {
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}
type Request struct {
	Model      string          `json:"model"`
	Messages   []Message       `json:"messages"`
	System     json.RawMessage `json:"system"`
	Tools      []Tool          `json:"tools"`
	ToolChoice struct {
		Type string `json:"type"`
		Name string `json:"name"`
	} `json:"tool_choice"`
	OutputConfig struct {
		Effort string          `json:"effort"`
		Format json.RawMessage `json:"format"`
	} `json:"output_config"`
	StopSequences []string `json:"stop_sequences"`
}

type ContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	Thinking  string          `json:"thinking,omitempty"`
	Signature string          `json:"signature,omitempty"`
	Data      string          `json:"data,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
	IsError   bool            `json:"is_error,omitempty"`
	Source    struct {
		Type      string `json:"type"`
		MediaType string `json:"media_type"`
		Data      string `json:"data"`
	} `json:"source,omitempty"`
	CacheControl json.RawMessage `json:"cache_control,omitempty"`
}

type Payload struct {
	ConversationState Conversation   `json:"conversationState"`
	ProfileARN        string         `json:"profileArn,omitempty"`
	AdditionalFields  map[string]any `json:"additionalModelRequestFields,omitempty"`
}
type Conversation struct {
	ID      string               `json:"conversationId"`
	Trigger string               `json:"chatTriggerType"`
	Task    string               `json:"agentTaskType"`
	Current map[string]UserInput `json:"currentMessage"`
	History []HistoryEntry       `json:"history,omitempty"`
}
type HistoryEntry struct {
	User      *UserInput      `json:"userInputMessage,omitempty"`
	Assistant *AssistantInput `json:"assistantResponseMessage,omitempty"`
}
type UserInput struct {
	Content    string            `json:"content"`
	Model      string            `json:"modelId"`
	Origin     string            `json:"origin"`
	Context    *InputContext     `json:"userInputMessageContext,omitempty"`
	Images     []Image           `json:"images,omitempty"`
	CachePoint map[string]string `json:"cachePoint,omitempty"`
}
type Image struct {
	Format string            `json:"format"`
	Source map[string]string `json:"source"`
}
type InputContext struct {
	Tools   []map[string]any `json:"tools,omitempty"`
	Results []ToolResult     `json:"toolResults,omitempty"`
}
type ToolResult struct {
	ID      string              `json:"toolUseId"`
	Status  string              `json:"status"`
	Content []map[string]string `json:"content"`
}
type AssistantInput struct {
	Content   string            `json:"content"`
	Tools     []HistoryTool     `json:"toolUses,omitempty"`
	Reasoning map[string]string `json:"reasoningContent,omitempty"`
}
type HistoryTool struct {
	ID    string          `json:"toolUseId"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

var modelPattern = regexp.MustCompile("^[a-zA-Z0-9][a-zA-Z0-9._-]{0,199}$")
var datedModelPattern = regexp.MustCompile("-[0-9]{8}$")
var claudeModelPattern = regexp.MustCompile("^claude-(sonnet|opus|haiku|fable)-[0-9]+([.-][0-9]+)?$")
var toolPattern = regexp.MustCompile("^[a-zA-Z0-9_-]{1,64}$")

func DefaultModels() []string {
	return []string{"claude-sonnet-4-6", "claude-opus-4-6", "claude-haiku-4-5-20251001", "claude-sonnet-4-5-20250929", "claude-opus-4-5-20251101"}
}

func ResolveModel(model string) (string, error) {
	original := model
	oneMillion := strings.HasSuffix(strings.ToLower(model), "[1m]")
	if oneMillion {
		model = model[:len(model)-4]
	}
	undated := datedModelPattern.ReplaceAllString(model, "")
	if !claudeModelPattern.MatchString(undated) {
		if oneMillion || !modelPattern.MatchString(original) {
			return "", fmt.Errorf("Kiro 模型标识无效或不支持该模型的 [1m] 后缀")
		}
		return original, nil
	}
	model = undated
	for _, family := range []string{"sonnet", "opus", "haiku", "fable"} {
		prefix := "claude-" + family + "-"
		if strings.HasPrefix(model, prefix) {
			version := strings.TrimPrefix(model, prefix)
			if len(version) == 3 && version[1] == '-' {
				model = prefix + version[:1] + "." + version[2:]
			}
		}
	}
	if oneMillion {
		if model != "claude-sonnet-4.5" && model != "claude-sonnet-4.6" {
			return "", fmt.Errorf("Kiro 该模型不支持 [1m] 后缀")
		}
		model += "-1m"
	}
	if !modelPattern.MatchString(model) {
		return "", fmt.Errorf("Kiro 模型标识无效")
	}
	return model, nil
}

func BuildPayload(body []byte, model, profile, conversationID string) (*Payload, error) {
	var request Request
	if err := json.Unmarshal(body, &request); err != nil {
		return nil, fmt.Errorf("Kiro 请求 JSON 无效")
	}
	if len(request.Messages) == 0 {
		return nil, fmt.Errorf("Kiro messages 不能为空")
	}
	if len(request.StopSequences) > 0 {
		return nil, fmt.Errorf("Kiro 尚不支持 stop_sequences")
	}
	if len(request.OutputConfig.Format) > 0 && string(request.OutputConfig.Format) != "null" {
		return nil, fmt.Errorf("Kiro 尚不支持结构化输出约束")
	}
	if request.ToolChoice.Type != "" && request.ToolChoice.Type != "auto" && request.ToolChoice.Type != "none" {
		return nil, fmt.Errorf("Kiro 仅支持 auto/none 工具选择")
	}
	modelID, err := ResolveModel(model)
	if err != nil {
		return nil, err
	}
	system, systemBlocks, err := textBlocks(request.System)
	if err != nil {
		return nil, fmt.Errorf("Kiro system: %w", err)
	}
	for _, block := range systemBlocks {
		if block.Type != "text" {
			return nil, fmt.Errorf("Kiro system 仅支持文本内容")
		}
	}
	tools := make([]map[string]any, 0, len(request.Tools))
	knownTools := make(map[string]bool)
	for _, tool := range request.Tools {
		if tool.Type != "" && tool.Type != "custom" {
			return nil, fmt.Errorf("Kiro 不支持内置工具 %s", tool.Type)
		}
		if !toolPattern.MatchString(tool.Name) || knownTools[tool.Name] || tool.Name == "thinking" {
			return nil, fmt.Errorf("Kiro 工具名称无效、重复或保留")
		}
		knownTools[tool.Name] = true
		if tool.InputSchema == nil {
			return nil, fmt.Errorf("Kiro 工具 %s 缺少 input_schema", tool.Name)
		}
		tools = append(tools, map[string]any{"toolSpecification": map[string]any{"name": tool.Name, "description": tool.Description, "inputSchema": map[string]any{"json": tool.InputSchema}}})
	}
	if request.ToolChoice.Type == "none" {
		tools = nil
	}
	entries := make([]HistoryEntry, 0, len(request.Messages))
	pendingTools := make(map[string]bool)
	usedIDs := make(map[string]bool)
	for messageIndex, message := range request.Messages {
		text, blocks, err := textBlocks(message.Content)
		if err != nil {
			return nil, fmt.Errorf("Kiro message %d: %w", messageIndex, err)
		}
		if message.Role == "user" {
			user := &UserInput{Content: text, Model: modelID, Origin: "KIRO_CLI"}
			inputContext := &InputContext{}
			for _, block := range blocks {
				switch block.Type {
				case "text":
					if len(block.CacheControl) > 0 {
						user.CachePoint = map[string]string{"type": "default"}
					}
				case "image":
					if block.Source.Type != "base64" {
						return nil, fmt.Errorf("Kiro 仅支持 base64 图片")
					}
					format := strings.TrimPrefix(block.Source.MediaType, "image/")
					if format != "png" && format != "jpeg" && format != "gif" && format != "webp" {
						return nil, fmt.Errorf("Kiro 图片格式不支持")
					}
					decoded, err := base64.StdEncoding.DecodeString(block.Source.Data)
					if err != nil || len(decoded) == 0 || len(decoded) > 4<<20 {
						return nil, fmt.Errorf("Kiro 图片数据无效或超过4MiB")
					}
					user.Images = append(user.Images, Image{Format: format, Source: map[string]string{"bytes": block.Source.Data}})
				case "tool_result":
					if !pendingTools[block.ToolUseID] {
						return nil, fmt.Errorf("Kiro tool_result 没有匹配的工具调用")
					}
					resultText, resultBlocks, err := textBlocks(block.Content)
					if err != nil {
						return nil, fmt.Errorf("Kiro tool_result 无效")
					}
					for _, resultBlock := range resultBlocks {
						if resultBlock.Type != "text" {
							return nil, fmt.Errorf("Kiro tool_result 仅支持文本")
						}
					}
					status := "success"
					if block.IsError {
						status = "error"
					}
					inputContext.Results = append(inputContext.Results, ToolResult{ID: block.ToolUseID, Status: status, Content: []map[string]string{{"text": resultText}}})
					delete(pendingTools, block.ToolUseID)
				default:
					return nil, fmt.Errorf("Kiro 用户内容类型 %s 不支持", block.Type)
				}
			}
			if len(inputContext.Results) > 0 {
				user.Context = inputContext
			}
			if user.Content == "" && len(user.Images) > 0 {
				user.Content = "请查看图片。"
			}
			if len(entries) > 0 && entries[len(entries)-1].User != nil {
				previous := entries[len(entries)-1].User
				previous.Content = joinText(previous.Content, user.Content)
				previous.Images = append(previous.Images, user.Images...)
				if user.CachePoint != nil {
					previous.CachePoint = user.CachePoint
				}
				if user.Context != nil {
					if previous.Context == nil {
						previous.Context = &InputContext{}
					}
					previous.Context.Results = append(previous.Context.Results, user.Context.Results...)
				}
			} else {
				entries = append(entries, HistoryEntry{User: user})
			}
		} else if message.Role == "assistant" {
			if len(pendingTools) > 0 && (len(entries) == 0 || entries[len(entries)-1].Assistant == nil) {
				return nil, fmt.Errorf("Kiro 工具结果未全部返回")
			}
			assistant := &AssistantInput{Content: text}
			for _, block := range blocks {
				switch block.Type {
				case "text", "thinking":
				case "redacted_thinking":
					assistant.Reasoning = map[string]string{"redactedContent": block.Data}
				case "tool_use":
					if block.ID == "" || usedIDs[block.ID] || !toolPattern.MatchString(block.Name) {
						return nil, fmt.Errorf("Kiro 工具调用ID或名称无效")
					}
					if !isJSONObject(block.Input) {
						return nil, fmt.Errorf("Kiro 工具输入必须为 JSON 对象")
					}
					usedIDs[block.ID] = true
					pendingTools[block.ID] = true
					assistant.Tools = append(assistant.Tools, HistoryTool{ID: block.ID, Name: block.Name, Input: block.Input})
				default:
					return nil, fmt.Errorf("Kiro 助手内容类型 %s 不支持", block.Type)
				}
			}
			if len(entries) > 0 && entries[len(entries)-1].Assistant != nil {
				previous := entries[len(entries)-1].Assistant
				previous.Content = joinText(previous.Content, assistant.Content)
				previous.Tools = append(previous.Tools, assistant.Tools...)
				if assistant.Reasoning != nil {
					previous.Reasoning = assistant.Reasoning
				}
			} else {
				entries = append(entries, HistoryEntry{Assistant: assistant})
			}
		} else {
			return nil, fmt.Errorf("Kiro 消息角色必须为 user/assistant")
		}
	}
	if len(pendingTools) > 0 {
		return nil, fmt.Errorf("Kiro 工具结果未全部返回")
	}
	if len(entries) == 0 || entries[0].User == nil || entries[len(entries)-1].User == nil {
		return nil, fmt.Errorf("Kiro 会话须由用户消息开始和结束，暂不支持助手预填")
	}
	entries[0].User.Content = joinText(system, entries[0].User.Content)
	current := entries[len(entries)-1].User
	if current.Content == "" && (current.Context == nil || len(current.Context.Results) == 0) {
		return nil, fmt.Errorf("Kiro 当前消息不能为空")
	}
	if len(tools) > 0 {
		if current.Context == nil {
			current.Context = &InputContext{}
		}
		current.Context.Tools = tools
	}
	payload := &Payload{ProfileARN: profile, ConversationState: Conversation{ID: conversationID, Trigger: "MANUAL", Task: "vibe", Current: map[string]UserInput{"userInputMessage": *current}, History: entries[:len(entries)-1]}}
	if effort := request.OutputConfig.Effort; effort != "" {
		switch effort {
		case "low", "medium", "high", "xhigh", "max":
		default:
			return nil, fmt.Errorf("Kiro reasoning effort 无效")
		}
		key := "output_config"
		if strings.HasPrefix(modelID, "gpt-") {
			key = "reasoning"
		}
		payload.AdditionalFields = map[string]any{key: map[string]string{"effort": effort}}
	}
	return payload, nil
}

func textBlocks(input json.RawMessage) (string, []ContentBlock, error) {
	if len(input) == 0 || string(input) == "null" {
		return "", nil, nil
	}
	var text string
	if err := json.Unmarshal(input, &text); err == nil {
		return text, nil, nil
	}
	var blocks []ContentBlock
	if err := json.Unmarshal(input, &blocks); err != nil {
		return "", nil, fmt.Errorf("内容必须为字符串或内容块数组")
	}
	var texts []string
	for _, block := range blocks {
		if block.Type == "text" {
			texts = append(texts, block.Text)
		}
	}
	return strings.Join(texts, "\n"), blocks, nil
}

func joinText(left, right string) string {
	if left == "" {
		return right
	}
	if right == "" {
		return left
	}
	return left + "\n\n" + right
}
func isJSONObject(input []byte) bool {
	var object map[string]json.RawMessage
	return json.Unmarshal(input, &object) == nil && object != nil
}
