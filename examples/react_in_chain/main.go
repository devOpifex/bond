package main

import (
	"context"
	"fmt"

	"github.com/devOpifex/bond/models"
	"github.com/devOpifex/bond/reasoning"
	"github.com/devOpifex/bond/tools"
)

// MockProvider simulates a provider for demonstration purposes
type MockProvider struct {
	systemPrompt string
	tools        map[string]models.ToolExecutor
	callCount    int
}

func NewMockProvider() *MockProvider {
	return &MockProvider{
		tools:     make(map[string]models.ToolExecutor),
		callCount: 0,
	}
}

func (m *MockProvider) SetSystemPrompt(prompt string) {
	m.systemPrompt = prompt
}

func (m *MockProvider) SetModel(model string) {}

func (m *MockProvider) SetMaxTokens(tokens int) {}

func (m *MockProvider) RegisterTool(tool models.ToolExecutor) {
	m.tools[tool.GetName()] = tool
}

func (m *MockProvider) SendMessage(ctx context.Context, message models.Message) (string, error) {
	return "This is a mock response", nil
}

func (m *MockProvider) SendMessageWithTools(ctx context.Context, message models.Message) (string, error) {
	m.callCount++

	// First call - simulate reasoning and tool use for calculator
	if m.callCount == 1 {
		return `<thought>
I need to calculate what 21 + 21 equals. I can use the calculator tool for this.
</thought>

` + "```json" + `
{
  "name": "calculator",
  "input": "{\"expression\": \"21 + 21\"}"
}
` + "```", nil
	}

	// Second call - simulate final response
	return "Based on the calculator's result, 21 + 21 = 42.", nil
}

func main() {
	// Create a mock provider
	provider := NewMockProvider()

	// Create a ReAct agent with tools
	reactAgent := reasoning.NewReActAgent(provider)
	
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

	// Create a chain that uses the ReAct agent as a step
	chain := reasoning.NewChain()
	
	// Add preprocessing step
	chain.Add(reasoning.WithProcessor(
		"Preprocess Input",
		"Reformats the input for the agent",
		func(ctx context.Context, input string) (string, error) {
			return fmt.Sprintf("I need help with this question: %s", input), nil
		},
	)).
	// Add the ReAct agent as a step
	Then(reactAgent.AsStep(
		"Solve Problem",
		"Uses a ReAct agent with tools to solve the problem",
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

	// Print the message trace statistics for analysis
	fmt.Printf("Total provider calls: %d\n", provider.callCount)
}