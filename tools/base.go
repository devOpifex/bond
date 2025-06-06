// Package tools implements a framework for creating and managing tools that AI models can use.
// Tools are functions that AI agents can call to perform actions or retrieve information
// during their reasoning process. This package provides base types, validation, and registration
// mechanisms for tools.
package tools

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/devOpifex/bond/models"
)

// BaseTool provides a standard implementation of the ToolExecutor interface.
// It handles parameter validation and execution of tool functions through a handler.
type BaseTool struct {
	// Name is the identifier used to call this tool
	Name string

	// Description explains what the tool does, helping the AI decide when to use it
	Description string

	// Schema defines the structure of inputs that this tool accepts
	Schema models.InputSchema

	// Handler is the function that implements the tool's actual functionality
	Handler func(params map[string]any) (string, error)
}

// IsNamespaced returns true if the tool is namespaced, meaning it has a namespace prefix.
// This method implements part of the ToolExecutor interface.
func (b *BaseTool) IsNamespaced() bool {
	return strings.Contains(b.Name, ":")
}

// Namespace adds a namespace prefix to the tool's name.
func (b *BaseTool) Namespace(namespace string) {
	b.Name = namespace + ":" + b.Name
}

// GetName returns the tool's name, which is used to identify it when called.
// This method implements part of the ToolExecutor interface.
func (b *BaseTool) GetName() string {
	return b.Name
}

// GetDescription returns a human-readable description of what the tool does.
// This method implements part of the ToolExecutor interface.
func (b *BaseTool) GetDescription() string {
	return b.Description
}

// GetSchema returns the input schema that defines the structure of input parameters.
// This method implements part of the ToolExecutor interface.
func (b *BaseTool) GetSchema() models.InputSchema {
	return b.Schema
}

// Execute processes the JSON input using the tool's handler function.
// It validates that all required parameters are present before calling the handler.
// This method implements part of the ToolExecutor interface.
func (b *BaseTool) Execute(input json.RawMessage) (string, error) {
	if b.Handler == nil {
		return "", errors.New("tool handler not implemented")
	}

	// Parse the input into a generic map
	var params map[string]any
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

// NewTool creates a new BaseTool instance with the provided configuration.
// This is a convenience function for creating tools with proper input validation.
func NewTool(name, description string, schema models.InputSchema, handler func(map[string]any) (string, error)) *BaseTool {
	return &BaseTool{
		Name:        name,
		Description: description,
		Schema:      schema,
		Handler:     handler,
	}
}
