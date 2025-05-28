package reasoning

import (
	"context"
	"fmt"
)

// Workflow represents a directed graph of steps
type Workflow struct {
	steps  map[string]*Step
	memory *Memory
}

// NewWorkflow creates a new workflow
func NewWorkflow() *Workflow {
	return &Workflow{
		steps:  make(map[string]*Step),
		memory: NewMemory(),
	}
}

// AddStep adds a step to the workflow
func (w *Workflow) AddStep(step *Step) error {
	if _, exists := w.steps[step.ID]; exists {
		return fmt.Errorf("step with ID %s already exists", step.ID)
	}
	w.steps[step.ID] = step
	return nil
}

// Execute runs the workflow, ensuring steps are executed in the correct order
func (w *Workflow) Execute(ctx context.Context, input string, entryPointID string) (string, error) {
	// Validate entry point
	if _, exists := w.steps[entryPointID]; !exists {
		return "", fmt.Errorf("entry point step %s not found", entryPointID)
	}

	// Store the input for reference
	w.memory.Set("workflow.input", input)

	// Execute the workflow starting from the entry point
	result, err := w.executeStep(ctx, entryPointID, input, make(map[string]bool))
	if err != nil {
		return "", err
	}

	return result, nil
}

// executeStep executes a single step and its dependencies
func (w *Workflow) executeStep(ctx context.Context, stepID string, input string, visited map[string]bool) (string, error) {
	// Check for cycles
	if visited[stepID] {
		return "", fmt.Errorf("cycle detected in workflow at step %s", stepID)
	}
	
	// Create a copy of the visited map for this execution path
	// This prevents false cycle detection in different branches
	visitedCopy := make(map[string]bool)
	for k, v := range visited {
		visitedCopy[k] = v
	}
	visitedCopy[stepID] = true

	step, exists := w.steps[stepID]
	if !exists {
		return "", fmt.Errorf("step %s not found", stepID)
	}

	// Check if we've already executed this step
	if output, ok := w.memory.GetString(fmt.Sprintf("step.%s.output", stepID)); ok {
		return output, nil
	}

	// Execute dependencies first
	for _, depID := range step.DependsOn {
		// Use original input for dependencies unless specified otherwise
		depInput := input
		if storedInput, ok := w.memory.GetString(fmt.Sprintf("dependency.%s.input", depID)); ok {
			depInput = storedInput
		}

		_, err := w.executeStep(ctx, depID, depInput, visitedCopy)
		if err != nil {
			return "", err
		}
	}

	// Execute the step
	stepResult, err := step.Execute(ctx, input, w.memory)
	if err != nil {
		return "", fmt.Errorf("error executing step %s: %w", stepID, err)
	}

	// Store results in memory
	w.memory.Set(fmt.Sprintf("step.%s.output", stepID), stepResult.Output)
	for k, v := range stepResult.Metadata {
		w.memory.Set(fmt.Sprintf("step.%s.%s", stepID, k), v)
	}

	return stepResult.Output, nil
}

// GetStepOutput gets the output of a specific step
func (w *Workflow) GetStepOutput(stepID string) (string, bool) {
	return w.memory.GetString(fmt.Sprintf("step.%s.output", stepID))
}

// GetMemory returns the workflow's memory store
func (w *Workflow) GetMemory() *Memory {
	return w.memory
}