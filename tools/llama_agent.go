package tools

import (
	"context"
	"encoding/json"

	"github/devOpifex/bond/agent"
	"github/devOpifex/bond/models"
)

// LlamaAgentTool is a tool that delegates tasks to specialized Llama agents
type LlamaAgentTool struct {
	agentManager *agent.AgentManager
}

// NewLlamaAgentTool creates a new LlamaAgentTool with the provided agent manager
func NewLlamaAgentTool(manager *agent.AgentManager) *LlamaAgentTool {
	return &LlamaAgentTool{agentManager: manager}
}

// GetName returns the name of the tool
func (l *LlamaAgentTool) GetName() string {
	return "call_llama_agent"
}

// GetDescription returns the description of the tool
func (l *LlamaAgentTool) GetDescription() string {
	return "Call a specialized Llama agent for specific tasks like code generation or data analysis"
}

// GetSchema returns the schema for the tool's input
func (l *LlamaAgentTool) GetSchema() models.InputSchema {
	return models.InputSchema{
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
	}
}

// Execute processes the input and returns the result from the appropriate agent
func (l *LlamaAgentTool) Execute(input json.RawMessage) (string, error) {
	var params struct {
		Capability string `json:"capability"`
		Query      string `json:"query"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	// Call your agent manager
	result, err := l.agentManager.ProcessWithBestAgent(
		context.Background(),
		params.Capability,
		params.Query,
	)
	if err != nil {
		return "", err
	}

	return result, nil
}

