# Bond

Bond is a Go package for building and managing AI assistant tools and agents, with a focus on extensibility and ease of use.

## Overview

Bond provides a flexible framework for:

1. Creating custom tools for AI assistants
2. Managing specialized agents for different tasks
3. Connecting with AI providers like Claude and OpenAI

## Packages

- **models**: Core interfaces and types used throughout the project
- **tools**: Framework for creating, managing, and executing tools
- **agent**: Framework for building specialized agents for different capabilities
- **providers**: Integration with AI providers (Claude, OpenAI)

## Installation

```bash
go get github.com/devOpifex/bond
```

## Basic Example

Here's a simple example showing how to create a custom tool, register it with Claude, and use it:

```go
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
					Description: "The city and state or city/country (e.g., 'San Francisco, CA')",
				},
			},
			Required: []string{"location"},
		},
		func(params map[string]interface{}) (string, error) {
			location, _ := params["location"].(string)
			// In a real implementation, you would call a weather API here
			return fmt.Sprintf("The weather in %s is 10°C and raining as usual", location), nil
		},
	)

	// Register the tool with the provider
	provider.RegisterTool(weatherTool)

	// Send a message that will use the tool
	ctx := context.Background()
	response, err := provider.SendMessageWithTools(
		ctx,
		"What's the weather like in Brussesl, Belgium?",
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Claude's response: %s\n", response)
}
```

## Creating Custom Agents

You can also create custom agents and use them through tools:

```go
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
```
