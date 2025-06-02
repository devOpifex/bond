// Package common provides shared functionality for all LLM provider implementations.
// It includes a base client with common methods and utilities for HTTP communication,
// tool management, and configuration. This package reduces code duplication across
// different provider implementations.
package common

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

// HTTPRequest represents a generic HTTP request to be sent to an LLM provider API.
// It abstracts away the details of the HTTP protocol to simplify provider implementations.
type HTTPRequest struct {
	// Method is the HTTP method (GET, POST, etc.)
	Method string
	
	// URL is the endpoint to send the request to
	URL string
	
	// Headers contains HTTP headers to include in the request
	Headers map[string]string
	
	// Body contains the request payload as JSON bytes
	Body []byte
}

// BaseClient contains common fields and methods shared by all provider clients.
// It implements parts of the models.Provider interface and provides utility
// methods that specific provider implementations can use.
type BaseClient struct {
	// ApiKey is the authentication key for the provider API
	ApiKey string
	
	// BaseURL is the endpoint URL for the provider API
	BaseURL string
	
	// HttpClient is used for making HTTP requests to the provider API
	HttpClient *http.Client
	
	// Tools is a registry of tools that can be called by the model
	Tools map[string]models.ToolExecutor
	
	// Model is the specific model version to use (e.g., "gpt-4", "claude-3-sonnet")
	Model string
	
	// MaxTokens limits the length of the model's response
	MaxTokens int
	
	// Temperature controls randomness in the model's response generation
	// Higher values (e.g., 0.8) make output more random, while lower values (e.g., 0.2)
	// make it more deterministic. Range is typically 0.0-1.0.
	Temperature float64
	
	// SystemPrompt contains instructions included in all requests
	SystemPrompt string
}

// NewBaseClient creates a new base client with common configuration.
// It initializes fields that all provider clients need, such as the API key,
// HTTP client, and tool registry.
func NewBaseClient(apiKey string, baseURL string, defaultModel string) BaseClient {
	return BaseClient{
		ApiKey:  apiKey,
		BaseURL: baseURL,
		HttpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Tools:       make(map[string]models.ToolExecutor),
		Model:       defaultModel,
		MaxTokens:   1000,
		Temperature: 0.7, // Default temperature
	}
}

// RegisterTool adds a tool that the provider can call during its reasoning process.
// This implements part of the models.Provider interface.
func (c *BaseClient) RegisterTool(tool models.ToolExecutor) {
	c.Tools[tool.GetName()] = tool
}

// SetModel configures which specific model version to use for this provider.
// This implements part of the models.Provider interface.
func (c *BaseClient) SetModel(model string) {
	c.Model = model
}

// SetMaxTokens configures the maximum number of tokens in the model's response.
// This implements part of the models.Provider interface.
func (c *BaseClient) SetMaxTokens(tokens int) {
	c.MaxTokens = tokens
}

// SetSystemPrompt sets a system prompt that will be included in all requests.
// This implements part of the models.Provider interface.
func (c *BaseClient) SetSystemPrompt(prompt string) {
	c.SystemPrompt = prompt
}

// SetTemperature configures the temperature parameter for the model's response generation.
// Higher values (e.g., 0.8) make output more random, while lower values (e.g., 0.2)
// make it more deterministic. Range is typically 0.0-1.0.
// This implements part of the models.Provider interface.
func (c *BaseClient) SetTemperature(temperature float64) {
	c.Temperature = temperature
}

// DoHTTPRequest performs an HTTP request and returns the response body.
// It handles the details of creating the request, setting headers, sending it,
// and processing the response, including error handling.
func (c *BaseClient) DoHTTPRequest(ctx context.Context, req HTTPRequest) ([]byte, error) {
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bytes.NewBuffer(req.Body))
	if err != nil {
		return nil, err
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Execute request
	resp, err := c.HttpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check for non-200 status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// HandleToolCall executes the requested tool with the provided input.
// It looks up the tool in the registry, executes it with the given input,
// and returns the result or an error if the tool is not found or execution fails.
func (c *BaseClient) HandleToolCall(ctx context.Context, toolName string, input json.RawMessage) (string, error) {
	tool, exists := c.Tools[toolName]
	if !exists {
		return "", fmt.Errorf("tool %s not found", toolName)
	}

	result, err := tool.Execute(input)
	if err != nil {
		return "", fmt.Errorf("tool execution failed: %w", err)
	}

	return result, nil
}