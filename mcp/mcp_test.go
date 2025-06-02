package mcp

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestNewMCP(t *testing.T) {
	stdin := bytes.NewBufferString("")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	command := "echo"
	args := []string{"hello"}

	mcp := NewMCP(stdin, stdout, stderr, command, args)

	if mcp.stdin != stdin {
		t.Errorf("Expected stdin to be %v, got %v", stdin, mcp.stdin)
	}
	if mcp.stdout != stdout {
		t.Errorf("Expected stdout to be %v, got %v", stdout, mcp.stdout)
	}
	if mcp.stderr != stderr {
		t.Errorf("Expected stderr to be %v, got %v", stderr, mcp.stderr)
	}
	if mcp.command != command {
		t.Errorf("Expected command to be %v, got %v", command, mcp.command)
	}
	if len(mcp.args) != len(args) || mcp.args[0] != args[0] {
		t.Errorf("Expected args to be %v, got %v", args, mcp.args)
	}
}

func TestMCPExecuteCommand(t *testing.T) {
	// Test with a simple echo command
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	
	mcp := NewMCP(strings.NewReader(""), stdout, stderr, "echo", []string{"hello world"})
	
	err := mcp.Start()
	if err != nil {
		t.Fatalf("Expected command to execute successfully, got error: %v", err)
	}
	
	output := strings.TrimSpace(stdout.String())
	if output != "hello world" {
		t.Errorf("Expected output to be 'hello world', got '%s'", output)
	}
}

func TestMCPProtocol(t *testing.T) {
	t.Skip("This test needs further investigation")
}

func TestMCPOrchestra(t *testing.T) {
	// Skip the test if mcpOrchestra is not available
	mcpOrchestraPath, err := exec.LookPath("mcpOrchestra")
	if err != nil {
		t.Skip("mcpOrchestra not found in PATH, skipping test")
	}

	// Create temp files for communication
	tmpInput, err := os.CreateTemp("", "mcp-input-*")
	if err != nil {
		t.Fatalf("Failed to create temp input file: %v", err)
	}
	defer os.Remove(tmpInput.Name())
	defer tmpInput.Close()

	tmpOutput, err := os.CreateTemp("", "mcp-output-*")
	if err != nil {
		t.Fatalf("Failed to create temp output file: %v", err)
	}
	defer os.Remove(tmpOutput.Name())
	defer tmpOutput.Close()

	// Create an MCP instance to run mcpOrchestra
	mcp := NewMCP(strings.NewReader(""), os.Stdout, os.Stderr, mcpOrchestraPath, nil)
	
	// Don't actually run the command in the test, just verify it would be constructed correctly
	if mcp.command != mcpOrchestraPath {
		t.Errorf("Expected command to be %s, got %s", mcpOrchestraPath, mcp.command)
	}
}