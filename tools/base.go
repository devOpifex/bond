package tools

import (
	"encoding/json"
	"errors"

	"github/devOpifex/bond/models"
)

// BaseTool provides common functionality for tools
type BaseTool struct {
	Name        string
	Description string
	Schema      models.InputSchema
	Handler     func(params map[string]interface{}) (string, error)
}

// GetName returns the tool name
func (b *BaseTool) GetName() string {
	return b.Name
}

// GetDescription returns the tool description
func (b *BaseTool) GetDescription() string {
	return b.Description
}

// GetSchema returns the tool's input schema
func (b *BaseTool) GetSchema() models.InputSchema {
	return b.Schema
}

// Execute processes the input using the tool's handler
func (b *BaseTool) Execute(input json.RawMessage) (string, error) {
	if b.Handler == nil {
		return "", errors.New("tool handler not implemented")
	}

	// Parse the input into a generic map
	var params map[string]interface{}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", err
	}

	// Validate required parameters
	for _, required := range b.Schema.Required {
		if _, exists := params[required]; !exists {
			return "", errors.New("missing required parameter: " + required)
		}
	}

	return b.Handler(params)
}

// NewTool creates a new tool with the provided configuration
func NewTool(name, description string, schema models.InputSchema, handler func(map[string]interface{}) (string, error)) *BaseTool {
	return &BaseTool{
		Name:        name,
		Description: description,
		Schema:      schema,
		Handler:     handler,
	}
}