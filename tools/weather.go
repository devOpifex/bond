package tools

import (
	"encoding/json"
	"fmt"

	"github/devOpifex/bond/models"
)

// WeatherTool provides weather information for a location
type WeatherTool struct{}

func (w *WeatherTool) GetName() string {
	return "get_weather"
}

func (w *WeatherTool) GetDescription() string {
	return "Get current weather information for a location"
}

func (w *WeatherTool) GetSchema() models.InputSchema {
	return models.InputSchema{
		Type: "object",
		Properties: map[string]models.Property{
			"location": {
				Type:        "string",
				Description: "The city and state/country (e.g., 'San Francisco, CA')",
			},
		},
		Required: []string{"location"},
	}
}

func (w *WeatherTool) Execute(input json.RawMessage) (string, error) {
	var params struct {
		Location string `json:"location"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	// Simulate weather API call
	return fmt.Sprintf("The weather in %s is 72°F and sunny", params.Location), nil
}