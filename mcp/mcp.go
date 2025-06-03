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
)

// ResponseHandler is a function that handles JSON-RPC responses
type ResponseHandler func(*Response)

// pendingRequest represents a request awaiting a response
type pendingRequest struct {
	response chan *Response
	done     chan struct{}
	timeout  *time.Timer
}

// MCP represents a Machine Consumable Protocol handler
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
	}
}

// New creates a new MCP instance with standard IO
func New(command string, args []string) *MCP {
	return NewMCP(os.Stdin, os.Stdout, os.Stderr, command, args)
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

	m.writeToStderr(fmt.Sprintf("Starting command: %s %v\n", m.command, m.args))

	// Start the command
	if err := m.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Start a goroutine to handle responses
	go m.handleResponses()

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
		} else {
			m.writeToStderr("Command completed successfully\n")
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

		// Write the response to stdout for logging if needed
		m.writeToStdout(line + "\n")

		// Look for a pending request with this ID
		if response.ID != nil {
			m.pendingMtx.RLock()
			req, ok := m.pending[response.ID]
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
				delete(m.pending, response.ID)
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
