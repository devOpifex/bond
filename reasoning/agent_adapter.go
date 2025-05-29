package reasoning

import (
	"context"
	"fmt"

	"github.com/devOpifex/bond/agent"
	"github.com/devOpifex/bond/models"
)

// WithAgent creates a reasoning step that uses an agent
func WithAgent(name string, description string, capability string, agentManager *agent.AgentManager) *Step {
	return &Step{
		Name:        name,
		Description: description,
		Execute: func(ctx context.Context, input string) (string, error) {
			return agentManager.ProcessWithBestAgent(ctx, capability, input)
		},
	}
}

// WithProvider creates a reasoning step that uses an AI provider directly
func WithProvider(name string, description string, promptTemplate string, provider models.Provider) *Step {
	return &Step{
		Name:        name,
		Description: description,
		Execute: func(ctx context.Context, input string) (string, error) {
			// Apply the template to the input
			prompt := fmt.Sprintf(promptTemplate, input)

			// Send to provider
			return provider.SendMessage(ctx, models.Message{
				Role:    models.RoleUser,
				Content: prompt,
			})
		},
	}
}

// Processor is a function that processes input data
type Processor func(ctx context.Context, input string) (string, error)

// WithProcessor creates a reasoning step that runs a custom processor function
func WithProcessor(name string, description string, processor Processor) *Step {
	return &Step{
		Name:        name,
		Description: description,
		Execute:     processor,
	}
}

