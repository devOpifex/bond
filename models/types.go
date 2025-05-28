package models

// Role constants for message roles
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
	RoleFunction  = "function"
)

// ToolUse represents Claude's request to use a tool
type ToolUse struct {
	Name  string `json:"name"`
	Input string `json:"input"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ToolName string `json:"tool_name"`
	Result   string `json:"result"`
}

// Message represents a chat message
type Message struct {
	Role       string      `json:"role"`
	Content    string      `json:"content"`
	ToolUse    *ToolUse    `json:"tool_use,omitempty"`
	ToolResult *ToolResult `json:"tool_result,omitempty"`
}

// Tool defines a function that the AI model can call
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"input_schema"`
}

// InputSchema defines the structure of tool inputs
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

// Property defines a single property in a tool schema
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}