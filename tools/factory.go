package tools

import (
	"github/devOpifex/bond/models"
)

// Factory creates and initializes tools
type Factory struct {
	registry *Registry
}

// NewFactory creates a new tool factory
func NewFactory() *Factory {
	return &Factory{
		registry: NewRegistry(),
	}
}

// Registry returns the tool registry
func (f *Factory) Registry() *Registry {
	return f.registry
}

// CreateTool creates a new tool with the provided parameters and adds it to the registry
func (f *Factory) CreateTool(
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