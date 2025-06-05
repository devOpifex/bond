// Package tools implements a framework for creating and managing tools that AI models can use.
package tools

import (
	"github.com/devOpifex/bond/models"
)

// CreateAndRegisterTool creates a new tool and registers it in the provided registry
func CreateAndRegisterTool(
	registry *Registry,
	name, description string,
	schema models.InputSchema,
	handler func(map[string]interface{}) (string, error),
) (*BaseTool, error) {
	tool := NewTool(name, description, schema, handler)
	if err := registry.Register(tool); err != nil {
		return nil, err
	}
	return tool, nil
}

