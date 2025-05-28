package tools

import (
	"fmt"

	"github/devOpifex/bond/models"
)

// NewCalculatorTool creates a new calculator tool
func NewCalculatorTool() *BaseTool {
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

