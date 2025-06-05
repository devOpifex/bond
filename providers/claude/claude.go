// Package claude implements the Provider interface for Anthropic's Claude models.
// It handles communication with Claude's API, including message formatting,
// tool registration, and configuration of model parameters.
package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/devOpifex/bond/models"
	"github.com/devOpifex/bond/providers/common"
)

// Provider implements the Provider interface for Claude AI models.
// It manages the connection to Claude's API and handles the specific
// requirements of Claude's message format and API interactions.
type Provider struct {
	// APIKey is the authentication token for Claude's API
	APIKey string
	
	// BaseURL is the endpoint for Claude's API
	BaseURL string
	
	// Model specifies which Claude model to use
	Model string
	
	// HTTPClient is used for making requests to Claude's API
	HTTPClient *http.Client
	
	// System prompt provides context and instructions for the model
	SystemPrompt string
	
	// MaxTokens limits the length of the model's response
	MaxTokens int
	
	// Temperature controls randomness in the model's output (0.0-1.0)
	Temperature float64
	
	// Tools that have been registered for use with this provider
	Tools []models.ToolExecutor
}

// New creates a new Claude provider with the given API key.
// It sets up default values that can be customized through
// the provider's configuration methods.
func New(apiKey string) *Provider {
	return &Provider{
		APIKey:      apiKey,
		BaseURL:     "https://api.anthropic.com/v1/messages",
		Model:       "claude-3-opus-20240229",
		HTTPClient:  &http.Client{Timeout: 60 * time.Second},
		MaxTokens:   1024,
		Temperature: 0.7,
		Tools:       []models.ToolExecutor{},
	}
}

// NewClient is an alias for New to maintain backward compatibility
func NewClient(apiKey string) *Provider {
	return New(apiKey)
}

// SetSystemPrompt sets the system prompt that guides Claude's behavior.
// The system prompt is included at the beginning of each conversation
// to provide context and instructions to the model.
func (p *Provider) SetSystemPrompt(prompt string) {
	p.SystemPrompt = prompt
}

// SetModel configures which specific Claude model to use.
// Available models include "claude-3-opus-20240229", "claude-3-sonnet-20240229", etc.
func (p *Provider) SetModel(model string) {
	p.Model = model
}

// SetMaxTokens configures the maximum number of tokens in Claude's response.
// Higher values allow for longer responses but may impact performance.
func (p *Provider) SetMaxTokens(tokens int) {
	p.MaxTokens = tokens
}

// SetTemperature configures the randomness of Claude's responses.
// Values closer to 0 make responses more deterministic, while
// values closer to 1 make responses more creative and varied.
func (p *Provider) SetTemperature(temperature float64) {
	p.Temperature = temperature
}

// RegisterTool adds a tool to the provider's available tools.
// These tools will be included in the API request to Claude,
// allowing the model to use them during its reasoning process.
func (p *Provider) RegisterTool(tool models.ToolExecutor) {
	// Check if the tool is already registered
	for _, existingTool := range p.Tools {
		if existingTool.GetName() == tool.GetName() {
			// Tool with this name already exists, don't add it again
			return
		}
	}
	
	// Tool not registered yet, add it
	p.Tools = append(p.Tools, tool)
}

// SendMessage sends a message to Claude and returns the model's response.
// This method does not include tool information in the request.
func (p *Provider) SendMessage(ctx context.Context, message models.Message) (string, error) {
	return p.sendRequest(ctx, message, false)
}

// SendMessageWithTools sends a message to Claude with available tools.
// This method includes information about registered tools in the request,
// allowing Claude to call these tools during its reasoning process.
func (p *Provider) SendMessageWithTools(ctx context.Context, message models.Message) (string, error) {
	fmt.Printf("SendMessageWithTools called with %d registered tools\n", len(p.Tools))
	for i, tool := range p.Tools {
		fmt.Printf("Tool %d: %s - %s\n", i, tool.GetName(), tool.GetDescription())
	}
	return p.sendRequest(ctx, message, true)
}

// prepareChatContext constructs the conversation history to send to Claude.
// It handles extracting message history from the context if available,
// and applying Claude-specific formatting for tool results.
func (p *Provider) prepareChatContext(ctx context.Context, message models.Message) []models.Message {
	// Extract message history from context if available
	var messageHistory []models.Message
	if history, ok := ctx.Value("message_history").([]models.Message); ok {
		messageHistory = make([]models.Message, len(history))
		copy(messageHistory, history)
	} else {
		messageHistory = []models.Message{}
	}

	// Special handling for messages to Claude API
	// Tool results need to be formatted specially - for Claude we need to preserve
	// the original message and include the tool result as a user message
	if message.Role == models.RoleFunction && message.ToolResult != nil {
		// Format the tool result message
		userMsg := models.Message{
			Role: models.RoleUser,
			Content: fmt.Sprintf("Tool '%s' returned: %s",
				message.ToolResult.Name, message.Content),
		}

		// Add the tool result as a user message to the history
		messageHistory = append(messageHistory, userMsg)
	} else {
		// For regular messages, add to history
		messageHistory = append(messageHistory, message)
	}

	// Create a clean message list for Claude with only supported roles (user/assistant)
	return convertMessagesToClaudeFormat(messageHistory)
}

// prepareToolsForRequest converts the registered tools to Claude's API format.
// It builds a list of tool definitions that Claude can understand and use.
func (p *Provider) prepareToolsForRequest() []map[string]interface{} {
	toolDefinitions := make([]map[string]interface{}, 0, len(p.Tools))
	
	for _, tool := range p.Tools {
		// Convert each tool to Claude's expected format
		schema := tool.GetSchema()
		
		// Convert the schema to a JSON Schema compatible format
		toolDef := map[string]interface{}{
			"name":        tool.GetName(),
			"description": tool.GetDescription(),
			"input_schema": map[string]interface{}{
				"type":       schema.Type,
				"properties": schema.Properties,
			},
		}
		
		// Add required fields if any
		if len(schema.Required) > 0 {
			toolDef["input_schema"].(map[string]interface{})["required"] = schema.Required
		}
		
		toolDefinitions = append(toolDefinitions, toolDef)
	}
	
	return toolDefinitions
}

// sendRequest handles the common logic for sending requests to Claude's API.
// It prepares the request payload, sends it to the API, and processes the response.
func (p *Provider) sendRequest(ctx context.Context, message models.Message, withTools bool) (string, error) {
	// Build the Claude-formatted messages
	messages := p.prepareChatContext(ctx, message)
	
	// Create the request payload
	payload := map[string]interface{}{
		"model":       p.Model,
		"messages":    messages,
		"max_tokens":  p.MaxTokens,
		"temperature": p.Temperature,
	}
	
	// Add system prompt if provided
	if p.SystemPrompt != "" {
		payload["system"] = p.SystemPrompt
	}
	
	// Add tools if requested and available
	if withTools && len(p.Tools) > 0 {
		toolDefinitions := p.prepareToolsForRequest()
		fmt.Printf("Sending %d tool definitions to Claude\n", len(toolDefinitions))
		payload["tools"] = toolDefinitions
	}
	
	// Convert the payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Claude request: %w", err)
	}
	
	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return "", fmt.Errorf("failed to create Claude request: %w", err)
	}
	
	// Set the required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	
	// Send the request
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Claude API request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Claude API response: %w", err)
	}
	
	// Check for API error
	if resp.StatusCode != http.StatusOK {
		// Try to parse the error message
		var errorResp common.ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error != nil {
			return "", fmt.Errorf("Claude API error (%d): %s", resp.StatusCode, errorResp.Error.Message)
		}
		return "", fmt.Errorf("Claude API error (%d): %s", resp.StatusCode, string(body))
	}
	
	// Parse the successful response
	fmt.Printf("Got successful response from Claude: %s\n", string(body))
	
	var claudeResp struct {
		Content []struct {
			Type  string          `json:"type"`
			Text  string          `json:"text,omitempty"`
			Name  string          `json:"name,omitempty"`
			Input json.RawMessage `json:"input,omitempty"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
	}
	
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", fmt.Errorf("failed to parse Claude API response: %w", err)
	}
	
	// Check if Claude wants to use a tool
	if claudeResp.StopReason == "tool_use" {
		// Find the tool use request
		for _, content := range claudeResp.Content {
			if content.Type == "tool_use" {
				fmt.Printf("Tool use request: %s with input %s\n", content.Name, string(content.Input))
				
				// Find the tool
				tool, exists := p.findTool(content.Name)
				if !exists {
					return fmt.Sprintf("Error: Tool '%s' not found", content.Name), nil
				}
				
				// Execute the tool
				result, err := tool.Execute(content.Input)
				if err != nil {
					return fmt.Sprintf("Error executing tool '%s': %v", content.Name, err), nil
				}
				
				// Now we need to send the tool result back to Claude in a new request
				toolResultMessage := models.Message{
					Role:    models.RoleUser,
					Content: fmt.Sprintf("Tool '%s' returned: %s", content.Name, result),
				}
				
				// Recursive call to send the tool result back to Claude
				return p.sendRequest(ctx, toolResultMessage, true)
			}
		}
	}
	
	// Extract the text from the response
	var responseText string
	for _, content := range claudeResp.Content {
		fmt.Printf("Response content type: %s\n", content.Type)
		if content.Type == "text" {
			responseText += content.Text
		}
	}
	
	return responseText, nil
}

// findTool looks up a tool by name in the provider's tools
func (p *Provider) findTool(name string) (models.ToolExecutor, bool) {
	for _, tool := range p.Tools {
		if tool.GetName() == name {
			return tool, true
		}
	}
	return nil, false
}

// convertMessagesToClaudeFormat transforms our internal message format to Claude's API format.
// Claude only supports user and assistant roles, so this function handles the conversion
// of system and function messages appropriately.
func convertMessagesToClaudeFormat(messages []models.Message) []models.Message {
	// Claude only supports user and assistant roles
	claudeMessages := []models.Message{}
	
	for _, msg := range messages {
		switch msg.Role {
		case models.RoleUser:
			// User messages pass through directly
			claudeMessages = append(claudeMessages, msg)
		case models.RoleAssistant:
			// Assistant messages pass through directly
			claudeMessages = append(claudeMessages, msg)
		case models.RoleSystem:
			// System messages in Claude are handled separately, not as a message
			// We don't include them in the messages array
			continue
		case models.RoleFunction:
			// Function messages are transformed to user messages in the prepare method
			// This should already be handled, but just in case
			userMsg := models.Message{
				Role:    models.RoleUser,
				Content: fmt.Sprintf("Function returned: %s", msg.Content),
			}
			claudeMessages = append(claudeMessages, userMsg)
		default:
			// Unknown roles are sent as user messages
			userMsg := models.Message{
				Role:    models.RoleUser,
				Content: msg.Content,
			}
			claudeMessages = append(claudeMessages, userMsg)
		}
	}
	
	return claudeMessages
}