package reasoning

import (
	"context"
	"fmt"
)

// StepResult represents the result of a reasoning step
type StepResult struct {
	Output   string
	Metadata map[string]interface{}
	Error    error
}

// Step defines a single reasoning step
type Step struct {
	id          string // Internal use only
	Name        string
	Description string
	Execute     StepExecutor
}

// StepExecutor is a function that executes a single step
type StepExecutor func(ctx context.Context, input string, memory *Memory) (StepResult, error)

// Chain represents a sequence of steps that can be executed in order
type Chain struct {
	steps  []*Step
	memory *Memory
}

// NewChain creates a new reasoning chain
func NewChain() *Chain {
	return &Chain{
		steps:  make([]*Step, 0),
		memory: NewMemory(),
	}
}

// AddStep adds a step to the chain and returns the chain for method chaining
func (c *Chain) AddStep(step *Step) *Chain {
	c.steps = append(c.steps, step)
	return c
}

// Execute runs all steps in the chain in sequence
func (c *Chain) Execute(ctx context.Context, input string) (string, error) {
	var result string
	currentInput := input

	for i, step := range c.steps {
		// Generate an ID if not already set
		if step.id == "" {
			step.id = fmt.Sprintf("step_%d", i)
		}

		stepResult, err := step.Execute(ctx, currentInput, c.memory)
		if err != nil {
			return "", fmt.Errorf("error executing step %s: %w", step.Name, err)
		}

		// Store results in memory
		c.memory.Set(fmt.Sprintf("step.%s.output", step.id), stepResult.Output)
		for k, v := range stepResult.Metadata {
			c.memory.Set(fmt.Sprintf("step.%s.%s", step.id, k), v)
		}

		// Use this step's output as input to the next step
		currentInput = stepResult.Output
		result = stepResult.Output
	}

	return result, nil
}

// GetMemory returns the chain's memory store
func (c *Chain) GetMemory() *Memory {
	return c.memory
}

// Then is an alias for AddStep to provide a more fluent API
func (c *Chain) Then(step *Step) *Chain {
	return c.AddStep(step)
}