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

// HTTPRequest represents a generic HTTP request
type HTTPRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
}

// BaseClient contains common fields and methods shared by all provider clients
type BaseClient struct {
	ApiKey       string
	BaseURL      string
	HttpClient   *http.Client
	Tools        map[string]models.ToolExecutor
	Model        string
	MaxTokens    int
	SystemPrompt string
}

// NewBaseClient creates a new base client with common configuration
func NewBaseClient(apiKey string, baseURL string, defaultModel string) BaseClient {
	return BaseClient{
		ApiKey:  apiKey,
		BaseURL: baseURL,
		HttpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Tools:     make(map[string]models.ToolExecutor),
		Model:     defaultModel,
		MaxTokens: 1000,
	}
}

// RegisterTool adds a tool that the provider can call
func (c *BaseClient) RegisterTool(tool models.ToolExecutor) {
	c.Tools[tool.GetName()] = tool
}

// SetModel configures which model to use
func (c *BaseClient) SetModel(model string) {
	c.Model = model
}

// SetMaxTokens configures the maximum number of tokens in the response
func (c *BaseClient) SetMaxTokens(tokens int) {
	c.MaxTokens = tokens
}

// SetSystemPrompt sets a system prompt that will be included in all requests
func (c *BaseClient) SetSystemPrompt(prompt string) {
	c.SystemPrompt = prompt
}

// DoHTTPRequest performs an HTTP request and returns the response body
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

// HandleToolCall executes the requested tool
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