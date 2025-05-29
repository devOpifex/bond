package reasoning

import (
	"context"
	"testing"
)

func TestChain(t *testing.T) {
	// Create a chain with a few simple steps
	chain := NewChain()
	
	// Add a step that appends " World"
	chain.Add(&Step{
		Name:        "Append World",
		Description: "Appends ' World' to the input",
		Execute: func(ctx context.Context, input string) (string, error) {
			return input + " World", nil
		},
	})
	
	// Add a step that appends "!"
	chain.Then(&Step{
		Name:        "Append Exclamation",
		Description: "Appends '!' to the input",
		Execute: func(ctx context.Context, input string) (string, error) {
			return input + "!", nil
		},
	})
	
	// Execute the chain
	result, err := chain.Execute(context.Background(), "Hello")
	if err != nil {
		t.Fatalf("Chain execution failed: %v", err)
	}
	
	// Check the result
	expected := "Hello World!"
	if result != expected {
		t.Errorf("Expected result to be '%s', got '%s'", expected, result)
	}
}

func TestChainWithProcessor(t *testing.T) {
	// Create a chain with processor-based steps
	chain := NewChain()
	
	// Add steps using the WithProcessor helper
	chain.Add(WithProcessor(
		"Uppercase", 
		"Converts text to uppercase",
		func(ctx context.Context, input string) (string, error) {
			return "HELLO", nil
		},
	)).
	Then(WithProcessor(
		"Add Greeting",
		"Adds a greeting",
		func(ctx context.Context, input string) (string, error) {
			return input + " WORLD", nil
		},
	))
	
	// Execute the chain
	result, err := chain.Execute(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Chain execution failed: %v", err)
	}
	
	// Check the result
	expected := "HELLO WORLD"
	if result != expected {
		t.Errorf("Expected result to be '%s', got '%s'", expected, result)
	}
}

func TestWorkflowCompatibility(t *testing.T) {
	// Test that the Workflow type alias works correctly
	workflow := NewWorkflow()
	
	// Add a step using the old API
	workflow.AddStep(&Step{
		Name:        "Test Step",
		Description: "A test step",
		Execute: func(ctx context.Context, input string) (string, error) {
			return input + " processed", nil
		},
	})
	
	// Execute the workflow
	result, err := workflow.Execute(context.Background(), "input")
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}
	
	// Check the result
	expected := "input processed"
	if result != expected {
		t.Errorf("Expected result to be '%s', got '%s'", expected, result)
	}
}