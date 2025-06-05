// Package models defines the core interfaces and data types used throughout the Bond framework.
// It serves as the foundation layer containing shared data structures and interfaces that
// other packages implement and use.
package models

import (
	"context"
	"encoding/json"
)

// ToolExecutor defines the interface for tools that can be executed by AI models.
// Tools are functions that AI models can call to perform actions or retrieve information.
type ToolExecutor interface {
	// GetName returns the name of the tool, which is used to identify it when called.
	GetName() string

	// GetDescription returns a human-readable description of what the tool does.
	GetDescription() string

	// GetSchema returns the input schema that defines the structure of input parameters.
	GetSchema() InputSchema

	// Execute runs the tool with the provided JSON input and returns a string result or error.
	Execute(input json.RawMessage) (string, error)
}

// Provider defines the interface that all AI providers (like OpenAI, Claude) must implement.
// It handles communication with LLM APIs, including message formatting, tool registration,
// and configuration of model parameters.
type Provider interface {
	// SendMessage sends a message with the specified role and content to the AI provider
	// and returns the model's response as a string.
	SendMessage(ctx context.Context, message Message) (string, error)

	// SendMessageWithTools sends a message with the specified role and content to the AI provider,
	// along with information about available tools that the model can call.
	// It returns the model's response as a string.
	SendMessageWithTools(ctx context.Context, message Message) (string, error)

	// RegisterTool adds a tool that the AI provider can call during its reasoning process.
	// Tools are registered with the provider so they can be included in API requests.
	RegisterTool(tool ToolExecutor)

	// SetSystemPrompt sets a system prompt that will be included in all requests to guide
	// the model's behavior and provide context for the conversation.
	SetSystemPrompt(prompt string)

	// SetModel configures which specific model version to use for this provider
	// (e.g., "gpt-4", "claude-3-sonnet").
	SetModel(model string)

	// SetMaxTokens configures the maximum number of tokens that the model should
	// generate in its response.
	SetMaxTokens(tokens int)

	// SetTemperature configures the temperature parameter for the model's response generation.
	// Higher values (e.g., 0.8) make output more random, while lower values (e.g., 0.2)
	// make it more deterministic. Range is typically 0.0-1.0.
	SetTemperature(temperature float64)

	// register an MCP with the provider
	registerMCP(command string, args []string) error
}

// Agent defines the interface that all AI agents must implement.
// Agents are higher-level constructs that process user inputs and manage
// the interaction flow with AI models, potentially using multiple steps
// or tools to fulfill requests.
type Agent interface {
	// Process handles the user input, performs any necessary reasoning or tool use,
	// and returns a final response. The context can carry conversation history
	// or other state information needed for processing.
	Process(ctx context.Context, input string) (string, error)
}

