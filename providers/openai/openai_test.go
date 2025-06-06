package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

// TestNewClient tests that the OpenAI client is properly initialized
func TestNewClient(t *testing.T) {
	client := NewClient("test-api-key")

	if client.ApiKey != "test-api-key" {
		t.Errorf("Expected API key 'test-api-key', got '%s'", client.ApiKey)
	}

	if client.BaseURL != "https://api.openai.com/v1/chat/completions" {
		t.Errorf("Expected base URL 'https://api.openai.com/v1/chat/completions', got '%s'", client.BaseURL)
	}

	if client.Model != "gpt-4o" {
		t.Errorf("Expected default model 'gpt-4o', got '%s'", client.Model)
	}

	if client.MaxTokens != 1000 {
		t.Errorf("Expected default max tokens 1000, got %d", client.MaxTokens)
	}

	if client.Temperature != 0.7 {
		t.Errorf("Expected default temperature 0.7, got %f", client.Temperature)
	}

	if len(client.Tools) != 0 {
		t.Errorf("Expected empty tools map, got %d tools", len(client.Tools))
	}
}

// TestSetTemperature tests temperature configuration
func TestSetTemperature(t *testing.T) {
	client := NewClient("test-api-key")
	
	// Test default temperature
	if client.Temperature != 0.7 {
		t.Errorf("Expected default temperature 0.7, got %f", client.Temperature)
	}
	
	// Change temperature
	client.SetTemperature(0.2)
	
	// Verify temperature was changed
	if client.Temperature != 0.2 {
		t.Errorf("Expected temperature 0.2, got %f", client.Temperature)
	}
	
	// Test edge case - very low temperature
	client.SetTemperature(0.0)
	if client.Temperature != 0.0 {
		t.Errorf("Expected temperature 0.0, got %f", client.Temperature)
	}
	
	// Test edge case - very high temperature
	client.SetTemperature(1.0)
	if client.Temperature != 1.0 {
		t.Errorf("Expected temperature 1.0, got %f", client.Temperature)
	}
}

// TestSendMessageWithTemperature tests that temperature is included in API requests
func TestSendMessageWithTemperature(t *testing.T) {
	// Create a mock HTTP server that validates temperature in requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Read and verify request body
		var requestBody OpenAIRequest
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&requestBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Check if temperature is set correctly
		if requestBody.Temperature != 0.3 {
			t.Errorf("Expected temperature 0.3, got %f", requestBody.Temperature)
		}

		// Send mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"id": "test-id",
			"object": "chat.completion",
			"created": 1677649420,
			"model": "gpt-4o",
			"choices": [
				{
					"index": 0,
					"message": {
						"role": "assistant",
						"content": "This is a test response"
					},
					"finish_reason": "stop"
				}
			],
			"usage": {
				"prompt_tokens": 10,
				"completion_tokens": 10,
				"total_tokens": 20
			}
		}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create client with mocked server URL
	client := NewClient("test-api-key")
	client.BaseURL = server.URL
	client.SetTemperature(0.3)

	// Send a test message
	response, err := client.SendMessage(context.Background(), models.Message{
		Role:    models.RoleUser,
		Content: "Hello, OpenAI!",
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