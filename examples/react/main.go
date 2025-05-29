package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/devOpifex/bond/models"
	"github.com/devOpifex/bond/providers/claude"
	"github.com/devOpifex/bond/reasoning"
	"github.com/devOpifex/bond/tools"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	// Create a Claude provider
	provider := claude.NewClient(apiKey)

	// Set system prompt and model if needed
	provider.SetSystemPrompt(`You are a helpful AI assistant that can use tools. 
When asked a question that requires calculation, use the calculator tool.
You MUST format your tool calls exactly as follows:

<thought>
Your reasoning here...
</thought>

` + "```json" + `
{
  "name": "calculator",
  "input": "expression to calculate"
}
` + "```" + `

This exact format is critical for the tool parser to work properly.`)
	provider.SetMaxTokens(1000)

	// Create a ReAct agent
	reactAgent := reasoning.NewReActAgent(provider)
	
	// Set a simplified system prompt that helps the model understand context
	reactAgent.SetSystemPrompt(`You are a helpful assistant that can use tools to solve problems.
When using tools, please follow this format:

<thought>
Your reasoning here...
</thought>

` + "```json" + `
{
  "name": "toolName",
  "input": {"param": "value"}
}
` + "```" + `

After receiving tool results, provide a concise final answer to the user's question.`)

	// Register calculator tool
	calculator := tools.NewTool(
		"calculator",
		"Perform arithmetic calculations",
		models.InputSchema{
			Type: "object",
			Properties: map[string]models.Property{
				"expression": {
					Type:        "string",
					Description: "The mathematical expression to evaluate",
				},
			},
			Required: []string{"expression"},
		},
		func(params map[string]interface{}) (string, error) {
			// Simple calculator that actually evaluates the expression
			expr, ok := params["expression"].(string)
			if !ok {
				return "", fmt.Errorf("expression must be a string")
			}
			
			// Handle simple addition
			if strings.Contains(expr, "+") {
				parts := strings.Split(expr, "+")
				if len(parts) != 2 {
					return "", fmt.Errorf("can only handle simple addition with two numbers")
				}
				
				a, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
				if err != nil {
					return "", fmt.Errorf("invalid first number: %v", err)
				}
				
				b, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
				if err != nil {
					return "", fmt.Errorf("invalid second number: %v", err)
				}
				
				result := a + b
				return fmt.Sprintf("The result of %v + %v = %v", a, b, result), nil
			}
			
			return fmt.Sprintf("Sorry, I can only handle addition for now."), nil
		},
	)

	reactAgent.RegisterTool(calculator)
	reactAgent.SetMaxIterations(3) // Limit iterations to prevent infinite loops

	// Process a query using the ReAct agent
	fmt.Println("Asking: What is 21 + 21?")
	ctx := context.Background()
	
	// Process the query
	result, err := reactAgent.Process(ctx, "What is 21 + 21?")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Result:", result)
	
	// Try a more complex query
	fmt.Println("\nAsking: What is 13 + 29?")
	result, err = reactAgent.Process(ctx, "What is 13 + 29?")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	
	fmt.Println("Result:", result)
	
	// Try a question with different expression
	fmt.Println("\nAsking: Can you calculate 100 + 50?")
	result, err = reactAgent.Process(ctx, "Can you calculate 100 + 50?")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	
	fmt.Println("Result:", result)
}
