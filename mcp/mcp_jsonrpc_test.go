package mcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

// MockCmd is a mock command that simulates a JSON-RPC server
type MockCmd struct {
	input       *bytes.Buffer
	output      *bytes.Buffer
	errOutput   *bytes.Buffer
	inputReader io.ReadCloser
	outputWrite io.WriteCloser
	closeChan   chan struct{}
	wg          sync.WaitGroup
}

// NewMockCmd creates a new mock command
func NewMockCmd() *MockCmd {
	input := new(bytes.Buffer)
	output := new(bytes.Buffer)
	errOutput := new(bytes.Buffer)
	
	// Create a pipe for the input
	pr, pw := io.Pipe()
	
	return &MockCmd{
		input:       input,
		output:      output,
		errOutput:   errOutput,
		inputReader: pr,
		outputWrite: pw,
		closeChan:   make(chan struct{}),
	}
}

// Start starts the mock command, simulating a JSON-RPC server
func (m *MockCmd) Start() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		
		scanner := bufio.NewScanner(m.inputReader)
		for scanner.Scan() {
			line := scanner.Text()
			
			// Check if we're handling a batch request
			if strings.HasPrefix(strings.TrimSpace(line), "[") {
				// Parse batch requests
				var requests []*Request
				if err := json.Unmarshal([]byte(line), &requests); err != nil {
					// Invalid batch request
					errorResp := NewErrorResponse(ParseErrorCode, ParseErrorMsg, nil, nil)
					respData, _ := json.Marshal(errorResp)
					fmt.Fprintln(m.outputWrite, string(respData))
					continue
				}
				
				// Process each request in the batch
				responses := make([]*Response, len(requests))
				for i, req := range requests {
					responses[i] = m.handleRequest(req)
				}
				
				// Send batch response
				respData, _ := json.Marshal(responses)
				fmt.Fprintln(m.outputWrite, string(respData))
			} else {
				// Parse single request
				var req Request
				if err := json.Unmarshal([]byte(line), &req); err != nil {
					// Invalid request
					errorResp := NewErrorResponse(ParseErrorCode, ParseErrorMsg, nil, nil)
					respData, _ := json.Marshal(errorResp)
					fmt.Fprintln(m.outputWrite, string(respData))
					continue
				}
				
				// Handle the request
				resp := m.handleRequest(&req)
				
				// Send response if it's not a notification (has an ID)
				if req.ID != nil {
					respData, _ := json.Marshal(resp)
					fmt.Fprintln(m.outputWrite, string(respData))
				}
			}
		}
	}()
}

// handleRequest processes a single JSON-RPC request and returns a response
func (m *MockCmd) handleRequest(req *Request) *Response {
	// Check for notifications (no ID)
	if req.ID == nil {
		// Just log notifications to stderr
		fmt.Fprintf(m.errOutput, "Notification received: %s\n", req.Method)
		return nil
	}
	
	// Handle different methods
	switch req.Method {
	case "ping":
		// Echo back the params
		return NewResponse(map[string]interface{}{
			"pong":   true,
			"params": req.Params,
		}, req.ID)
		
	case "echo":
		// Echo back the exact params
		return NewResponse(req.Params, req.ID)
		
	case "delay":
		// Simulate a delayed response
		time.Sleep(100 * time.Millisecond)
		return NewResponse("delayed response", req.ID)
		
	case "error":
		// Return an error response
		return NewErrorResponse(InvalidRequestCode, "Requested error", nil, req.ID)
		
	default:
		// Method not found
		return NewErrorResponse(MethodNotFoundCode, fmt.Sprintf("Method '%s' not found", req.Method), nil, req.ID)
	}
}

// Stop stops the mock command
func (m *MockCmd) Stop() {
	close(m.closeChan)
	m.outputWrite.Close()
	m.wg.Wait()
}

// TestMCPJSONRPC tests the JSON-RPC functionality of the MCP struct
func TestMCPJSONRPC(t *testing.T) {
	mock := NewMockCmd()
	mock.Start()
	defer mock.Stop()
	
	// Create a new MCP instance with the mock command
	mcpInstance := NewMCP(mock.input, mock.output, mock.errOutput, "mock", []string{})
	mcpInstance.SetDefaultTimeout(500 * time.Millisecond)
	
	// Replace the command with our mock
	mcpInstance.cmdStdin = mock.outputWrite
	mcpInstance.cmdStdout = mock.inputReader
	
	// Set running to true (since we're not actually starting a command)
	mcpInstance.running = true
	
	// Start the response handler
	go mcpInstance.handleResponses()
	
	// Test sending a request and receiving a response
	t.Run("Call", func(t *testing.T) {
		resp, err := mcpInstance.Call("ping", map[string]interface{}{
			"message": "Hello, MCP!",
		})
		
		if err != nil {
			t.Errorf("Call failed: %v", err)
		}
		
		if resp == nil {
			t.Fatal("Expected response, got nil")
		}
		
		result, ok := resp.Result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected result to be map[string]interface{}, got %T", resp.Result)
		}
		
		pong, ok := result["pong"].(bool)
		if !ok || !pong {
			t.Errorf("Expected pong to be true, got %v", result["pong"])
		}
	})
	
	// Test sending a notification
	t.Run("Notify", func(t *testing.T) {
		err := mcpInstance.Notify("log", map[string]interface{}{
			"level":   "info",
			"message": "Test notification",
		})
		
		if err != nil {
			t.Errorf("Notify failed: %v", err)
		}
		
		// Check if notification was logged to stderr
		if !strings.Contains(mock.errOutput.String(), "Notification received: log") {
			t.Errorf("Expected notification to be logged, got: %s", mock.errOutput.String())
		}
	})
	
	// Test error response
	t.Run("ErrorResponse", func(t *testing.T) {
		resp, err := mcpInstance.Call("error", nil)
		
		if err != nil {
			t.Errorf("Call failed: %v", err)
		}
		
		if resp == nil {
			t.Fatal("Expected response, got nil")
		}
		
		if resp.Error == nil {
			t.Fatal("Expected error in response, got nil")
		}
		
		if resp.Error.Code != InvalidRequestCode {
			t.Errorf("Expected error code %d, got %d", InvalidRequestCode, resp.Error.Code)
		}
	})
	
	// Test timeout
	t.Run("Timeout", func(t *testing.T) {
		mcpInstance.SetDefaultTimeout(50 * time.Millisecond) // Set timeout shorter than delay method
		
		_, err := mcpInstance.Call("delay", nil)
		
		if err == nil {
			t.Error("Expected timeout error, got nil")
		}
		
		// Reset timeout to avoid affecting other tests
		mcpInstance.SetDefaultTimeout(500 * time.Millisecond)
	})
}