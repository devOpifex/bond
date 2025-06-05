package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/devOpifex/bond/models"
	"github.com/devOpifex/bond/tools"
)

// ResponseHandler is a function that handles JSON-RPC responses
type ResponseHandler func(*Response)

// pendingRequest represents a request awaiting a response
type pendingRequest struct {
	response chan *Response
	done     chan struct{}
	timeout  *time.Timer
}

// MCPCapabilities represents the capabilities supported by an MCP server
type MCPCapabilities struct {
	Tools ToolsCapability `json:"tools,omitempty"`
}

// ToolsCapability represents the tools capabilities of an MCP server
type ToolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

// MCP represents a Model Context Protocol handler
type MCP struct {
	stdin          io.Reader
	stdout         io.Writer
	stderr         io.Writer
	command        string
	args           []string
	cmd            *exec.Cmd
	cmdStdin       io.WriteCloser
	cmdStdout      io.ReadCloser
	nextID         int
	idMutex        sync.Mutex
	running        bool
	runningMtx     sync.Mutex
	pending        map[any]*pendingRequest
	pendingMtx     sync.RWMutex
	defaultTimeout time.Duration
	handlers       map[string]ResponseHandler
	handlersMtx    sync.RWMutex
	ioMutex        sync.Mutex // Protects stdout/stderr access from data races
	capabilities   *MCPCapabilities
	toolRegistry   *tools.Registry
}

// NewMCP creates a new MCP instance with the provided IO and command
func NewMCP(stdin io.Reader, stdout, stderr io.Writer, command string, args []string) *MCP {
	return &MCP{
		stdin:          stdin,
		stdout:         stdout,
		stderr:         stderr,
		command:        command,
		args:           args,
		nextID:         1,
		pending:        make(map[any]*pendingRequest),
		defaultTimeout: 30 * time.Second, // Default timeout for requests
		handlers:       make(map[string]ResponseHandler),
		capabilities:   &MCPCapabilities{},
		toolRegistry:   tools.NewRegistry(),
	}
}

// New creates a new MCP instance with standard IO
func New(command string, args []string) *MCP {
	return NewMCP(os.Stdin, os.Stdout, os.Stderr, command, args)
}

// WithToolRegistry sets a custom tool registry for the MCP
func (m *MCP) WithToolRegistry(registry *tools.Registry) *MCP {
	m.toolRegistry = registry
	return m
}

// RegisterTool adds a tool to the MCP registry
func (m *MCP) RegisterTool(tool models.ToolExecutor) error {
	return m.toolRegistry.Register(tool)
}

// convertToMCPTool converts a Bond tool to an MCP-compatible tool definition
func convertToMCPTool(tool models.ToolExecutor) models.Tool {
	return models.Tool{
		Name:        tool.GetName(),
		Description: tool.GetDescription(),
		InputSchema: tool.GetSchema(),
	}
}

// SetDefaultTimeout sets the default timeout for requests
func (m *MCP) SetDefaultTimeout(timeout time.Duration) {
	m.defaultTimeout = timeout
}

// RegisterHandler registers a handler for a specific method
func (m *MCP) RegisterHandler(method string, handler ResponseHandler) {
	m.handlersMtx.Lock()
	defer m.handlersMtx.Unlock()
	m.handlers[method] = handler
}

// Start begins the MCP protocol handling loop and executes the command
func (m *MCP) Start() error {
	m.runningMtx.Lock()
	if m.running {
		m.runningMtx.Unlock()
		return fmt.Errorf("MCP is already running")
	}
	m.running = true
	m.runningMtx.Unlock()

	// Execute the command
	m.cmd = exec.Command(m.command, m.args...)

	var err error
	// Create pipes for stdin and stdout
	m.cmdStdin, err = m.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	m.cmdStdout, err = m.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Set stderr to the provided stderr
	m.cmd.Stderr = m.stderr

	// Start the command
	if err := m.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Start a goroutine to handle responses
	go m.handleResponses()

	// Set up a default handler for tools list changed notifications
	m.RegisterHandler("notifications/tools/list_changed", func(response *Response) {
		// Typically you would refresh the tool list here
		// This is a default handler - users can override it with their own handler
	})

	// Initialize and get server capabilities
	_, err = m.GetCapabilities()
	if err != nil {
		m.writeToStderr(fmt.Sprintf("Warning: failed to get server capabilities: %v\n", err))
		// Continue anyway - server might not support initialization
	}

	// Wait for the command to complete in a goroutine
	go func() {
		err := m.cmd.Wait()
		m.runningMtx.Lock()
		m.running = false
		m.runningMtx.Unlock()

		// Clean up any pending requests
		m.pendingMtx.Lock()
		for id, req := range m.pending {
			close(req.done)
			if req.timeout != nil {
				req.timeout.Stop()
			}
			delete(m.pending, id)
		}
		m.pendingMtx.Unlock()

		if err != nil {
			m.writeToStderr(fmt.Sprintf("Command exited with error: %v\n", err))
		}
	}()

	return nil
}

// Stop terminates the MCP process
func (m *MCP) Stop() error {
	m.runningMtx.Lock()
	defer m.runningMtx.Unlock()

	if !m.running {
		return fmt.Errorf("MCP is not running")
	}

	// Close stdin to signal the command to terminate
	if m.cmdStdin != nil {
		m.cmdStdin.Close()
	}

	// Kill the process if it doesn't terminate gracefully
	if m.cmd.Process != nil {
		return m.cmd.Process.Kill()
	}

	return nil
}

// getNextID returns the next available request ID
func (m *MCP) getNextID() int {
	m.idMutex.Lock()
	defer m.idMutex.Unlock()
	id := m.nextID
	m.nextID++
	return id
}

// handleResponses reads and processes responses from the MCP command
// writeToStdout writes to stdout with mutex protection
func (m *MCP) writeToStdout(s string) {
	m.ioMutex.Lock()
	defer m.ioMutex.Unlock()
	fmt.Fprint(m.stdout, s)
}

// writeToStderr writes to stderr with mutex protection
func (m *MCP) writeToStderr(s string) {
	m.ioMutex.Lock()
	defer m.ioMutex.Unlock()
	fmt.Fprint(m.stderr, s)
}

func (m *MCP) handleResponses() {
	scanner := bufio.NewScanner(m.cmdStdout)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024) // Increase scanner buffer size

	for scanner.Scan() {
		line := scanner.Text()

		// Parse the response
		var response Response
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			m.writeToStderr(fmt.Sprintf("Failed to parse JSON-RPC response: %v\n", err))
			continue
		}

		// We don't log responses to stdout by default to avoid cluttering the output
		// Uncomment the next line for debugging purposes only
		// m.writeToStdout(line + "\n")

		// Look for a pending request with this ID
		if response.ID != nil {
			// JSON numbers are unmarshaled as float64, but our ID might be int
			var lookupID any = response.ID
			if floatID, ok := response.ID.(float64); ok {
				lookupID = int(floatID)
			}

			m.pendingMtx.RLock()
			req, ok := m.pending[lookupID]
			if !ok && lookupID != response.ID {
				// Try with original ID if conversion didn't match
				req, ok = m.pending[response.ID]
			}
			m.pendingMtx.RUnlock()

			if ok {
				// Found a matching request, send the response
				select {
				case req.response <- &response:
					// Response sent successfully
				case <-req.done:
					// Request was cancelled or timed out
				default:
					// Response channel is full, this shouldn't happen
					m.writeToStderr(fmt.Sprintf("Warning: response channel full for request ID %v\n", response.ID))
				}

				// For requests that only need the first response, we can clean up
				// More complex protocols might need to keep the request around for multiple responses
				m.pendingMtx.Lock()
				// Delete using the same ID type that we found in the map
				if floatID, ok := response.ID.(float64); ok {
					delete(m.pending, int(floatID))
				} else {
					delete(m.pending, response.ID)
				}
				m.pendingMtx.Unlock()

				if req.timeout != nil {
					req.timeout.Stop()
				}
			} else {
				// No pending request found for this ID, check for registered handlers
				m.dispatchToHandler(&response)
			}
		} else {
			// Response has no ID (notification), dispatch to handlers
			m.dispatchToHandler(&response)
		}
	}

	if err := scanner.Err(); err != nil {
		m.writeToStderr(fmt.Sprintf("Error reading from MCP command: %v\n", err))
	}
}

// dispatchToHandler sends the response to any registered handler for its method
func (m *MCP) dispatchToHandler(response *Response) {
	// Check if this is a notification (no ID)
	if response.ID == nil {
		// For notifications like tools/list_changed
		if method, ok := response.Result.(map[string]any)["method"].(string); ok {
			m.handlersMtx.RLock()
			handler, ok := m.handlers[method]
			m.handlersMtx.RUnlock()

			if ok && handler != nil {
				// Run the handler in a separate goroutine to avoid blocking
				go handler(response)
			}
		}
		return
	}

	// Extract method from result if available
	method := ""
	if result, ok := response.Result.(map[string]any); ok {
		if methodVal, ok := result["method"].(string); ok {
			method = methodVal
		}
	}

	if method != "" {
		m.handlersMtx.RLock()
		handler, ok := m.handlers[method]
		m.handlersMtx.RUnlock()

		if ok && handler != nil {
			// Run the handler in a separate goroutine to avoid blocking
			go handler(response)
		}
	}
}

// Call sends a JSON-RPC request to the MCP command and returns the response
func (m *MCP) Call(method string, params any) (*Response, error) {
	return m.callWithTimeout(method, params, m.defaultTimeout)
}

// ListTools queries the MCP server for available tools
func (m *MCP) ListTools(cursor string) (*ToolListResult, error) {
	m.runningMtx.Lock()
	running := m.running
	m.runningMtx.Unlock()

	// If MCP is running, query the server
	if running {
		params := map[string]string{}
		if cursor != "" {
			params["cursor"] = cursor
		}

		response, err := m.Call("tools/list", params)
		if err != nil {
			return nil, fmt.Errorf("failed to list tools: %w", err)
		}

		if response.Error != nil {
			return nil, fmt.Errorf("server error: %s (code: %d)", response.Error.Message, response.Error.Code)
		}

		// Convert the result to a ToolListResult
		resultBytes, err := json.Marshal(response.Result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tool list result: %w", err)
		}

		var toolList ToolListResult
		if err := json.Unmarshal(resultBytes, &toolList); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tool list result: %w", err)
		}

		return &toolList, nil
	}

	// If MCP is not running, return the local tool registry
	tools := m.toolRegistry.GetAll()
	mcpTools := make([]models.Tool, 0, len(tools))

	for _, tool := range tools {
		mcpTools = append(mcpTools, convertToMCPTool(tool))
	}

	return &ToolListResult{
		Tools: mcpTools,
	}, nil
}

// CallTool invokes a tool on the MCP server with the given name and arguments
func (m *MCP) CallTool(name string, arguments map[string]any) (*models.ToolResult, error) {
	m.runningMtx.Lock()
	running := m.running
	m.runningMtx.Unlock()

	// If MCP is running, call the tool on the server
	if running {
		params := ToolCallParams{
			Name:      name,
			Arguments: arguments,
		}

		response, err := m.Call("tools/call", params)
		if err != nil {
			return nil, fmt.Errorf("failed to call tool: %w", err)
		}

		if response.Error != nil {
			return nil, fmt.Errorf("server error: %s (code: %d)", response.Error.Message, response.Error.Code)
		}

		// Convert the result to a ToolResult
		resultBytes, err := json.Marshal(response.Result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tool call result: %w", err)
		}

		// First try to unmarshal into our standard format
		var toolResult models.ToolResult
		if err := json.Unmarshal(resultBytes, &toolResult); err != nil {
			// If that fails, try to handle the legacy MCP format
			var mcpResult struct {
				Content []struct {
					Type     string `json:"type"`
					Text     string `json:"text,omitempty"`
					Data     string `json:"data,omitempty"`
					MimeType string `json:"mimeType,omitempty"`
					Resource *struct {
						URI      string `json:"uri"`
						MimeType string `json:"mimeType"`
						Text     string `json:"text,omitempty"`
					} `json:"resource,omitempty"`
				} `json:"content"`
				IsError bool `json:"isError"`
			}

			if err := json.Unmarshal(resultBytes, &mcpResult); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tool call result: %w", err)
			}

			// Convert to our standard format
			toolResult = models.ToolResult{
				Name:    name,
				IsError: mcpResult.IsError,
				Content: make([]models.ContentItem, 0, len(mcpResult.Content)),
			}

			// Convert each content item
			for _, item := range mcpResult.Content {
				contentItem := models.ContentItem{
					Type:     item.Type,
					Text:     item.Text,
					Data:     item.Data,
					MimeType: item.MimeType,
				}

				if item.Resource != nil {
					contentItem.Resource = &models.Resource{
						URI:      item.Resource.URI,
						MimeType: item.Resource.MimeType,
						Text:     item.Resource.Text,
					}
				}

				toolResult.Content = append(toolResult.Content, contentItem)
			}

			// If we have a single text item, set it as the Result for backward compatibility
			if len(mcpResult.Content) == 1 && mcpResult.Content[0].Type == "text" {
				toolResult.Result = mcpResult.Content[0].Text
			}
		}

		return &toolResult, nil
	}

	// If MCP is not running, execute the tool locally
	tool, exists := m.toolRegistry.Get(name)
	if !exists {
		return nil, fmt.Errorf("tool '%s' not found in registry", name)
	}

	// Convert arguments to JSON
	argsBytes, err := json.Marshal(arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tool arguments: %w", err)
	}

	// Execute the tool
	result, err := tool.Execute(argsBytes)
	if err != nil {
		return &models.ToolResult{
			Name:    name,
			Result:  fmt.Sprintf("Error executing tool: %v", err),
			IsError: true,
			Content: []models.ContentItem{
				{
					Type: "text",
					Text: fmt.Sprintf("Error executing tool: %v", err),
				},
			},
		}, nil
	}

	// Return the result
	return &models.ToolResult{
		Name:    name,
		Result:  result,
		IsError: false,
		Content: []models.ContentItem{
			{
				Type: "text",
				Text: result,
			},
		},
	}, nil
}

// InitCapabilities starts the MCP if it's not running and gets the capabilities
func (m *MCP) InitCapabilities() (*MCPCapabilities, error) {
	m.runningMtx.Lock()
	running := m.running
	m.runningMtx.Unlock()

	// Start the MCP if it's not running
	if !running {
		if err := m.Start(); err != nil {
			return nil, fmt.Errorf("failed to start MCP: %w", err)
		}
	}

	// Get the capabilities
	capabilities, err := m.GetCapabilities()
	if err != nil {
		return nil, err
	}

	// Get the tool list to initialize tools
	_, err = m.ListTools("")
	if err != nil {
		m.writeToStderr(fmt.Sprintf("Warning: failed to list tools: %v\n", err))
		// Continue anyway - tool listing might fail but we still have capabilities
	}

	return capabilities, nil
}

// GetCapabilities fetches and stores the server's capabilities
func (m *MCP) GetCapabilities() (*MCPCapabilities, error) {
	m.runningMtx.Lock()
	running := m.running
	m.runningMtx.Unlock()

	if !running {
		// If not running, return default capabilities
		return &MCPCapabilities{
			Tools: ToolsCapability{
				ListChanged: true,
			},
		}, nil
	}

	response, err := m.Call("initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo": map[string]any{
			"name":    "bond-client",
			"version": "1.0.0",
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get capabilities: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("server error: %s (code: %d)", response.Error.Message, response.Error.Code)
	}

	// Extract capabilities from the result
	resultMap, ok := response.Result.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid capabilities response format")
	}

	capabilitiesMap, ok := resultMap["capabilities"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("capabilities not found in response")
	}

	// Convert the capabilities to our struct
	capabilitiesBytes, err := json.Marshal(capabilitiesMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal capabilities: %w", err)
	}

	var capabilities MCPCapabilities
	if err := json.Unmarshal(capabilitiesBytes, &capabilities); err != nil {
		return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
	}

	// Store the capabilities
	m.capabilities = &capabilities

	// Send the initialized notification to complete the handshake
	// This is required according to the MCP protocol specification
	_, err = m.Call("notifications/initialized", map[string]any{})
	if err != nil {
		// Just log the error but don't fail - the initialization was successful
		m.writeToStderr(fmt.Sprintf("Warning: failed to send initialized notification: %v\n", err))
	}

	return &capabilities, nil
}

// callWithTimeout sends a JSON-RPC request to the MCP command and waits for the response with a timeout
func (m *MCP) callWithTimeout(method string, params any, timeout time.Duration) (*Response, error) {
	m.runningMtx.Lock()
	if !m.running {
		m.runningMtx.Unlock()
		return nil, fmt.Errorf("MCP is not running")
	}
	m.runningMtx.Unlock()

	// Create a new request with the next available ID
	id := m.getNextID()
	request := NewRequest(method, params, id)

	// Set up channels to receive the response
	responseChan := make(chan *Response, 1)
	doneChan := make(chan struct{})

	// Create and register the pending request
	req := &pendingRequest{
		response: responseChan,
		done:     doneChan,
	}

	// Create a timer for the timeout
	if timeout > 0 {
		req.timeout = time.AfterFunc(timeout, func() {
			m.pendingMtx.Lock()
			delete(m.pending, id)
			m.pendingMtx.Unlock()
			close(doneChan)
		})
	}

	// Register the pending request
	m.pendingMtx.Lock()
	m.pending[id] = req
	m.pendingMtx.Unlock()

	// Marshal the request to JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		m.pendingMtx.Lock()
		delete(m.pending, id)
		m.pendingMtx.Unlock()
		if req.timeout != nil {
			req.timeout.Stop()
		}
		return nil, fmt.Errorf("failed to marshal JSON-RPC request: %w", err)
	}

	// Send the request to the MCP command
	if _, err := fmt.Fprintln(m.cmdStdin, string(requestJSON)); err != nil {
		m.pendingMtx.Lock()
		delete(m.pending, id)
		m.pendingMtx.Unlock()
		if req.timeout != nil {
			req.timeout.Stop()
		}
		return nil, fmt.Errorf("failed to send JSON-RPC request: %w", err)
	}

	// Wait for the response or timeout
	select {
	case response := <-responseChan:
		return response, nil
	case <-doneChan:
		return nil, fmt.Errorf("request timed out after %v", timeout)
	}
}
