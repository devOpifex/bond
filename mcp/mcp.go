package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// MCP represents a Machine Consumable Protocol handler
type MCP struct {
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
	command string
	args    []string
}

// NewMCP creates a new MCP instance with the provided IO and command
func NewMCP(stdin io.Reader, stdout, stderr io.Writer, command string, args []string) *MCP {
	return &MCP{
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
		command: command,
		args:    args,
	}
}

// New creates a new MCP instance with standard IO
func New(command string, args []string) *MCP {
	return NewMCP(os.Stdin, os.Stdout, os.Stderr, command, args)
}

// Start begins the MCP protocol handling loop and executes the command
func (m *MCP) Start() error {
	// Only start the MCP protocol if a command is provided
	if m.command == "" {
		return m.startProtocol()
	}

	// Execute the command
	cmd := exec.Command(m.command, m.args...)
	cmd.Stdin = m.stdin
	cmd.Stdout = m.stdout
	cmd.Stderr = m.stderr

	fmt.Fprintf(m.stderr, "Starting command: %s %v\n", m.command, m.args)
	
	return cmd.Run()
}

// startProtocol handles the MCP protocol communication
func (m *MCP) startProtocol() error {
	decoder := json.NewDecoder(m.stdin)
	encoder := json.NewEncoder(m.stdout)

	for {
		var request MCPRequest
		if err := decoder.Decode(&request); err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("failed to decode request: %w", err)
		}

		// Process the request and prepare a response
		response := m.processRequest(request)

		if err := encoder.Encode(response); err != nil {
			return fmt.Errorf("failed to encode response: %w", err)
		}
	}
}

// processRequest handles an incoming MCP request and returns a response
func (m *MCP) processRequest(request MCPRequest) MCPResponse {
	response := MCPResponse{
		Status: "ok",
		ID:     request.ID,
	}

	switch request.Command {
	case "ping":
		response.Data = map[string]any{
			"message": "pong",
		}
	case "tool_use":
		// Handle tool use request - this is a placeholder
		// In a full implementation, this would execute the requested tool
		response.ToolUse = &ToolUseResponse{
			ToolName: request.ToolName,
			Result: map[string]any{
				"message": "Tool execution not yet implemented",
			},
		}
	default:
		response.Status = "error"
		response.Error = fmt.Sprintf("unknown command: %s", request.Command)
	}

	return response
}

