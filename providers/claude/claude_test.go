package claude

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devOpifex/bond/models"
)

// MockTool is a simple tool implementation for testing
type MockTool struct {
	name        string
	description string
	schema      models.InputSchema
}

// GetName implements the ToolExecutor interface
func (t *MockTool) GetName() string {
	return t.name
}

// GetDescription implements the ToolExecutor interface
func (t *MockTool) GetDescription() string {
	return t.description
}

// GetSchema implements the ToolExecutor interface
func (t *MockTool) GetSchema() models.InputSchema {
	return t.schema
}

// Execute implements the ToolExecutor interface
func (t *MockTool) Execute(input json.RawMessage) (string, error) {
	return "MockTool executed successfully", nil
}

// TestNewClient tests that the client is properly initialized
func TestNewClient(t *testing.T) {
	client := NewClient("test-api-key")

	if client.ApiKey != "test-api-key" {
		t.Errorf("Expected API key 'test-api-key', got '%s'", client.ApiKey)
	}

	if client.BaseURL != "https://api.anthropic.com/v1/messages" {
		t.Errorf("Expected base URL 'https://api.anthropic.com/v1/messages', got '%s'", client.BaseURL)
	}

	if client.Model != "claude-3-sonnet-20240229" {
		t.Errorf("Expected default model 'claude-3-sonnet-20240229', got '%s'", client.Model)
	}

	if client.MaxTokens != 1000 {
		t.Errorf("Expected default max tokens 1000, got %d", client.MaxTokens)
	}

	if len(client.Tools) != 0 {
		t.Errorf("Expected empty tools map, got %d tools", len(client.Tools))
	}
}

// TestRegisterTool tests tool registration
func TestRegisterTool(t *testing.T) {
	client := NewClient("test-api-key")

	// Create and register a mock tool
	mockTool := &MockTool{
		name:        "test_tool",
		description: "A test tool",
		schema: models.InputSchema{
			Type: "object",
			Properties: map[string]models.Property{
				"param": {
					Type:        "string",
					Description: "A test parameter",
				},
			},
			Required: []string{"param"},
		},
	}

	client.RegisterTool(mockTool)

	// Verify the tool was registered
	if len(client.Tools) != 1 {
		t.Errorf("Expected 1 registered tool, got %d", len(client.Tools))
	}

	// Check if the tool can be retrieved
	tool, exists := client.Tools["test_tool"]
	if !exists {
		t.Error("Tool not found in registered tools")
	}

	if tool.GetName() != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got '%s'", tool.GetName())
	}
}

// TestSetModel tests model configuration
func TestSetModel(t *testing.T) {
	client := NewClient("test-api-key")
	
	// Test default model
	if client.Model != "claude-3-sonnet-20240229" {
		t.Errorf("Expected default model 'claude-3-sonnet-20240229', got '%s'", client.Model)
	}
	
	// Change model
	client.SetModel("claude-3-opus-20240229")
	
	// Verify model was changed
	if client.Model != "claude-3-opus-20240229" {
		t.Errorf("Expected model 'claude-3-opus-20240229', got '%s'", client.Model)
	}
}

// TestSetMaxTokens tests max tokens configuration
func TestSetMaxTokens(t *testing.T) {
	client := NewClient("test-api-key")
	
	// Test default max tokens
	if client.MaxTokens != 1000 {
		t.Errorf("Expected default max tokens 1000, got %d", client.MaxTokens)
	}
	
	// Change max tokens
	client.SetMaxTokens(2000)
	
	// Verify max tokens was changed
	if client.MaxTokens != 2000 {
		t.Errorf("Expected max tokens 2000, got %d", client.MaxTokens)
	}
}

// TestSendMessage tests sending a simple message
func TestSendMessage(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("x-api-key") != "test-api-key" {
			t.Errorf("Expected x-api-key header 'test-api-key', got '%s'", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("Expected anthropic-version header '2023-06-01', got '%s'", r.Header.Get("anthropic-version"))
		}

		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Send mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"content": [
				{
					"type": "text",
					"text": "This is a test response"
				}
			],
			"stop_reason": "end_turn"
		}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create client with mocked server URL
	client := NewClient("test-api-key")
	client.BaseURL = server.URL

	// Send a test message
	response, err := client.SendMessage(context.Background(), models.Message{
		Role:    models.RoleUser,
		Content: "Hello, Claude!",
	})
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Verify response
	expected := "This is a test response"
	if response != expected {
		t.Errorf("Expected response '%s', got '%s'", expected, response)
	}
}

// TestSendMessageWithTools tests sending a message with tools
func TestSendMessageWithTools(t *testing.T) {
	// Create a mock HTTP server that validates the request contains tools
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read request body to verify it contains tools
		var requestBody map[string]interface{}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&requestBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Verify the request contains tools
		tools, ok := requestBody["tools"].([]interface{})
		if !ok {
			t.Error("Request does not contain tools field")
		}

		if len(tools) == 0 {
			t.Error("Tools array is empty")
		}

		// Check the first tool's name
		tool := tools[0].(map[string]interface{})
		if tool["name"] != "test_tool" {
			t.Errorf("Expected tool name 'test_tool', got '%v'", tool["name"])
		}

		// Send mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"content": [
				{
					"type": "text",
					"text": "Response with tools available"
				}
			],
			"stop_reason": "end_turn"
		}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create client with mocked server URL
	client := NewClient("test-api-key")
	client.BaseURL = server.URL

	// Register a mock tool
	mockTool := &MockTool{
		name:        "test_tool",
		description: "A test tool",
		schema: models.InputSchema{
			Type: "object",
			Properties: map[string]models.Property{
				"param": {
					Type:        "string",
					Description: "A test parameter",
				},
			},
			Required: []string{"param"},
		},
	}
	client.RegisterTool(mockTool)

	// Send a test message with tools
	response, err := client.SendMessageWithTools(context.Background(), models.Message{
		Role:    models.RoleUser,
		Content: "Hello, Claude! Use tools if needed.",
	})
	if err != nil {
		t.Fatalf("Failed to send message with tools: %v", err)
	}

	// Verify response
	expected := "Response with tools available"
	if response != expected {
		t.Errorf("Expected response '%s', got '%s'", expected, response)
	}
}

// TestHandleToolCall tests the tool call handling functionality
func TestHandleToolCall(t *testing.T) {
	// Create a mock HTTP server that responds with a tool call
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send a response with a tool call
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"content": [
				{
					"type": "tool_use",
					"name": "test_tool",
					"input": {"param": "test value"}
				}
			],
			"stop_reason": "tool_use"
		}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create client with mocked server URL
	client := NewClient("test-api-key")
	client.BaseURL = server.URL

	// Register a mock tool
	mockTool := &MockTool{
		name:        "test_tool",
		description: "A test tool",
		schema: models.InputSchema{
			Type: "object",
			Properties: map[string]models.Property{
				"param": {
					Type:        "string",
					Description: "A test parameter",
				},
			},
			Required: []string{"param"},
		},
	}
	client.RegisterTool(mockTool)

	// Send a test message that will trigger a tool call
	response, err := client.SendMessageWithTools(context.Background(), models.Message{
		Role:    models.RoleUser,
		Content: "Use the test_tool please",
	})
	if err != nil {
		t.Fatalf("Failed to handle tool call: %v", err)
	}

	// Our new implementation includes the tool result within a formatted response
	// Just check that the tool result is contained in the response
	if !strings.Contains(response, "MockTool executed successfully") {
		t.Errorf("Response does not contain expected tool execution result: %s", response)
	}
}