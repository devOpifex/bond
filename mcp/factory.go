package mcp

import (
	"fmt"

	"github.com/devOpifex/bond/models"
)

// NewMCPProvider creates a new MCP provider based on configuration
func NewMCPProvider(config models.MCPConfig) (models.MCPProvider, error) {
	switch config.Type {
	case models.MCPTypeHTTP:
		// Create HTTP MCP
		return NewHTTPMCP(HTTPMCPConfig{
			ID:        config.ID,
			URL:       config.URL,
			Headers:   config.Headers,
			TimeoutMs: config.TimeoutMillis,
		}), nil
	case models.MCPTypeStdio:
		// Create stdio MCP
		return NewStdioMCP(StdioMCPConfig{
			ID:         config.ID,
			Command:    config.Command,
			Args:       config.Args,
			WorkingDir: config.WorkingDir,
		}), nil
	default:
		return nil, fmt.Errorf("unsupported MCP type: %s", config.Type)
	}
}