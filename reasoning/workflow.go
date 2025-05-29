// This file is kept for backward compatibility
// New code should use Chain from step.go instead

package reasoning

// Workflow represents a sequence of steps (deprecated, use Chain instead)
type Workflow = Chain

// NewWorkflow creates a new workflow (deprecated, use NewChain instead)
func NewWorkflow() *Workflow {
	return NewChain()
}

// AddStep adds a step to the workflow (deprecated, use Add instead)
func (w *Workflow) AddStep(step *Step) *Workflow {
	return w.Add(step)
}

