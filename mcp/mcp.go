package mcp

import (
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
	// Execute the command
	cmd := exec.Command(m.command, m.args...)
	cmd.Stdin = m.stdin
	cmd.Stdout = m.stdout
	cmd.Stderr = m.stderr

	fmt.Fprintf(m.stderr, "Starting command: %s %v\n", m.command, m.args)

	return cmd.Run()
}
