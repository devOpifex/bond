package mcp

import (
	"encoding/json"
	"fmt"
)

// Version is the JSON-RPC version
const Version = "2.0"

// Request represents a JSON-RPC 2.0 request
type Request struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
	ID      any    `json:"id,omitempty"`
}

// Response represents a JSON-RPC 2.0 response
type Response struct {
	JSONRPC string `json:"jsonrpc"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
	ID      any    `json:"id"`
}

// ToolListResult represents the result of a tools/list request
type ToolListResult struct {
	Tools      []Tool `json:"tools"`
	NextCursor string `json:"nextCursor,omitempty"`
}

// Tool represents a tool definition in the Model Context Protocol
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// ToolCallParams represents the parameters for a tools/call request
type ToolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// ToolCallResult represents the result of a tools/call request
type ToolCallResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError"`
}

// ToolContent represents a content item in a tool result
type ToolContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mimeType,omitempty"`

	// For resource type content
	Resource *ResourceContent `json:"resource,omitempty"`
}

// ResourceContent represents an embedded resource in a tool result
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
}

// Error represents a JSON-RPC 2.0 error object
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Error codes as defined in JSON-RPC 2.0 spec
const (
	// Standard JSON-RPC 2.0 errors
	ParseErrorCode     = -32700
	InvalidRequestCode = -32600
	MethodNotFoundCode = -32601
	InvalidParamsCode  = -32602
	InternalErrorCode  = -32603

	// Implementation specific error codes should be between -32000 and -32099
	ServerErrorCodeMin = -32099
	ServerErrorCodeMax = -32000
)

// Common error messages
const (
	ParseErrorMsg     = "Parse error"
	InvalidRequestMsg = "Invalid request"
	MethodNotFoundMsg = "Method not found"
	InvalidParamsMsg  = "Invalid params"
	InternalErrorMsg  = "Internal error"
)

// NewRequest creates a new JSON-RPC 2.0 request
func NewRequest(method string, params any, id any) *Request {
	return &Request{
		JSONRPC: Version,
		Method:  method,
		Params:  params,
		ID:      id,
	}
}

// NewResponse creates a new JSON-RPC 2.0 response
func NewResponse(result any, id any) *Response {
	return &Response{
		JSONRPC: Version,
		Result:  result,
		ID:      id,
	}
}

// NewErrorResponse creates a new JSON-RPC 2.0 error response
func NewErrorResponse(code int, message string, data any, id any) *Response {
	return &Response{
		JSONRPC: Version,
		Error: &Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
		ID: id,
	}
}

// Parse parses a JSON string into a Request
func Parse(data []byte) (*Request, error) {
	var request Request
	err := json.Unmarshal(data, &request)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON-RPC request: %w", err)
	}

	// Validate basic JSON-RPC 2.0 requirements
	if request.JSONRPC != Version {
		return &request, fmt.Errorf("invalid JSON-RPC version: expected %s", Version)
	}

	if request.Method == "" {
		return &request, fmt.Errorf("missing method in JSON-RPC request")
	}

	return &request, nil
}

// ParseResponse parses a JSON string into a Response
func ParseResponse(data []byte) (*Response, error) {
	var response Response
	err := json.Unmarshal(data, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON-RPC response: %w", err)
	}

	// Validate basic JSON-RPC 2.0 requirements
	if response.JSONRPC != Version {
		return &response, fmt.Errorf("invalid JSON-RPC version: expected %s", Version)
	}

	// A response must have either a result or an error, but not both
	if response.Result != nil && response.Error != nil {
		return &response, fmt.Errorf("invalid JSON-RPC response: both result and error present")
	}

	if response.Result == nil && response.Error == nil {
		return &response, fmt.Errorf("invalid JSON-RPC response: missing both result and error")
	}

	return &response, nil
}

// IsBatch determines if the JSON data is a batch request (array of requests)
func IsBatch(data []byte) bool {
	return len(data) > 0 && data[0] == '['
}

// ParseBatch parses a JSON array into multiple Request objects
func ParseBatch(data []byte) ([]*Request, error) {
	var requests []*Request
	err := json.Unmarshal(data, &requests)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON-RPC batch request: %w", err)
	}

	// Validate each request
	for i, req := range requests {
		if req.JSONRPC != Version {
			return requests, fmt.Errorf("invalid JSON-RPC version in request %d: expected %s", i, Version)
		}
		if req.Method == "" {
			return requests, fmt.Errorf("missing method in JSON-RPC request %d", i)
		}
	}

	return requests, nil
}
