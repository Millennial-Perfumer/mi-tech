package llm

import (
	"context"
	"encoding/json"
)

type ChatRole string

const (
	RoleUser      ChatRole = "user"
	RoleAssistant ChatRole = "assistant"
	RoleSystem    ChatRole = "system"
	RoleTool      ChatRole = "tool"
)

type ChatMessage struct {
	Role       ChatRole   `json:"role"`
	Content    string     `json:"content"`
	ToolCallID *string    `json:"tool_call_id,omitempty"` // For 'tool' role
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`   // For 'assistant' role
}

func (m ChatMessage) MarshalJSON() ([]byte, error) {
	type Alias ChatMessage
	switch m.Role {
	case RoleAssistant:
		// Assistant can have ToolCalls but not ToolCallID
		return json.Marshal(&struct {
			Alias
			ToolCallID *string `json:"tool_call_id,omitempty"`
		}{
			Alias:      Alias(m),
			ToolCallID: nil,
		})
	case RoleTool:
		// Tool must have ToolCallID but not ToolCalls
		return json.Marshal(&struct {
			Alias
			ToolCalls []ToolCall `json:"tool_calls,omitempty"`
		}{
			Alias:     Alias(m),
			ToolCalls: nil,
		})
	default:
		// User/System have neither
		return json.Marshal(&struct {
			Alias
			ToolCallID *string    `json:"tool_call_id,omitempty"`
			ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
		}{
			Alias:      Alias(m),
			ToolCallID: nil,
			ToolCalls:  nil,
		})
	}
}

type ToolDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // always "function"
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type ChatResponse struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type StreamChunk struct {
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Error     error      `json:"error,omitempty"`
	Done      bool       `json:"done"`
}

type LLMProvider interface {
	Chat(ctx context.Context, messages []ChatMessage, tools []ToolDef) (*ChatResponse, error)
	StreamChat(ctx context.Context, messages []ChatMessage, tools []ToolDef) (<-chan StreamChunk, error)
	Name() string
}
