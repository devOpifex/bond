package agent

import (
	"context"
	"testing"
)

// TestAgentManager tests the agent manager functionality
func TestAgentManager(t *testing.T) {
	// Create an agent manager
	manager := NewAgentManager()

	// Register test agents
	manager.RegisterAgent("test-agent", &TestAgent{Name: "TestAgent"})
	manager.RegisterAgent("math-agent", &TestAgent{Name: "MathAgent"})

	// Test processing with an agent
	result, err := manager.ProcessWithBestAgent(context.Background(), "test-agent", "Hello, world!")
	if err != nil {
		t.Fatalf("Failed to process with agent: %v", err)
	}

	expected := "Agent TestAgent processed: Hello, world!"
	if result != expected {
		t.Errorf("Expected result '%s', got '%s'", expected, result)
	}

	// Test with non-existent agent
	_, err = manager.ProcessWithBestAgent(context.Background(), "unknown-agent", "test")
	if err == nil {
		t.Error("Expected error for unknown agent, got nil")
	}
}

// TestAgent is a simple implementation of the Agent interface for testing
type TestAgent struct {
	Name string
}

// Process implements the Agent interface for testing
func (t *TestAgent) Process(ctx context.Context, input string) (string, error) {
	return "Agent " + t.Name + " processed: " + input, nil
}