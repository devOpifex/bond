package tools

import (
	"fmt"

	"github/devOpifex/bond/models"
)

// NewWeatherTool creates a new weather tool
func NewWeatherTool() *BaseTool {
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