package claude

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

// ClaudeRequest represents a request to the Claude API
type ClaudeRequest struct {
	Model     string           `json:"model"`
	MaxTokens int              `json:"max_tokens"`
	Messages  []models.Message `json:"messages"`
	Tools     []models.Tool    `json:"tools,omitempty"`
}

// ClaudeResponse represents a response from the Claude API
type ClaudeResponse struct {
	Content    []ContentBlock `json:"content"`
	StopReason string         `json:"stop_reason"`
}

// ContentBlock represents a content block in a Claude response
type ContentBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

// Client is the Claude API client implementation
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	tools      map[string]models.ToolExecutor
	model      string
	maxTokens  int
}

// NewClient creates a new Claude client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://api.anthropic.com/v1/messages",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		tools:     make(map[string]models.ToolExecutor),
		model:     "claude-3-sonnet-20240229",
		maxTokens: 1000,
	}
}

// RegisterTool adds a tool that Claude can call
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

// SendMessage sends a simple text message to Claude
func (c *Client) SendMessage(ctx context.Context, content string) (string, error) {
	request := ClaudeRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		Messages: []models.Message{
			{Role: "user", Content: content},
		},
	}

	return c.sendRequest(ctx, request)
}

// SendMessageWithTools sends a message to Claude with registered tools
func (c *Client) SendMessageWithTools(ctx context.Context, content string) (string, error) {
	// Convert registered tools to Claude tool format
	var tools []models.Tool
	for _, tool := range c.tools {
		tools = append(tools, models.Tool{
			Name:        tool.GetName(),
			Description: tool.GetDescription(),
			InputSchema: tool.GetSchema(),
		})
	}

	// Create the user message
	userMessage := models.Message{Role: "user", Content: content}

	request := ClaudeRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		Messages:  []models.Message{userMessage},
		Tools:     tools,
	}

	// Store the original message in context
	ctx = context.WithValue(ctx, "original_message", userMessage)

	return c.sendRequest(ctx, request)
}

// sendRequest sends a request to the Claude API
func (c *Client) sendRequest(ctx context.Context, request ClaudeRequest) (string, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", err
	}

	// If stop_reason is "tool_use", Claude wants to use a tool
	if claudeResp.StopReason == "tool_use" {
		// Find the tool_use block
		var toolUseBlock *ContentBlock
		var textBlocks []string
		
		for _, block := range claudeResp.Content {
			if block.Type == "tool_use" {
				toolUseBlock = &block
			} else if block.Type == "text" {
				textBlocks = append(textBlocks, block.Text)
			}
		}
		
		if toolUseBlock != nil {
			result, err := c.handleToolCall(ctx, toolUseBlock.Name, toolUseBlock.Input)
			if err != nil {
				return "", err
			}
			
			// Combine the text blocks with the tool result
			combinedResponse := ""
			for _, text := range textBlocks {
				combinedResponse += text + "\n"
			}
			combinedResponse += result
			
			return combinedResponse, nil
		}
	}
	
	// If Claude didn't request a tool or we couldn't find the tool_use block,
	// just return any text blocks
	var textResponse string
	for _, block := range claudeResp.Content {
		if block.Type == "text" {
			textResponse += block.Text
		}
	}
	
	if textResponse != "" {
		return textResponse, nil
	}

	return "No response received", nil
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