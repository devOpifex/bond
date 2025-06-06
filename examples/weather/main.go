package main

import (
	"context"
	"fmt"
	"os"

	"github.com/devOpifex/bond/models"
	"github.com/devOpifex/bond/providers/claude"
	"github.com/devOpifex/bond/tools"
)

func main() {
	// Create a Claude provider
	provider := claude.NewClient(os.Getenv("ANTHROPIC_API_KEY"))

	// Configure provider
	provider.SetModel("claude-3-sonnet-20240229")
	provider.SetMaxTokens(1000)
	provider.SetSystemPrompt("You are a weather assistant. Always answer questions about weather concisely. When someone asks about weather in a location, use the get_weather tool.")
	provider.SetTemperature(0.2)

	// Create a weather tool
	weatherTool := tools.NewTool(
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
		func(params map[string]any) (string, error) {
			fmt.Println("Weather tool called with params:", params)
			location, _ := params["location"].(string)
			// In a real implementation, you would call a weather API here
			result := fmt.Sprintf("The weather in %s is 5°C and rainy as usual", location)
			fmt.Println("Weather tool returning:", result)
			return result, nil
		},
	)

	// Register the tool with the provider
	provider.RegisterTool(weatherTool)

	// Send a message that will use the tool
	ctx := context.Background()
	response, err := provider.SendMessageWithTools(
		ctx,
		models.Message{
			Role:    models.RoleUser,
			Content: "What's the weather like in Brussels, Belgium?",
		},
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Claude's response: %s\n", response)
}
