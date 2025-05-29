# Bond

Bond is a Go package for building and managing AI assistant tools and agents, with a focus on extensibility and ease of use.

## Overview

Bond provides a flexible framework for:

1. Creating custom tools for AI assistants
2. Managing specialized agents for different tasks
3. Connecting with AI providers like Claude and OpenAI
4. Orchestrating complex multi-step reasoning workflows

## Packages

- **models**: Core interfaces and types used throughout the project
- **tools**: Framework for creating, managing, and executing tools
- **agent**: Framework for building specialized agents for different capabilities
- **providers**: Integration with AI providers (Claude, OpenAI)
- **reasoning**: Multi-step reasoning with state management and workflow orchestration

## Installation

```bash
go get github.com/devOpifex/bond
```

## Key Concepts

### Messages and Roles

Bond uses a structured approach to messages with different roles:

```go
// Create a message with a specific role
message := models.Message{
    Role:    models.RoleUser,     // Use predefined role constants
    Content: "Your message here",
}
```

Available role constants:
- `models.RoleUser` - For messages from the user
- `models.RoleAssistant` - For messages from the AI assistant
- `models.RoleSystem` - For system instructions (Claude uses a separate field)
- `models.RoleFunction` - For function/tool responses

### System Prompts

You can set a system prompt to guide the AI's behavior across all messages:

```go
// Set a system prompt to guide the AI's behavior
provider.SetSystemPrompt("You are a helpful assistant that specializes in data analysis.")
```

The system prompt provides context and instructions that persist across the conversation.

## Basic Example

Here's a simple example showing how to create a custom tool, register it with Claude, and use it:

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/devOpifex/bond/models"
	"github.com/devOpifex/bond/providers/claude"
	"github.com/devOpifex/bond/tools"
)

func main() {
	// Create a Claude provider
	provider := claude.NewClient(os.Getenv("ANTHROPIC_API_KEY"))

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

	// Set a system prompt to guide the AI's behavior
	provider.SetSystemPrompt("You are a weather assistant. Always be concise and factual.")

	// Send a message that will use the tool
	ctx := context.Background()
	response, err := provider.SendMessageWithTools(
		ctx,
		models.Message{
			Role:    models.RoleUser,
			Content: "What's the weather like in Brussels, Belgium?",
		},
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
	"github.com/devOpifex/bond/providers/claude"
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
	provider := claude.NewClient(os.Getenv("ANTHROPIC_API_KEY"))

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

	// Set a system prompt to guide the AI's behavior
	provider.SetSystemPrompt("You are a coding assistant. Focus on providing clear, well-commented code.")

	// Send a message that will use the agent through the tool
	ctx := context.Background()
	response, err := provider.SendMessageWithTools(
		ctx, 
		models.Message{
			Role:    models.RoleUser,
			Content: "Can you generate a Python function to calculate Fibonacci numbers?",
		},
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Claude's response: %s\n", response)
}
```

## ReAct Agent

Bond provides a ReAct (Reasoning + Acting) agent implementation that combines reasoning with tool use in an iterative process:

```go
// Create a provider
provider := claude.NewClient(os.Getenv("ANTHROPIC_API_KEY"))

// Create a ReAct agent
reactAgent := reasoning.NewReActAgent(provider)

// Configure the agent (optional)
reactAgent.SetMaxIterations(5)  // Default is 10
reactAgent.SetSystemPrompt("Custom system prompt")  // Override default prompt

// Register tools with the agent
reactAgent.RegisterTool(myTool)

// Process a query using the ReAct agent
ctx := context.Background()
result, err := reactAgent.Process(ctx, "What is 21 + 21?")
```

The ReAct agent:
- Maintains an internal conversation state
- Parses tool usage from model responses
- Executes tools and feeds results back to the model
- Continues iterations until a final answer is reached
- Supports easy integration with the Chain API

### Using ReAct in Chains

The ReAct agent can be integrated into reasoning chains:

```go
// Create a chain
chain := reasoning.NewChain()

// Add preprocessing step
chain.Add(reasoning.WithProcessor(
    "Preprocess Input",
    "Reformats the input for the agent",
    func(ctx context.Context, input string) (string, error) {
        return fmt.Sprintf("I need help with this question: %s", input), nil
    },
))

// Add the ReAct agent as a step
chain.Then(reactAgent.AsStep(
    "Solve Problem",
    "Uses a ReAct agent with tools to solve the problem",
))

// Add postprocessing step
chain.Then(reasoning.WithProcessor(
    "Format Output",
    "Formats the agent's response",
    func(ctx context.Context, input string) (string, error) {
        return fmt.Sprintf("Final answer: %s", input), nil
    },
))

// Execute the chain
result, err := chain.Execute(ctx, "What is 21 + 21?")
```

See the `examples/react/main.go` and `examples/react_in_chain/main.go` for complete examples.

## Simplified Reasoning Chains

Bond's reasoning framework has been refactored to provide a simpler Chain API that replaces the older Workflow approach:

```go
// Create a chain
chain := reasoning.NewChain()

// Add steps with a fluent API
chain.Add(myFirstStep).
     Then(mySecondStep).
     Then(myThirdStep)

// Execute the chain
result, err := chain.Execute(ctx, "User input")
```

The simplified Chain API:
- Provides a more intuitive sequential execution model
- Eliminates complex dependency management
- Simplifies state management between steps
- Offers a fluent interface for better readability
- Seamlessly integrates with ReAct agents

Each step in a chain takes the output of the previous step as its input, creating a clean data flow throughout the execution process.

## Multi-Step Reasoning

For more complex tasks, Bond provides a multi-step reasoning framework in the `reasoning` package. This allows you to break down complex tasks into a series of steps with state management:

```go
// Create a workflow
workflow := reasoning.NewWorkflow()

// Add steps with dependencies
workflow.AddStep(reasoning.ProcessorStep(
    "extract-info",
    "Extract information from input",
    "Processes the input to extract key information",
    func(ctx context.Context, input string, memory *reasoning.Memory) (string, map[string]interface{}, error) {
        // Extract and process information
        return "Processed result", map[string]interface{}{"key": "value"}, nil
    },
))

// Add a step that depends on the previous step
workflow.AddStep(reasoning.AgentStep(
    "process-data",
    "Process data",
    "Uses an agent to process the extracted data",
    "agent-capability",
    agentManager,
))

// Execute the workflow
result, err := workflow.Execute(ctx, userInput, "process-data")
```

The reasoning package provides:

- **Memory management**: Share data between steps
- **Dependency resolution**: Steps execute only after their dependencies
- **Agent integration**: Use specialized agents in workflow steps
- **Provider integration**: Direct access to AI providers in steps
- **Custom processors**: Implement custom logic in any step

See the `examples/code/main.go` for a complete example of using multi-step reasoning for a complex code generation and analysis task.
