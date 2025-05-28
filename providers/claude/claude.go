package claude

import (
	"context"
	"encoding/json"

	"github.com/devOpifex/bond/models"
	"github.com/devOpifex/bond/providers/common"
)

// ClaudeRequest represents a request to the Claude API
type ClaudeRequest struct {
	Model     string           `json:"model"`
	MaxTokens int              `json:"max_tokens"`
	System    string           `json:"system,omitempty"`
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
	common.BaseClient
}

// NewClient creates a new Claude client
func NewClient(apiKey string) *Client {
	baseClient := common.NewBaseClient(
		apiKey,
		"https://api.anthropic.com/v1/messages",
		"claude-3-sonnet-20240229",
	)
	
	return &Client{
		BaseClient: baseClient,
	}
}

// SendMessage sends a message with specified role to Claude
func (c *Client) SendMessage(ctx context.Context, message models.Message) (string, error) {
	request := ClaudeRequest{
		Model:     c.Model,
		MaxTokens: c.MaxTokens,
		Messages:  []models.Message{message},
	}

	// Add system prompt if set
	if c.SystemPrompt != "" {
		request.System = c.SystemPrompt
	}

	return c.sendRequest(ctx, request)
}

// SendMessageWithTools sends a message with specified role to Claude with registered tools
func (c *Client) SendMessageWithTools(ctx context.Context, message models.Message) (string, error) {
	// Convert registered tools to Claude tool format
	var tools []models.Tool
	for _, tool := range c.Tools {
		tools = append(tools, models.Tool{
			Name:        tool.GetName(),
			Description: tool.GetDescription(),
			InputSchema: tool.GetSchema(),
		})
	}

	request := ClaudeRequest{
		Model:     c.Model,
		MaxTokens: c.MaxTokens,
		Messages:  []models.Message{message},
		Tools:     tools,
	}

	// Add system prompt if set
	if c.SystemPrompt != "" {
		request.System = c.SystemPrompt
	}

	// Store the original message in context
	ctx = context.WithValue(ctx, "original_message", message)

	return c.sendRequest(ctx, request)
}

// sendRequest sends a request to the Claude API
func (c *Client) sendRequest(ctx context.Context, request ClaudeRequest) (string, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	// Prepare HTTP request
	headers := map[string]string{
		"Content-Type":       "application/json",
		"x-api-key":          c.ApiKey,
		"anthropic-version":  "2023-06-01",
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
			result, err := c.HandleToolCall(ctx, toolUseBlock.Name, toolUseBlock.Input)
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