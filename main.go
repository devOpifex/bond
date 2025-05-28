package main

import (
	"context"
	"fmt"
	"os"

	"github/devOpifex/bond/agent"
	"github/devOpifex/bond/providers"
	"github/devOpifex/bond/tools"
)

func main() {
	// Create a provider using the factory
	provider, err := providers.NewProvider(providers.Claude, os.Getenv("ANTHROPIC_API_KEY"))
	if err != nil {
		fmt.Printf("Error creating provider: %v\n", err)
		return
	}

	// Configure provider
	provider.SetModel("claude-3-sonnet-20240229")
	provider.SetMaxTokens(1000)

	// Create agent manager
	agentManager := agent.NewAgentManager()
	
	// Register example agents
	agentManager.RegisterAgent("code-generation", &agent.SimpleAgent{Name: "CodeGen"})
	agentManager.RegisterAgent("data-analysis", &agent.SimpleAgent{Name: "DataAnalyst"})

	// Create tool factory and initialize built-in tools
	toolFactory := tools.NewFactory(agentManager)
	if err := toolFactory.CreateBuiltInTools(); err != nil {
		fmt.Printf("Error creating tools: %v\n", err)
		return
	}

	// Register all tools with the provider
	for _, tool := range toolFactory.Registry().GetAll() {
		provider.RegisterTool(tool)
	}

	// Example usage with tools
	ctx := context.Background()
	response, err := provider.SendMessageWithTools(ctx, "What's the weather in San Francisco and calculate 15 * 23?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("AI response: %s\n", response)

	// Example with Llama agent
	response, err = provider.SendMessageWithTools(ctx, "Generate a Python function to calculate fibonacci numbers using my code generation agent")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Code generation response: %s\n", response)
}