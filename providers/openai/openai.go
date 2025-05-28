package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/devOpifex/bond/models"
)

// OpenAIRequest represents a request to the OpenAI API
type OpenAIRequest struct {
	Model       string            `json:"model"`
	MaxTokens   int               `json:"max_tokens"`
	Messages    []OpenAIMessage   `json:"messages"`
	Tools       []OpenAITool      `json:"tools,omitempty"`
	ToolChoice  string            `json:"tool_choice,omitempty"`
	Temperature float64           `json:"temperature"`
}

// OpenAIMessage represents a message in OpenAI format
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAITool represents a tool in OpenAI format
type OpenAITool struct {
	Type     string          `json:"type"`
	Function OpenAIFunction  `json:"function"`
}

// OpenAIFunction represents a function in OpenAI format
type OpenAIFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// OpenAIResponse represents a response from the OpenAI API
type OpenAIResponse struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Choices []OpenAIChoice       `json:"choices"`
	Usage   map[string]int       `json:"usage"`
}

// OpenAIChoice represents a choice in an OpenAI response
type OpenAIChoice struct {
	Index        int              `json:"index"`
	Message      OpenAIRespMessage `json:"message"`
	FinishReason string           `json:"finish_reason"`
}

// OpenAIRespMessage represents a message in an OpenAI response
type OpenAIRespMessage struct {
	Role         string              `json:"role"`
	Content      string              `json:"content"`
	ToolCalls    []OpenAIToolCall    `json:"tool_calls,omitempty"`
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

// Client is the OpenAI API client implementation
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	tools      map[string]models.ToolExecutor
	model      string
	maxTokens  int
}

// NewClient creates a new OpenAI client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1/chat/completions",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		tools:     make(map[string]models.ToolExecutor),
		model:     "gpt-4o",
		maxTokens: 1000,
	}
}

// RegisterTool adds a tool that OpenAI can call
func (c *Client) RegisterTool(tool models.ToolExecutor) {
	c.tools[tool.GetName()] = tool
}

// SetModel configures which model to use
func (c *Client) SetModel(model string) {
	c.model = model
}

// SetMaxTokens configures the maximum number of tokens in the response
func (c *Client) SetMaxTokens(tokens int) {
	c.maxTokens = tokens
}

// SendMessage sends a message with specified role to OpenAI
func (c *Client) SendMessage(ctx context.Context, message models.Message) (string, error) {
	request := OpenAIRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		Messages: []OpenAIMessage{
			{Role: message.Role, Content: message.Content},
		},
		Temperature: 0.7,
	}

	return c.sendRequest(ctx, request)
}

// convertToolSchema converts our schema format to OpenAI's format
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

// SendMessageWithTools sends a message with specified role to OpenAI with registered tools
func (c *Client) SendMessageWithTools(ctx context.Context, message models.Message) (string, error) {
	// Convert registered tools to OpenAI tool format
	var tools []OpenAITool
	for _, tool := range c.tools {
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

	request := OpenAIRequest{
		Model:      c.model,
		MaxTokens:  c.maxTokens,
		Messages: []OpenAIMessage{
			{Role: message.Role, Content: message.Content},
		},
		Tools:       tools,
		ToolChoice:  "auto",
		Temperature: 0.7,
	}

	return c.sendRequest(ctx, request)
}

// sendRequest sends a request to the OpenAI API
func (c *Client) sendRequest(ctx context.Context, request OpenAIRequest) (string, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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
		return c.handleToolCall(ctx, toolCall.Function.Name, []byte(toolCall.Function.Arguments))
	}

	// Otherwise return the text content
	return choice.Message.Content, nil
}

// handleToolCall executes the requested tool
func (c *Client) handleToolCall(ctx context.Context, toolName string, input json.RawMessage) (string, error) {
	tool, exists := c.tools[toolName]
	if !exists {
		return "", fmt.Errorf("tool %s not found", toolName)
	}

	result, err := tool.Execute(input)
	if err != nil {
		return "", fmt.Errorf("tool execution failed: %w", err)
	}

	return result, nil
}