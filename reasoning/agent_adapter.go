package reasoning

import (
	"context"
	"fmt"

	"github.com/devOpifex/bond/agent"
	"github.com/devOpifex/bond/models"
)

// AgentStep creates a reasoning step that uses an agent
func AgentStep(id string, name string, description string, capability string, agentManager *agent.AgentManager) *Step {
	return &Step{
		ID:          id,
		Name:        name,
		Description: description,
		Execute: func(ctx context.Context, input string, memory *Memory) (StepResult, error) {
			output, err := agentManager.ProcessWithBestAgent(ctx, capability, input)
			if err != nil {
				return StepResult{Error: err}, err
			}

			return StepResult{
				Output: output,
				Metadata: map[string]interface{}{
					"capability": capability,
				},
			}, nil
		},
	}
}

// ProviderStep creates a reasoning step that uses an AI provider directly
func ProviderStep(id string, name string, description string, promptTemplate string, provider models.Provider) *Step {
	return &Step{
		ID:          id,
		Name:        name,
		Description: description,
		Execute: func(ctx context.Context, input string, memory *Memory) (StepResult, error) {
			// Apply the template to the input
			prompt := fmt.Sprintf(promptTemplate, input)
			
			// Send to provider
			output, err := provider.SendMessage(ctx, models.Message{
				Role:    models.RoleUser,
				Content: prompt,
			})
			if err != nil {
				return StepResult{Error: err}, err
			}

			return StepResult{
				Output: output,
				Metadata: map[string]interface{}{
					"prompt": prompt,
				},
			}, nil
		},
	}
}

// MemoryProcessor is a function that processes data from memory
type MemoryProcessor func(ctx context.Context, input string, memory *Memory) (string, map[string]interface{}, error)

// ProcessorStep creates a reasoning step that runs a custom processor function
func ProcessorStep(id string, name string, description string, processor MemoryProcessor) *Step {
	return &Step{
		ID:          id,
		Name:        name,
		Description: description,
		Execute: func(ctx context.Context, input string, memory *Memory) (StepResult, error) {
			output, metadata, err := processor(ctx, input, memory)
			if err != nil {
				return StepResult{Error: err}, err
			}

			return StepResult{
				Output:   output,
				Metadata: metadata,
			}, nil
		},
	}
}