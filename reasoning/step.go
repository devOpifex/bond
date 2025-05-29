package reasoning

import (
	"context"
	"fmt"
)

// StepResult represents the result of a reasoning step
type StepResult struct {
	Output string
	Error  error
}

// Step defines a single reasoning step
type Step struct {
	Name        string
	Description string
	Execute     StepExecutor
}

// StepExecutor is a function that executes a single step
type StepExecutor = func(ctx context.Context, input string) (string, error)

// Chain represents a sequence of steps that can be executed in order
type Chain struct {
	steps []*Step
}

// NewChain creates a new reasoning chain
func NewChain() *Chain {
	return &Chain{
		steps: make([]*Step, 0),
	}
}

// Add adds a step to the chain and returns the chain for method chaining
func (c *Chain) Add(step *Step) *Chain {
	c.steps = append(c.steps, step)
	return c
}

// Execute runs all steps in the chain in sequence
func (c *Chain) Execute(ctx context.Context, input string) (string, error) {
	currentInput := input

	for _, step := range c.steps {
		output, err := step.Execute(ctx, currentInput)
		if err != nil {
			return "", fmt.Errorf("error executing step %s: %w", step.Name, err)
		}

		// Use this step's output as input to the next step
		currentInput = output
	}

	return currentInput, nil
}

// Then is an alias for Add to provide a fluent API
func (c *Chain) Then(step *Step) *Chain {
	return c.Add(step)
}