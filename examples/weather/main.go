package main

import (
	"context"
	"fmt"
	"os"

	"github.com/devOpifex/bond/models"
	"github.com/devOpifex/bond/providers"
	"github.com/devOpifex/bond/tools"
)

func main() {
	// Create a Claude provider
	provider, err := providers.NewProvider(providers.Claude, os.Getenv("ANTHROPIC_API_KEY"))
	if err != nil {
		fmt.Printf("Error creating provider: %v\n", err)
		return
	}

	// Configure provider
	provider.SetModel("claude-3-sonnet-20240229")
	provider.SetMaxTokens(1000)

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
		func(params map[string]interface{}) (string, error) {
			location, _ := params["location"].(string)
			// In a real implementation, you would call a weather API here
			return fmt.Sprintf("The weather in %s is 72°F and sunny", location), nil
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
			Content: "What's the weather like in Boston, MA?",
		},
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Claude's response: %s\n", response)
}
