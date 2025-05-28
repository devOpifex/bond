package tools

import (
	"fmt"
	"testing"

	"github/devOpifex/bond/models"
)

// TestTools contains sample tool implementations for testing
func TestTools(t *testing.T) {
	// Create a registry
	registry := NewRegistry()

	// Create and register test tools
	weatherTool := newWeatherTool()
	calculatorTool := newCalculatorTool()

	// Register tools
	if err := registry.Register(weatherTool); err != nil {
		t.Fatalf("Failed to register weather tool: %v", err)
	}

	if err := registry.Register(calculatorTool); err != nil {
		t.Fatalf("Failed to register calculator tool: %v", err)
	}

	// Verify tools can be retrieved
	tool, exists := registry.Get("get_weather")
	if !exists {
		t.Fatal("Weather tool not found in registry")
	}

	if tool.GetName() != "get_weather" {
		t.Errorf("Expected tool name 'get_weather', got '%s'", tool.GetName())
	}
}

// newWeatherTool creates a weather tool for testing
func newWeatherTool() *BaseTool {
	return NewTool(
		"get_weather",
		"Get current weather information for a location",
		models.InputSchema{
			Type: "object",
			Properties: map[string]models.Property{
				"location": {
					Type:        "string",
					Description: "The city and state/country (e.g., 'San Francisco, CA')",
				},
			},
			Required: []string{"location"},
		},
		func(params map[string]interface{}) (string, error) {
			location, _ := params["location"].(string)
			// Simulate weather API call
			return fmt.Sprintf("The weather in %s is 72°F and sunny", location), nil
		},
	)
}

// newCalculatorTool creates a calculator tool for testing
func newCalculatorTool() *BaseTool {
	return NewTool(
		"calculator",
		"Perform basic mathematical calculations",
		models.InputSchema{
			Type: "object",
			Properties: map[string]models.Property{
				"expression": {
					Type:        "string",
					Description: "Mathematical expression to evaluate (e.g., '2 + 3 * 4')",
				},
			},
			Required: []string{"expression"},
		},
		func(params map[string]interface{}) (string, error) {
			expression, _ := params["expression"].(string)
			// Simulate calculation (in real implementation, use a math parser)
			return fmt.Sprintf("Result of '%s' is 42", expression), nil
		},
	)
}

