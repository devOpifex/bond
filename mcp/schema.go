package mcp

// MCPRequest represents the incoming request structure for the MCP protocol
type MCPRequest struct {
	Command  string         `json:"command"`
	ToolName string         `json:"tool_name,omitempty"`
	Inputs   map[string]any `json:"inputs,omitempty"`
	ID       string         `json:"id,omitempty"`
}

// MCPResponse represents the outgoing response structure for the MCP protocol
type MCPResponse struct {
	Status  string           `json:"status"`
	Data    map[string]any   `json:"data,omitempty"`
	Error   string           `json:"error,omitempty"`
	ID      string           `json:"id,omitempty"`
	ToolUse *ToolUseResponse `json:"tool_use,omitempty"`
}

// ToolUseResponse represents the results of a tool use request
type ToolUseResponse struct {
	Name   string         `json:"name"`
	Result map[string]any `json:"result,omitempty"`
	Error  string         `json:"error,omitempty"`
}
