package models

import (
	"context"
	"encoding/json"
)

// ToolExecutor interface for tools that can be executed
type ToolExecutor interface {
	GetName() string
	GetDescription() string
	GetSchema() InputSchema
	Execute(input json.RawMessage) (string, error)
}

// Provider defines the interface that all AI providers must implement
type Provider interface {
	// SendMessage sends a message with the specified role and content to the AI provider
	SendMessage(ctx context.Context, message Message) (string, error)
	
	// SendMessageWithTools sends a message with the specified role and content, plus available tools to the AI provider
	SendMessageWithTools(ctx context.Context, message Message) (string, error)
	
	// RegisterTool adds a tool that the AI provider can call
	RegisterTool(tool ToolExecutor)
	
	// SetModel configures which model to use for this provider
	SetModel(model string)
	
	// SetMaxTokens configures the maximum number of tokens in the response
	SetMaxTokens(tokens int)
}

// Agent defines the interface that all agents must implement
type Agent interface {
	// Process handles the input and returns a response
	Process(ctx context.Context, input string) (string, error)
}