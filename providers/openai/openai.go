// Package openai implements the OpenAI API integration for the Bond framework.
// It provides a client for communicating with OpenAI models like GPT-4,
// handling message formatting, tool calls, and response parsing according
// to OpenAI's specific API requirements.
package openai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/devOpifex/bond/models"
	"github.com/devOpifex/bond/providers/common"
)

// OpenAIRequest represents a request to the OpenAI API
type OpenAIRequest struct {
	Model       string          `json:"model"`
	MaxTokens   int             `json:"max_tokens"`
	Messages    []OpenAIMessage `json:"messages"`
	Tools       []OpenAITool    `json:"tools,omitempty"`
	ToolChoice  string          `json:"tool_choice,omitempty"`
	Temperature float64         `json:"temperature"`
}

// OpenAIMessage represents a message in OpenAI format
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAITool represents a tool in OpenAI format
type OpenAITool struct {
	Type     string         `json:"type"`
	Function OpenAIFunction `json:"function"`
}

// OpenAIFunction represents a function in OpenAI format
type OpenAIFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// OpenAIResponse represents a response from the OpenAI API
type OpenAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   map[string]int `json:"usage"`
}

// OpenAIChoice represents a choice in an OpenAI response
type OpenAIChoice struct {
	Index        int               `json:"index"`
	Message      OpenAIRespMessage `json:"message"`
	FinishReason string            `json:"finish_reason"`
}

// OpenAIRespMessage represents a message in an OpenAI response
type OpenAIRespMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []OpenAIToolCall `json:"tool_calls,omitempty"`
}

// OpenAIToolCall represents a tool call in an OpenAI response
type OpenAIToolCall struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Function OpenAIFuncCall `json:"function"`
}

// OpenAIFuncCall represents a function call in an OpenAI response
type OpenAIFuncCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Client is the OpenAI API client implementation.
// It handles communication with the OpenAI API, including authentication,
// request formatting, and response parsing. It implements the models.Provider interface.
type Client struct {
	common.BaseClient
}

// NewClient creates a new OpenAI client with the provided API key.
// It initializes the client with default settings for the OpenAI API,
// including the base URL and default model (gpt-4o).
func NewClient(apiKey string) *Client {
	baseClient := common.NewBaseClient(
		apiKey,
		"https://api.openai.com/v1/chat/completions",
		"gpt-4o",
	)

	return &Client{
		BaseClient: baseClient,
	}
}

// SendMessage sends a message with specified role to OpenAI and returns the response.
// This implements part of the models.Provider interface for basic message exchange
// without tool capabilities.
func (c *Client) SendMessage(ctx context.Context, message models.Message) (string, error) {
	var messages []OpenAIMessage

	// Add system prompt if set
	if c.SystemPrompt != "" {
		messages = append(messages, OpenAIMessage{
			Role:    models.RoleSystem,
			Content: c.SystemPrompt,
		})
	}

	// Add the user/assistant message
	messages = append(messages, OpenAIMessage{
		Role:    message.Role,
		Content: message.Content,
	})

	request := OpenAIRequest{
		Model:       c.Model,
		MaxTokens:   c.MaxTokens,
		Messages:    messages,
		Temperature: c.Temperature,
	}

	return c.sendRequest(ctx, request)
}

// convertToolSchema converts our schema format to OpenAI's JSON Schema format.
// OpenAI expects tool parameters in JSON Schema format, which differs slightly
// from our internal representation. This function handles the conversion.
func convertToolSchema(schema models.InputSchema) (json.RawMessage, error) {
	// OpenAI expects a JSON Schema format
	openaiSchema := map[string]interface{}{
		"type": schema.Type,
	}

	if len(schema.Properties) > 0 {
		props := make(map[string]interface{})
		for name, prop := range schema.Properties {
			props[name] = map[string]string{
				"type":        prop.Type,
				"description": prop.Description,
			}
		}
		openaiSchema["properties"] = props
	}

	if len(schema.Required) > 0 {
		openaiSchema["required"] = schema.Required
	}

	return json.Marshal(openaiSchema)
}

// SendMessageWithTools sends a message with specified role to OpenAI with registered tools.
// This implements part of the models.Provider interface for advanced interactions
// where the model may need to call tools during its reasoning process.
func (c *Client) SendMessageWithTools(ctx context.Context, message models.Message) (string, error) {
	// Convert registered tools to OpenAI tool format
	var tools []OpenAITool
	for _, tool := range c.Tools {
		// Convert our schema to OpenAI schema
		parametersJSON, err := convertToolSchema(tool.GetSchema())
		if err != nil {
			return "", fmt.Errorf("failed to convert tool schema: %w", err)
		}

		tools = append(tools, OpenAITool{
			Type: "function",
			Function: OpenAIFunction{
				Name:        tool.GetName(),
				Description: tool.GetDescription(),
				Parameters:  parametersJSON,
			},
		})
	}

	var messages []OpenAIMessage

	// Add system prompt if set
	if c.SystemPrompt != "" {
		messages = append(messages, OpenAIMessage{
			Role:    models.RoleSystem,
			Content: c.SystemPrompt,
		})
	}

	// Add the user/assistant message
	messages = append(messages, OpenAIMessage{
		Role:    message.Role,
		Content: message.Content,
	})

	request := OpenAIRequest{
		Model:       c.Model,
		MaxTokens:   c.MaxTokens,
		Messages:    messages,
		Tools:       tools,
		ToolChoice:  "auto",
		Temperature: c.Temperature,
	}

	return c.sendRequest(ctx, request)
}

// sendRequest sends a request to the OpenAI API and processes the response.
// It handles the HTTP communication, error handling, and response parsing.
// If OpenAI requests a tool, it manages the tool execution and returns the result.
func (c *Client) sendRequest(ctx context.Context, request OpenAIRequest) (string, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	// Prepare HTTP request
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + c.ApiKey,
	}

	httpReq := common.HTTPRequest{
		Method:  "POST",
		URL:     c.BaseURL,
		Headers: headers,
		Body:    jsonData,
	}

	// Send the request
	body, err := c.DoHTTPRequest(ctx, httpReq)
	if err != nil {
		return "", err
	}

	var openaiResp OpenAIResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return "", err
	}

	// Check if we have choices
	if len(openaiResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	// Get the first choice
	choice := openaiResp.Choices[0]

	// Handle tool calls if OpenAI wants to use a tool
	if len(choice.Message.ToolCalls) > 0 {
		toolCall := choice.Message.ToolCalls[0]
		return c.HandleToolCall(ctx, toolCall.Function.Name, []byte(toolCall.Function.Arguments))
	}

	// Otherwise return the text content
	return choice.Message.Content, nil
}

