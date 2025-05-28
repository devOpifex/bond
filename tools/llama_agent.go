package tools

import (
	"context"

	"github/devOpifex/bond/agent"
	"github/devOpifex/bond/models"
)

// NewLlamaAgentTool creates a new Llama agent tool that delegates to specialized agents
func NewLlamaAgentTool(manager *agent.AgentManager) *BaseTool {
	return NewTool(
		"call_llama_agent",
		"Call a specialized Llama agent for specific tasks like code generation or data analysis",
		models.InputSchema{
			Type: "object",
			Properties: map[string]models.Property{
				"capability": {
					Type:        "string",
					Description: "The capability needed (e.g., 'code-generation', 'data-analysis', 'chat')",
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

