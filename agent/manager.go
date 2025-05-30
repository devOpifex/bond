// Package agent provides functionality for managing multiple specialized AI agents.
// It handles the coordination and routing of requests to the appropriate agents
// based on their capabilities. This package sits at the top of the Bond framework's
// architecture, using the lower-level reasoning patterns and provider integrations.
package agent

import (
	"context"
	"fmt"
	
	"github.com/devOpifex/bond/models"
)

// Agent interface is now defined in models package
type Agent = models.Agent

// AgentManager coordinates multiple specialized agents with different capabilities.
// It maintains a registry of agents and routes requests to the appropriate agent
// based on the required capability.
type AgentManager struct {
	// agents maps capability names to agent implementations
	agents map[string]Agent
}

// NewAgentManager creates a new agent manager with an empty agent registry.
// Use RegisterAgent to add agents with specific capabilities.
func NewAgentManager() *AgentManager {
	return &AgentManager{
		agents: make(map[string]Agent),
	}
}

// RegisterAgent adds an agent with a specific capability to the manager.
// The capability string acts as a key to look up the appropriate agent.
func (m *AgentManager) RegisterAgent(capability string, agent Agent) {
	m.agents[capability] = agent
}

// ProcessWithBestAgent selects and uses the appropriate agent for a given capability.
// It returns an error if no agent is registered for the requested capability.
func (m *AgentManager) ProcessWithBestAgent(ctx context.Context, capability, input string) (string, error) {
	agent, exists := m.agents[capability]
	if !exists {
		return "", fmt.Errorf("no agent found for capability: %s", capability)
	}
	return agent.Process(ctx, input)
}

// SimpleAgent is a basic implementation of the Agent interface for testing and demonstration.
// It simply echoes back the input with the agent's name.
type SimpleAgent struct {
	// Name identifies this agent instance
	Name string
}

// Process implements the Agent interface for SimpleAgent.
// It returns a formatted string containing the agent name and the input.
func (s *SimpleAgent) Process(ctx context.Context, input string) (string, error) {
	return fmt.Sprintf("Agent %s processed: %s", s.Name, input), nil
}