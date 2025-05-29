package reasoning

import (
	"context"
	"fmt"
)

// Workflow represents a sequence of steps with automatic dependency resolution
type Workflow struct {
	steps  []*Step
	memory *Memory
}

// NewWorkflow creates a new workflow
func NewWorkflow() *Workflow {
	return &Workflow{
		steps:  make([]*Step, 0),
		memory: NewMemory(),
	}
}

// AddStep adds a step to the workflow and returns the workflow for method chaining
func (w *Workflow) AddStep(step *Step) *Workflow {
	w.steps = append(w.steps, step)
	return w
}

// Execute runs the workflow in sequential order
func (w *Workflow) Execute(ctx context.Context, input string) (string, error) {
	// Store the input for reference
	w.memory.Set("workflow.input", input)

	var result string
	currentInput := input

	// Execute steps in sequential order
	for i, step := range w.steps {
		// Generate an ID if not already set
		if step.id == "" {
			step.id = fmt.Sprintf("step_%d", i)
		}
		
		stepResult, err := step.Execute(ctx, currentInput, w.memory)
		if err != nil {
			return "", fmt.Errorf("error executing step %s: %w", step.Name, err)
		}

		// Store results in memory
		w.memory.Set(fmt.Sprintf("step.%s.output", step.id), stepResult.Output)
		for k, v := range stepResult.Metadata {
			w.memory.Set(fmt.Sprintf("step.%s.%s", step.id, k), v)
		}

		// Use this step's output as input to the next step
		currentInput = stepResult.Output
		result = stepResult.Output
	}

	return result, nil
}

// GetStepOutput gets the output of a specific step by index
func (w *Workflow) GetStepOutput(index int) (string, bool) {
	if index < 0 || index >= len(w.steps) {
		return "", false
	}
	return w.memory.GetString(fmt.Sprintf("step.%s.output", w.steps[index].id))
}

// GetMemory returns the workflow's memory store
func (w *Workflow) GetMemory() *Memory {
	return w.memory
}

// Then is an alias for AddStep to provide a more fluent API
func (w *Workflow) Then(step *Step) *Workflow {
	return w.AddStep(step)
}