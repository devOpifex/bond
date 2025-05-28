package main

import (
	"context"
	"fmt"
	"os"

	"github.com/devOpifex/bond/agent"
	"github.com/devOpifex/bond/models"
	"github.com/devOpifex/bond/providers"
	"github.com/devOpifex/bond/tools"
)

// Custom agent that generates code
type CodeGenerator struct{}

func (c *CodeGenerator) Process(ctx context.Context, input string) (string, error) {
	// In a real implementation, this might use a specialized model or service
	return fmt.Sprintf("Here's the code you requested:\n\n```python\ndef fibonacci(n):\n    if n <= 1:\n        return n\n    return fibonacci(n-1) + fibonacci(n-2)\n```"), nil
}

func main() {
	// Create a Claude provider
	provider, err := providers.NewProvider(providers.Claude, os.Getenv("ANTHROPIC_API_KEY"))
	if err != nil {
		fmt.Printf("Error creating provider: %v\n", err)
		return
	}

	// Configure the provider
	provider.SetModel("claude-3-sonnet-20240229")
	provider.SetMaxTokens(1000)

	// Create an agent manager and register our code generator
	agentManager := agent.NewAgentManager()
	agentManager.RegisterAgent("code-generation", &CodeGenerator{})

	// Create a tool that uses the agent
	agentTool := tools.NewTool(
		"generate_code",
		"Generate code using a specialized agent",
		models.InputSchema{
			Type: "object",
			Properties: map[string]models.Property{
				"language": {
					Type:        "string",
					Description: "The programming language (e.g., 'python', 'javascript')",
				},
				"task": {
					Type:        "string",
					Description: "What you want the code to do",
				},
			},
			Required: []string{"language", "task"},
		},
		func(params map[string]interface{}) (string, error) {
			// In a real implementation, you might use the language parameter
			// to select a specific agent or pass it to the agent
			task, _ := params["task"].(string)

			return agentManager.ProcessWithBestAgent(
				context.Background(),
				"code-generation",
				task,
			)
		},
	)

	// Register the agent tool with the provider
	provider.RegisterTool(agentTool)

	// Send a message that will use the agent through the tool
	ctx := context.Background()
	response, err := provider.SendMessageWithTools(
		ctx,
		"Can you generate a Python function to calculate Fibonacci numbers?",
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Claude's response: %s\n", response)
}
