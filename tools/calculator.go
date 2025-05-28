package tools

import (
	"encoding/json"
	"fmt"

	"github/devOpifex/bond/models"
)

// CalculatorTool performs basic mathematical calculations
type CalculatorTool struct{}

func (c *CalculatorTool) GetName() string {
	return "calculator"
}

func (c *CalculatorTool) GetDescription() string {
	return "Perform basic mathematical calculations"
}

func (c *CalculatorTool) GetSchema() models.InputSchema {
	return models.InputSchema{
		Type: "object",
		Properties: map[string]models.Property{
			"expression": {
				Type:        "string",
				Description: "Mathematical expression to evaluate (e.g., '2 + 3 * 4')",
			},
		},
		Required: []string{"expression"},
	}
}

func (c *CalculatorTool) Execute(input json.RawMessage) (string, error) {
	var params struct {
		Expression string `json:"expression"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	// Simulate calculation (in real implementation, use a math parser)
	return fmt.Sprintf("Result of '%s' is 42", params.Expression), nil
}

