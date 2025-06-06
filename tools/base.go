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

// ToolAnnotations provides structured metadata about a tool that helps
// guide both humans and AI models in how to use it appropriately.
type ToolAnnotations struct {
	// Audience specifies who the tool is intended for (e.g., "humans", "llms", etc.)
	Audience []string `json:"audience,omitempty"`

	// Level indicates the importance or risk level (e.g., "info", "warning", "danger")
	Level string `json:"level,omitempty"`

	// Experimental indicates if the tool is in experimental status
	Experimental bool `json:"experimental,omitempty"`

	// Since version when the tool was introduced
	Since string `json:"since,omitempty"`

	// Version indicates the current version of the tool
	Version string `json:"version,omitempty"`

	// Deprecated indicates if the tool is deprecated
	Deprecated bool `json:"deprecated,omitempty"`

	// DeprecatedSince indicates the version when the tool was deprecated
	DeprecatedSince string `json:"deprecatedSince,omitempty"`

	// DeprecationReason explains why the tool was deprecated
	DeprecationReason string `json:"deprecationReason,omitempty"`

	// ReplacedBy indicates which tool replaced this deprecated tool
	ReplacedBy string `json:"replacedBy,omitempty"`

	// RemovalDate indicates when the deprecated tool will be removed
	RemovalDate string `json:"removalDate,omitempty"`

	// Title provides a human-readable title for the tool
	Title string `json:"title,omitempty"`

	// ReadOnlyHint indicates if the tool is read-only (doesn't modify state)
	ReadOnlyHint bool `json:"readOnlyHint,omitempty"`

	// DestructiveHint indicates if the tool may destroy or overwrite data
	DestructiveHint bool `json:"destructiveHint,omitempty"`

	// IdempotentHint indicates if the tool is idempotent (can be called multiple times with same effect)
	IdempotentHint bool `json:"idempotentHint,omitempty"`

	// OpenWorldHint indicates if the tool operates on user-defined inputs (vs. fixed options)
	OpenWorldHint bool `json:"openWorldHint,omitempty"`

	// AuthRequired indicates if authentication is required
	AuthRequired bool `json:"authRequired,omitempty"`

	// AuthType specifies the type of authentication (e.g., "bearer_token", "oauth2", "api_key")
	AuthType string `json:"authType,omitempty"`

	// Scopes required for OAuth authentication
	Scopes []string `json:"scopes,omitempty"`

	// Dangerous indicates if the tool is potentially dangerous
	Dangerous bool `json:"dangerous,omitempty"`

	// RequiresConfirmation indicates if user confirmation is needed before execution
	RequiresConfirmation bool `json:"requiresConfirmation,omitempty"`

	// Sandboxed indicates if the tool runs in a sandboxed environment
	Sandboxed bool `json:"sandboxed,omitempty"`

	// RateLimit provides rate limiting information
	RateLimit *RateLimit `json:"rateLimit,omitempty"`

	// Additional unstructured annotations
	Additional map[string]any `json:"-"`
}

// UnmarshalJSON provides custom unmarshaling for ToolAnnotations to handle
// both structured and unstructured annotation formats from MCPs
func (a *ToolAnnotations) UnmarshalJSON(data []byte) error {
	// Define a type alias to avoid infinite recursion when unmarshaling
	type Alias ToolAnnotations

	// Try to unmarshal as structured format first
	structured := &struct {
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, structured); err != nil {
		// If structured format fails, try unstructured format
		var unstructured map[string]any
		if err := json.Unmarshal(data, &unstructured); err != nil {
			return err
		}

		// Process unstructured annotations
		for k, v := range unstructured {
			switch k {
			case "audience":
				if arr, ok := v.([]any); ok {
					a.Audience = make([]string, 0, len(arr))
					for _, item := range arr {
						if str, ok := item.(string); ok {
							a.Audience = append(a.Audience, str)
						}
					}
				}
			case "level":
				if str, ok := v.(string); ok {
					a.Level = str
				}
			case "experimental":
				if boolVal, ok := v.(bool); ok {
					a.Experimental = boolVal
				}
			case "since":
				if str, ok := v.(string); ok {
					a.Since = str
				}
			case "version":
				if str, ok := v.(string); ok {
					a.Version = str
				}
			case "deprecated":
				if boolVal, ok := v.(bool); ok {
					a.Deprecated = boolVal
				}
			case "deprecatedSince":
				if str, ok := v.(string); ok {
					a.DeprecatedSince = str
				}
			case "deprecationReason":
				if str, ok := v.(string); ok {
					a.DeprecationReason = str
				}
			case "replacedBy":
				if str, ok := v.(string); ok {
					a.ReplacedBy = str
				}
			case "removalDate":
				if str, ok := v.(string); ok {
					a.RemovalDate = str
				}
			case "title":
				if str, ok := v.(string); ok {
					a.Title = str
				}
			case "readOnlyHint":
				if boolVal, ok := v.(bool); ok {
					a.ReadOnlyHint = boolVal
				}
			case "destructiveHint":
				if boolVal, ok := v.(bool); ok {
					a.DestructiveHint = boolVal
				}
			case "idempotentHint":
				if boolVal, ok := v.(bool); ok {
					a.IdempotentHint = boolVal
				}
			case "openWorldHint":
				if boolVal, ok := v.(bool); ok {
					a.OpenWorldHint = boolVal
				}
			case "authRequired":
				if boolVal, ok := v.(bool); ok {
					a.AuthRequired = boolVal
				}
			case "authType":
				if str, ok := v.(string); ok {
					a.AuthType = str
				}
			case "scopes":
				if arr, ok := v.([]any); ok {
					a.Scopes = make([]string, 0, len(arr))
					for _, item := range arr {
						if str, ok := item.(string); ok {
							a.Scopes = append(a.Scopes, str)
						}
					}
				}
			case "dangerous":
				if boolVal, ok := v.(bool); ok {
					a.Dangerous = boolVal
				}
			case "requiresConfirmation":
				if boolVal, ok := v.(bool); ok {
					a.RequiresConfirmation = boolVal
				}
			case "sandboxed":
				if boolVal, ok := v.(bool); ok {
					a.Sandboxed = boolVal
				}
			case "rateLimit":
				if rl, ok := v.(map[string]any); ok {
					a.RateLimit = &RateLimit{}
					if req, ok := rl["requests"].(float64); ok {
						a.RateLimit.Requests = int(req)
					}
					if period, ok := rl["period"].(string); ok {
						a.RateLimit.Period = period
					}
				}
			default:
				// Store unknown fields in Additional
				if a.Additional == nil {
					a.Additional = make(map[string]any)
				}
				a.Additional[k] = v
			}
		}
	}

	return nil
}

// RateLimit defines rate limiting constraints for a tool
type RateLimit struct {
	// Requests is the maximum number of requests allowed
	Requests int `json:"requests"`

	// Period is the time period for the rate limit (e.g., "minute", "hour", "day")
	Period string `json:"period"`
}

// BaseTool provides a standard implementation of the ToolExecutor interface.
// It handles parameter validation and execution of tool functions through a handler.
type BaseTool struct {
	// Name is the identifier used to call this tool
	Name string `json:"name"`

	// Description explains what the tool does, helping the AI decide when to use it
	Description string `json:"description"`

	// Schema defines the structure of inputs that this tool accepts
	Schema models.InputSchema `json:"inputSchema"`

	// Handler is the function that implements the tool's actual functionality
	Handler func(params map[string]any) (string, error) `json:"-"`

	// Annotations provides structured metadata about the tool
	Annotations *ToolAnnotations `json:"annotations,omitempty"`

	// Legacy unstructured annotations, for backward compatibility
	LegacyAnnotations map[string]any `json:"-"`
}

// IsNamespaced returns true if the tool is namespaced, meaning it has a namespace prefix.
// This method implements part of the ToolExecutor interface.
func (b *BaseTool) IsNamespaced() bool {
	return strings.Contains(b.Name, "__")
}

// Namespace adds a namespace prefix to the tool's name.
func (b *BaseTool) Namespace(namespace string) {
	b.Name = namespace + "__" + b.Name
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
		Annotations: &ToolAnnotations{},
	}
}
