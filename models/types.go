package models

// Role constants define the standard roles used in LLM conversations.
// These roles indicate who is speaking in each message of a conversation.
const (
	// RoleUser represents messages from the end user.
	RoleUser = "user"
	
	// RoleAssistant represents messages from the AI assistant.
	RoleAssistant = "assistant"
	
	// RoleSystem represents system instructions that guide the AI's behavior.
	RoleSystem = "system"
	
	// RoleFunction represents messages containing results from function/tool calls.
	RoleFunction = "function"
)

// ToolUse represents an AI model's request to use a tool.
// It contains the tool name and input parameters formatted as a JSON string.
type ToolUse struct {
	// Name is the identifier of the tool to be called.
	Name string `json:"name"`
	
	// Input contains the parameters for the tool call as a JSON-formatted string.
	Input string `json:"input"`
}

// ToolResult represents the result of a tool execution.
// It's used to track and communicate the outcome of tool calls back to the AI.
type ToolResult struct {
	// ToolName is the name of the tool that was executed.
	ToolName string `json:"tool_name"`
	
	// Result contains the string output from the tool execution.
	Result string `json:"result"`
}

// Message represents a single message in a conversation with an AI model.
// It can be a user input, AI response, system instruction, or tool-related message.
type Message struct {
	// Role identifies who is speaking (user, assistant, system, or function).
	Role string `json:"role"`
	
	// Content contains the actual message text.
	Content string `json:"content"`
	
	// ToolUse is present when an AI requests to use a tool.
	ToolUse *ToolUse `json:"tool_use,omitempty"`
	
	// ToolResult is present when including the result of a tool execution.
	ToolResult *ToolResult `json:"tool_result,omitempty"`
}

// Tool defines a function that the AI model can call during its reasoning process.
// This structure is used to register tools with the provider and inform the AI
// about available tools and their parameters.
type Tool struct {
	// Name is the identifier used to call this tool.
	Name string `json:"name"`
	
	// Description explains what the tool does, helping the AI decide when to use it.
	Description string `json:"description"`
	
	// InputSchema defines the expected structure of inputs to the tool.
	InputSchema InputSchema `json:"input_schema"`
}

// InputSchema defines the structure of tool inputs following a simplified JSON Schema format.
// It specifies the parameters a tool accepts, their types, and which ones are required.
type InputSchema struct {
	// Type is usually "object" for tool inputs.
	Type string `json:"type"`
	
	// Properties maps parameter names to their type definitions.
	Properties map[string]Property `json:"properties"`
	
	// Required lists which parameters must be provided.
	Required []string `json:"required,omitempty"`
}

// Property defines a single parameter in a tool's input schema.
// It specifies the parameter's type and provides a description.
type Property struct {
	// Type is the JSON type of this property (string, number, boolean, etc.).
	Type string `json:"type"`
	
	// Description explains what this parameter is used for.
	Description string `json:"description"`
}