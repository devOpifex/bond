package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github/devOpifex/bond/models"
	"github/devOpifex/bond/providers"
)

// Tool implementations
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

// Custom agent tool that calls your Llama agents
type LlamaAgentTool struct {
	agentManager *AgentManager
}

func NewLlamaAgentTool(manager *AgentManager) *LlamaAgentTool {
	return &LlamaAgentTool{agentManager: manager}
}

func (l *LlamaAgentTool) GetName() string {
	return "call_llama_agent"
}

func (l *LlamaAgentTool) GetDescription() string {
	return "Call a specialized Llama agent for specific tasks like code generation or data analysis"
}

func (l *LlamaAgentTool) GetSchema() models.InputSchema {
	return models.InputSchema{
		Type: "object",
		Properties: map[string]models.Property{
			"capability": {
				Type:        "string",
				Description: "The capability needed (e.g., 'code-generation', 'data-analysis', 'chat')",
			},
			"query": {
				Type:        "string",
				Description: "The query or task to send to the agent",
			},
		},
		Required: []string{"capability", "query"},
	}
}

func (l *LlamaAgentTool) Execute(input json.RawMessage) (string, error) {
	var params struct {
		Capability string `json:"capability"`
		Query      string `json:"query"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	// Call your agent manager
	result, err := l.agentManager.ProcessWithBestAgent(
		context.Background(),
		params.Capability,
		params.Query,
	)
	if err != nil {
		return "", err
	}

	return result, nil
}

// AgentManager and related types (simplified version for this example)
type AgentManager struct {
	agents map[string]Agent
}

type Agent interface {
	Process(ctx context.Context, input string) (string, error)
}

type SimpleAgent struct {
	name string
}

func (s *SimpleAgent) Process(ctx context.Context, input string) (string, error) {
	return fmt.Sprintf("Agent %s processed: %s", s.name, input), nil
}

func NewAgentManager() *AgentManager {
	return &AgentManager{
		agents: make(map[string]Agent),
	}
}

func (m *AgentManager) RegisterAgent(capability string, agent Agent) {
	m.agents[capability] = agent
}

func (m *AgentManager) ProcessWithBestAgent(ctx context.Context, capability, input string) (string, error) {
	agent, exists := m.agents[capability]
	if !exists {
		return "", fmt.Errorf("no agent found for capability: %s", capability)
	}
	return agent.Process(ctx, input)
}

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

	// Register tools with provider
	provider.RegisterTool(&WeatherTool{})
	provider.RegisterTool(&CalculatorTool{})

	// Register custom Llama agent tool
	agentManager := NewAgentManager()
	// Register example agents
	agentManager.RegisterAgent("code-generation", &SimpleAgent{name: "CodeGen"})
	agentManager.RegisterAgent("data-analysis", &SimpleAgent{name: "DataAnalyst"})

	llamaTool := NewLlamaAgentTool(agentManager)
	provider.RegisterTool(llamaTool)

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

	// Example of using a different provider type (OpenAI)
	openaiProvider, err := providers.NewProvider(providers.OpenAI, "your-openai-key-here")
	if err != nil {
		fmt.Printf("Error creating OpenAI provider: %v\n", err)
		return
	}

	// Configure OpenAI provider
	openaiProvider.SetModel("gpt-4o")
	openaiProvider.SetMaxTokens(1000)

	// Register the same tools with OpenAI
	openaiProvider.RegisterTool(&WeatherTool{})
	openaiProvider.RegisterTool(&CalculatorTool{})
	openaiProvider.RegisterTool(llamaTool)

	// Use the OpenAI provider with the same query
	response, err = openaiProvider.SendMessageWithTools(ctx, "What's the weather in London and calculate 8 * 7?")
	if err != nil {
		fmt.Printf("Error with OpenAI: %v\n", err)
		return
	}

	fmt.Printf("OpenAI response: %s\n", response)
}
