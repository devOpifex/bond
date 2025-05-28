package tools

import (
	"github/devOpifex/bond/agent"
	"github/devOpifex/bond/models"
)

// Factory creates and initializes tools
type Factory struct {
	agentManager *agent.AgentManager
	registry     *Registry
}

// NewFactory creates a new tool factory
func NewFactory(agentManager *agent.AgentManager) *Factory {
	return &Factory{
		agentManager: agentManager,
		registry:     NewRegistry(),
	}
}

// Registry returns the tool registry
func (f *Factory) Registry() *Registry {
	return f.registry
}

// CreateBuiltInTools creates and registers all built-in tools
func (f *Factory) CreateBuiltInTools() error {
	// Create basic tools
	weatherTool := NewWeatherTool()
	calculatorTool := NewCalculatorTool()
	
	// Create agent-based tools if agent manager is available
	var llamaTool models.ToolExecutor
	if f.agentManager != nil {
		llamaTool = NewLlamaAgentTool(f.agentManager)
		if err := f.registry.Register(llamaTool); err != nil {
			return err
		}
	}
	
	// Register all tools
	if err := f.registry.Register(weatherTool); err != nil {
		return err
	}
	
	if err := f.registry.Register(calculatorTool); err != nil {
		return err
	}
	
	return nil
}

// CreateCustomTool creates a custom tool with the provided parameters
func (f *Factory) CreateCustomTool(
	name, description string,
	schema models.InputSchema,
	handler func(map[string]interface{}) (string, error),
) (*BaseTool, error) {
	tool := NewTool(name, description, schema, handler)
	if err := f.registry.Register(tool); err != nil {
		return nil, err
	}
	return tool, nil
}