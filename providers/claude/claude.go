package claude

import (
	"context"
	"encoding/json"
	"fmt"

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

	// Build message list - Claude only accepts user and assistant roles
	var messages []models.Message
	
	// Special handling for messages to Claude API
	// Tool results need to be formatted specially - for Claude we need to preserve
	// the original message and include the tool result as a user message
	
	// Check if we have a previous message from the context
	prevMessage, hasPrevMessage := ctx.Value("original_message").(models.Message)
	
	if message.Role == models.RoleFunction && message.ToolResult != nil {
		// Format the tool result message 
		userMsg := models.Message{
			Role:    models.RoleUser,
			Content: fmt.Sprintf("Tool '%s' returned: %s", 
				message.ToolResult.ToolName, message.Content),
		}
		
		// If we have a previous message and it's an assistant message, include it first
		if hasPrevMessage && prevMessage.Role == models.RoleAssistant {
			messages = append(messages, prevMessage)
		}
		
		// Then add the tool result as a user message
		messages = append(messages, userMsg)
	} else {
		// For regular messages, just pass them through
		messages = append(messages, message)
	}

	request := ClaudeRequest{
		Model:     c.Model,
		MaxTokens: c.MaxTokens,
		Messages:  messages,
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
			
			// Format response as expected by ReAct agent
			// This is the key modification - format tool calls in the way ReAct expects
			thoughtText := ""
			if len(textBlocks) > 0 {
				thoughtText = "<thought>\n"
				for _, text := range textBlocks {
					thoughtText += text + "\n"
				}
				thoughtText += "</thought>\n\n"
			}
			
			// Convert raw input JSON to a string for the ReAct format
			var inputMap map[string]interface{}
			json.Unmarshal(toolUseBlock.Input, &inputMap)
			inputJSON, _ := json.Marshal(map[string]interface{}{
				"expression": fmt.Sprintf("%v", inputMap["expression"]),
			})
			
			toolJSON, _ := json.Marshal(map[string]interface{}{
				"name": toolUseBlock.Name,
				"input": string(inputJSON),
			})
			
			combinedResponse := fmt.Sprintf("%s```json\n%s\n```\n\nTool result: %s", 
				thoughtText, string(toolJSON), result)
			
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