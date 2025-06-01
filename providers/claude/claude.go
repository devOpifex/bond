// Package claude implements the Anthropic Claude API integration for the Bond framework.
// It provides a client for communicating with Claude models, handling message formatting,
// tool calls, and response parsing according to Claude's specific API requirements.
package claude

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/devOpifex/bond/models"
	"github.com/devOpifex/bond/providers/common"
)

// ClaudeRequest represents a request to the Claude API.
// It follows the structure expected by Anthropic's Messages API.
type ClaudeRequest struct {
	// Model specifies which Claude model version to use
	Model string `json:"model"`

	// MaxTokens limits the length of the response
	MaxTokens int `json:"max_tokens"`

	// System contains instructions that guide Claude's behavior
	System string `json:"system,omitempty"`

	// Messages contains the conversation history
	Messages []models.Message `json:"messages"`

	// Tools defines functions that Claude can call
	Tools []models.Tool `json:"tools,omitempty"`
}

// ClaudeResponse represents a response from the Claude API.
// It contains content blocks and metadata about why the response ended.
type ClaudeResponse struct {
	// Content contains the response blocks (text or tool calls)
	Content []ContentBlock `json:"content"`

	// StopReason indicates why Claude stopped generating (length, tool_use, etc.)
	StopReason string `json:"stop_reason"`
}

// ContentBlock represents a single block in a Claude response.
// It can be either text content or a tool use request.
type ContentBlock struct {
	// Type indicates the block type ("text" or "tool_use")
	Type string `json:"type"`

	// Text contains the response text for text blocks
	Text string `json:"text,omitempty"`

	// Name contains the tool name for tool_use blocks
	Name string `json:"name,omitempty"`

	// Input contains the parameters for tool calls
	Input json.RawMessage `json:"input,omitempty"`
}

// Client is the Claude API client implementation.
// It handles communication with the Anthropic API, including authentication,
// request formatting, and response parsing. It implements the models.Provider interface.
type Client struct {
	common.BaseClient
}

// NewClient creates a new Claude client with the provided API key.
// It initializes the client with default settings for the Claude API.
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

// SendMessage sends a message with specified role to Claude and returns the response.
// This implements part of the models.Provider interface for basic message exchange
// without tool capabilities.
func (c *Client) SendMessage(ctx context.Context, message models.Message) (string, error) {
	// Get message history from context or create a new one
	var messageHistory []models.Message
	historyValue := ctx.Value("message_history")
	if historyValue != nil {
		if history, ok := historyValue.([]models.Message); ok {
			messageHistory = history
		}
	}

	// Add the current message to history
	messageHistory = append(messageHistory, message)
	
	// Create a clean message list for Claude with only supported roles (user/assistant)
	var cleanMessages []models.Message
	for _, msg := range messageHistory {
		// Skip any messages with unsupported roles
		if msg.Role == models.RoleUser || msg.Role == models.RoleAssistant {
			cleanMessages = append(cleanMessages, msg)
		}
	}

	request := ClaudeRequest{
		Model:     c.Model,
		MaxTokens: c.MaxTokens,
		Messages:  cleanMessages,
	}

	// Add system prompt if set
	if c.SystemPrompt != "" {
		request.System = c.SystemPrompt
	}

	// Update message history in context for next call
	ctx = context.WithValue(ctx, "message_history", messageHistory)

	return c.sendRequest(ctx, request)
}

// SendMessageWithTools sends a message with specified role to Claude with registered tools.
// This implements part of the models.Provider interface for advanced interactions
// where the model may need to call tools during its reasoning process.
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

	// Get message history from context or create a new one
	var messageHistory []models.Message
	historyValue := ctx.Value("message_history")
	if historyValue != nil {
		if history, ok := historyValue.([]models.Message); ok {
			messageHistory = history
		}
	}

	// Special handling for messages to Claude API
	// Tool results need to be formatted specially - for Claude we need to preserve
	// the original message and include the tool result as a user message
	if message.Role == models.RoleFunction && message.ToolResult != nil {
		// Format the tool result message
		userMsg := models.Message{
			Role: models.RoleUser,
			Content: fmt.Sprintf("Tool '%s' returned: %s",
				message.ToolResult.ToolName, message.Content),
		}

		// Add the tool result as a user message to the history
		messageHistory = append(messageHistory, userMsg)
	} else {
		// For regular messages, add to history
		messageHistory = append(messageHistory, message)
	}
	
	// Create a clean message list for Claude with only supported roles (user/assistant)
	var cleanMessages []models.Message
	for _, msg := range messageHistory {
		// Skip any messages with unsupported roles
		if msg.Role == models.RoleUser || msg.Role == models.RoleAssistant {
			cleanMessages = append(cleanMessages, msg)
		}
	}

	request := ClaudeRequest{
		Model:     c.Model,
		MaxTokens: c.MaxTokens,
		Messages:  cleanMessages,
		Tools:     tools,
	}

	// Add system prompt if set
	if c.SystemPrompt != "" {
		request.System = c.SystemPrompt
	}

	// Update message history in context for next call
	ctx = context.WithValue(ctx, "message_history", messageHistory)

	return c.sendRequest(ctx, request)
}

// sendRequest sends a request to the Claude API and processes the response.
// It handles the HTTP communication, error handling, and response parsing.
// If Claude requests a tool, it manages the tool execution and formats the result.
func (c *Client) sendRequest(ctx context.Context, request ClaudeRequest) (string, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	// Prepare HTTP request
	headers := map[string]string{
		"Content-Type":      "application/json",
		"x-api-key":         c.ApiKey,
		"anthropic-version": "2023-06-01",
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
				"name":  toolUseBlock.Name,
				"input": string(inputJSON),
			})

			combinedResponse := fmt.Sprintf("%s```json\n%s\n```\n\nTool result: %s",
				thoughtText, string(toolJSON), result)

			// Add tool response to message history
			historyValue := ctx.Value("message_history")
			if historyValue != nil {
				if messageHistory, ok := historyValue.([]models.Message); ok {
					// Create assistant message with the tool response
					assistantMessage := models.Message{
						Role:    models.RoleAssistant,
						Content: combinedResponse,
					}

					// Add to history and update context
					messageHistory = append(messageHistory, assistantMessage)
					ctx = context.WithValue(ctx, "message_history", messageHistory)
				}
			}

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
		// Add assistant's response to message history
		historyValue := ctx.Value("message_history")
		if historyValue != nil {
			if messageHistory, ok := historyValue.([]models.Message); ok {
				// Create assistant message with the response
				assistantMessage := models.Message{
					Role:    models.RoleAssistant,
					Content: textResponse,
				}

				// Add to history and update context
				messageHistory = append(messageHistory, assistantMessage)
				ctx = context.WithValue(ctx, "message_history", messageHistory)
			}
		}

		return textResponse, nil
	}

	return "No response received", nil
}

