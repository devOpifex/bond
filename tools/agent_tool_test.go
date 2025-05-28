package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/devOpifex/bond/agent"
	"github.com/devOpifex/bond/models"
)

// TestAgentTool tests a tool that delegates to agents
func TestAgentTool(t *testing.T) {
	// Create an agent manager
	manager := agent.NewAgentManager()
	
	// Register test agents
	manager.RegisterAgent("code-generation", &testAgent{name: "CodeGen"})
	manager.RegisterAgent("data-analysis", &testAgent{name: "DataAnalyst"})
	
	// Create the agent tool
	agentTool := newLlamaAgentTool(manager)
	
	// Test the agent tool
	input := map[string]interface{}{
		"capability": "code-generation",
		"query":      "Write a fibonacci function",
	}
	
	inputJSON, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}
	
	result, err := agentTool.Execute(inputJSON)
	if err != nil {
		t.Fatalf("Failed to execute agent tool: %v", err)
	}
	
	expected := "Agent CodeGen processed: Write a fibonacci function"
	if result != expected {
		t.Errorf("Expected result '%s', got '%s'", expected, result)
	}
}

// testAgent is a simple agent implementation for testing
type testAgent struct {
	name string
}

// Process implements the agent.Agent interface
func (a *testAgent) Process(ctx context.Context, input string) (string, error) {
	return fmt.Sprintf("Agent %s processed: %s", a.name, input), nil
}

// newLlamaAgentTool creates a new agent tool for testing
func newLlamaAgentTool(manager *agent.AgentManager) *BaseTool {
	return NewTool(
		"call_agent",
		"Call a specialized agent for specific tasks",
		models.InputSchema{
			Type: "object",
			Properties: map[string]models.Property{
				"capability": {
					Type:        "string",
					Description: "The capability needed (e.g., 'code-generation', 'data-analysis')",
				},
				"query": {
					Type:        "string",
					Description: "The query or task to send to the agent",
				},
			},
			Required: []string{"capability", "query"},
		},
		func(params map[string]interface{}) (string, error) {
			capability, _ := params["capability"].(string)
			query, _ := params["query"].(string)
			
			// Call the agent manager
			return manager.ProcessWithBestAgent(
				context.Background(),
				capability,
				query,
			)
		},
	)
}