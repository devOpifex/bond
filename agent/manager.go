package agent

import (
	"context"
	"fmt"
	
	"github.com/devOpifex/bond/models"
)

// Agent interface is now defined in models package
type Agent = models.Agent

// AgentManager manages different agents with different capabilities
type AgentManager struct {
	agents map[string]Agent
}

// NewAgentManager creates a new agent manager
func NewAgentManager() *AgentManager {
	return &AgentManager{
		agents: make(map[string]Agent),
	}
}

// RegisterAgent adds an agent with a specific capability to the manager
func (m *AgentManager) RegisterAgent(capability string, agent Agent) {
	m.agents[capability] = agent
}

// ProcessWithBestAgent selects and uses the appropriate agent for a given capability
func (m *AgentManager) ProcessWithBestAgent(ctx context.Context, capability, input string) (string, error) {
	agent, exists := m.agents[capability]
	if !exists {
		return "", fmt.Errorf("no agent found for capability: %s", capability)
	}
	return agent.Process(ctx, input)
}

// SimpleAgent is a basic implementation of the Agent interface
type SimpleAgent struct {
	Name string
}

// Process handles the input and returns a response
func (s *SimpleAgent) Process(ctx context.Context, input string) (string, error) {
	return fmt.Sprintf("Agent %s processed: %s", s.Name, input), nil
}