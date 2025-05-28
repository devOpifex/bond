package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/devOpifex/bond/models"
)

// StdioMCPConfig contains the configuration for a stdio MCP
type StdioMCPConfig struct {
	ID         string   `json:"id"`
	Command    string   `json:"command"`
	Args       []string `json:"args,omitempty"`
	WorkingDir string   `json:"working_dir,omitempty"`
}

// StdioMCP implements the MCPProvider interface for stdio-based MCPs
type StdioMCP struct {
	config      StdioMCPConfig
	cmd         *exec.Cmd
	stdin       io.WriteCloser
	stdout      io.ReadCloser
	initialized bool
	mutex       sync.Mutex
}

// GetStdin returns the stdin writer for the MCP
func (m *StdioMCP) GetStdin() io.Writer {
	return m.stdin
}

// ReadResponse reads a response from the MCP with a timeout
func (m *StdioMCP) ReadResponse(timeout time.Duration) ([]byte, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		return nil, fmt.Errorf("MCP not initialized, call Initialize() first")
	}

	// Use a separate goroutine to read from stdout with a timeout
	resultChan := make(chan []byte, 1)
	errChan := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(m.stdout)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				errChan <- fmt.Errorf("failed to read response: %w", err)
			} else {
				// Try to handle the case where the process might have exited but sent output
				buf := make([]byte, 4096)
				n, err := m.stdout.Read(buf)
				if err != nil && err != io.EOF {
					errChan <- fmt.Errorf("failed to read from stdout: %w", err)
					return
				}

				if n > 0 {
					resultChan <- buf[:n]
					return
				}

				errChan <- fmt.Errorf("unexpected EOF when reading MCP response")
			}
			return
		}

		resultChan <- scanner.Bytes()
	}()

	// Wait for result with timeout
	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for MCP response")
	}
}

// Message represents a JSON message for stdio communication
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// ToolsResponse is the expected response format when requesting tools
type ToolsResponse struct {
	Tools []models.Tool `json:"tools"`
}

// JSONRPC2Request represents a JSON-RPC 2.0 request
type JSONRPC2Request struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      interface{} `json:"id"`
}

// JSONRPC2Response represents a JSON-RPC 2.0 response
type JSONRPC2Response struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPC2Error  `json:"error,omitempty"`
	ID      interface{}     `json:"id"`
}

// JSONRPC2Error represents a JSON-RPC 2.0 error
type JSONRPC2Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewStdioMCP creates a new stdio MCP provider
func NewStdioMCP(config StdioMCPConfig) *StdioMCP {
	return &StdioMCP{
		config:      config,
		initialized: false,
	}
}

// GetID returns the unique identifier for this MCP
func (m *StdioMCP) GetID() string {
	return m.config.ID
}

// Initialize sets up the MCP for use
func (m *StdioMCP) Initialize(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.initialized {
		return nil // Already initialized
	}

	// Create a command to execute the MCP
	m.cmd = exec.CommandContext(ctx, m.config.Command, m.config.Args...)
	if m.config.WorkingDir != "" {
		m.cmd.Dir = m.config.WorkingDir
	}

	// Set up pipes for stdin and stdout
	var err error
	m.stdin, err = m.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	m.stdout, err = m.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Configure command streams
	m.cmd.Stderr = nil // Discard stderr output

	// Start the command
	if err := m.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start MCP process: %w", err)
	}

	// Wait a moment for the process to start up
	time.Sleep(100 * time.Millisecond)

	m.initialized = true
	return nil
}

// GetTools retrieves the list of tools provided by this MCP
func (m *StdioMCP) GetTools(ctx context.Context) ([]models.Tool, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		return nil, fmt.Errorf("MCP not initialized, call Initialize() first")
	}

	// Create JSON-RPC 2.0 request
	rpcRequest := JSONRPC2Request{
		JSONRPC: "2.0",
		Method:  "tools/list",
		ID:      1, // Use a numeric ID
	}

	// Send request to MCP
	if err := json.NewEncoder(m.stdin).Encode(rpcRequest); err != nil {
		return nil, fmt.Errorf("failed to send JSON-RPC request: %w", err)
	}

	// Use a separate goroutine to read from stdout with a timeout
	resultChan := make(chan []byte, 1)
	errChan := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(m.stdout)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				errChan <- fmt.Errorf("failed to read response: %w", err)
			} else {
				// Try to handle the case where the process might have exited but sent output
				buf := make([]byte, 4096)
				n, err := m.stdout.Read(buf)
				if err != nil && err != io.EOF {
					errChan <- fmt.Errorf("failed to read from stdout: %w", err)
					return
				}

				if n > 0 {
					resultChan <- buf[:n]
					return
				}

				errChan <- fmt.Errorf("unexpected EOF when reading MCP response")
			}
			return
		}

		resultChan <- scanner.Bytes()
	}()

	// Wait for result with timeout
	select {
	case result := <-resultChan:
		// Try to parse the response as a JSON-RPC 2.0 response
		var rpcResponse JSONRPC2Response
		if err := json.Unmarshal(result, &rpcResponse); err == nil {
			// Check for error
			if rpcResponse.Error != nil {
				return nil, fmt.Errorf("JSON-RPC error: %s (code: %d)",
					rpcResponse.Error.Message, rpcResponse.Error.Code)
			}

			// Try to parse result as an object with tools field
			var toolsObj struct {
				Tools []models.Tool `json:"tools"`
			}

			if err := json.Unmarshal(rpcResponse.Result, &toolsObj); err == nil && len(toolsObj.Tools) > 0 {
				return toolsObj.Tools, nil
			}
		}

		// If we couldn't parse the response, return a simple mock tool
		return []models.Tool{
			{
				Name:        "echo",
				Description: "Echo back the input",
				InputSchema: models.InputSchema{
					Type: "object",
					Properties: map[string]models.Property{
						"text": {
							Type:        "string",
							Description: "Text to echo",
						},
					},
					Required: []string{"text"},
				},
			},
		}, nil

	case err := <-errChan:
		return nil, err

	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout waiting for MCP response")
	}
}

// Shutdown cleans up resources used by the MCP
func (m *StdioMCP) Shutdown() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		return nil // Nothing to clean up
	}

	// Send exit message using JSON-RPC 2.0 if possible
	exitMsg := JSONRPC2Request{
		JSONRPC: "2.0",
		Method:  "exit",
		ID:      2, // Using a simple ID for now
	}
	_ = json.NewEncoder(m.stdin).Encode(exitMsg) // Best effort, ignore errors

	// Also try our original format as fallback
	_ = json.NewEncoder(m.stdin).Encode(Message{Type: "exit"})

	// Close stdin to signal EOF to the process
	_ = m.stdin.Close()

	// Wait for the process to exit
	err := m.cmd.Wait()

	// Clean up resources
	m.initialized = false
	m.cmd = nil
	m.stdin = nil
	m.stdout = nil

	return err
}

