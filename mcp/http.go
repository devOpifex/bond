package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/devOpifex/bond/models"
)

// HTTPMCPConfig contains the configuration for an HTTP MCP
type HTTPMCPConfig struct {
	ID           string            `json:"id"`
	URL          string            `json:"url"`
	Headers      map[string]string `json:"headers,omitempty"`
	TimeoutMs    int               `json:"timeout_ms,omitempty"`
	ToolsPath    string            `json:"tools_path,omitempty"`
	RegisterPath string            `json:"register_path,omitempty"`
}

// HTTPMCP implements the MCPProvider interface for HTTP-based MCPs
type HTTPMCP struct {
	config      HTTPMCPConfig
	httpClient  *http.Client
	initialized bool
}

// GetURL returns the base URL for the MCP
func (m *HTTPMCP) GetURL() string {
	return m.config.URL
}

// GetHeaders returns the headers for the MCP
func (m *HTTPMCP) GetHeaders() map[string]string {
	return m.config.Headers
}

// GetHTTPClient returns the HTTP client for the MCP
func (m *HTTPMCP) GetHTTPClient() *http.Client {
	return m.httpClient
}

// NewHTTPMCP creates a new HTTP MCP provider
func NewHTTPMCP(config HTTPMCPConfig) *HTTPMCP {
	// Set default timeout if not specified
	timeoutMs := config.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 30000 // Default 30 seconds
	}

	// Set default paths if not specified
	if config.ToolsPath == "" {
		config.ToolsPath = "/tools"
	}

	return &HTTPMCP{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutMs) * time.Millisecond,
		},
		initialized: false,
	}
}

// GetID returns the unique identifier for this MCP
func (m *HTTPMCP) GetID() string {
	return m.config.ID
}

// Initialize sets up the MCP for use
func (m *HTTPMCP) Initialize(ctx context.Context) error {
	// For HTTP MCPs, we just need to check connectivity
	req, err := http.NewRequestWithContext(ctx, "GET", m.config.URL+m.config.ToolsPath, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range m.config.Headers {
		req.Header.Add(key, value)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to MCP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("MCP returned non-OK status: %d - %s", resp.StatusCode, string(body))
	}

	m.initialized = true
	return nil
}

// GetTools retrieves the list of tools provided by this MCP
func (m *HTTPMCP) GetTools(ctx context.Context) ([]models.Tool, error) {
	if !m.initialized {
		return nil, fmt.Errorf("MCP not initialized, call Initialize() first")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", m.config.URL+m.config.ToolsPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range m.config.Headers {
		req.Header.Add(key, value)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get tools from MCP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("MCP returned non-OK status: %d - %s", resp.StatusCode, string(body))
	}

	var tools []models.Tool
	if err := json.NewDecoder(resp.Body).Decode(&tools); err != nil {
		return nil, fmt.Errorf("failed to decode tools response: %w", err)
	}

	return tools, nil
}

// Shutdown cleans up resources used by the MCP
func (m *HTTPMCP) Shutdown() error {
	// For HTTP MCPs, there's not much to clean up
	m.initialized = false
	return nil
}

