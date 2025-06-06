package main

import (
	"context"
	"fmt"
	"os"

	"github.com/devOpifex/bond/models"
	"github.com/devOpifex/bond/providers/claude"
	"github.com/devOpifex/bond/reasoning"
	"github.com/devOpifex/bond/tools"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: ANTHROPIC_API_KEY environment variable is not set")
		return
	}

	// Create a Claude provider
	provider := claude.NewClient(apiKey)

	// Create a React agent with tools
	reactAgent := reasoning.NewReactAgent(provider)
	
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
			expr, _ := params["expression"].(string)
			return fmt.Sprintf("Calculated result: %s = 42", expr), nil
		},
	)
	reactAgent.RegisterTool(calculator)

	// Create a chain that uses the React agent as a step
	chain := reasoning.NewChain()
	
	// Add preprocessing step
	chain.Add(reasoning.WithProcessor(
		"Preprocess Input",
		"Reformats the input for the agent",
		func(ctx context.Context, input string) (string, error) {
			return fmt.Sprintf("I need help with this question: %s", input), nil
		},
	)).
	// Add the React agent as a step
	Then(reactAgent.AsStep(
		"Solve Problem",
		"Uses a React agent with tools to solve the problem",
	)).
	// Add postprocessing step
	Then(reasoning.WithProcessor(
		"Format Output",
		"Formats the agent's response",
		func(ctx context.Context, input string) (string, error) {
			return fmt.Sprintf("Final answer: %s", input), nil
		},
	))

	// Execute the chain
	ctx := context.Background()
	result, err := chain.Execute(ctx, "What is 21 + 21?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(result)

	// The Claude provider doesn't expose a call count directly
	fmt.Println("Chain execution completed")
}