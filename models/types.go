package models

// Role constants define the standard roles used in LLM conversations.
// These roles indicate who is speaking in each message of a conversation.
const (
	// RoleUser represents messages from the end user.
	RoleUser = "user"

	// RoleAssistant represents messages from the AI assistant.
	RoleAssistant = "assistant"

	// RoleSystem represents system instructions that guide the AI's behavior.
	RoleSystem = "system"

	// RoleFunction represents messages containing results from function/tool calls.
	RoleFunction = "function"
)

// ToolUse represents an AI model's request to use a tool.
// It contains the tool name and input parameters formatted as a JSON string.
type ToolUse struct {
	// Name is the identifier of the tool to be called.
	Name string `json:"name"`

	// Input contains the parameters for the tool call as a JSON-formatted string.
	// This can be a JSON string or structured parameters depending on the provider.
	Input any `json:"input"`
}

// ToolResult represents the result of a tool execution.
// It's used to track and communicate the outcome of tool calls back to the AI.
type ToolResult struct {
	// Name is the name of the tool that was executed.
	Name string `json:"name"`

	// Result contains the string output from the tool execution.
	Result string `json:"result"`

	// Content contains structured output when plain text is insufficient.
	// This is used for rich content like images, tables, or other structured data.
	Content []ContentItem `json:"content,omitempty"`

	// IsError indicates if the tool execution resulted in an error.
	IsError bool `json:"is_error,omitempty"`
}

// ContentItem represents a single piece of content in a tool result.
// It can be text, binary data, or a reference to a resource.
type ContentItem struct {
	// Type is the kind of content (text, data, resource).
	Type string `json:"type"`

	// Text is the text content, if applicable.
	Text string `json:"text,omitempty"`

	// Data is the binary data, if applicable, encoded as a base64 string.
	Data string `json:"data,omitempty"`

	// MimeType is the MIME type of the content.
	MimeType string `json:"mime_type,omitempty"`

	// Resource is a reference to an external resource, if applicable.
	Resource *Resource `json:"resource,omitempty"`
}

// Resource represents an external resource referenced in a content item.
type Resource struct {
	// URI is the location of the resource.
	URI string `json:"uri"`

	// MimeType is the MIME type of the resource.
	MimeType string `json:"mime_type"`

	// Text is an optional text representation of the resource.
	Text string `json:"text,omitempty"`
}

// Message represents a single message in a conversation with an AI model.
// It can be a user input, AI response, system instruction, or tool-related message.
type Message struct {
	// Role identifies who is speaking (user, assistant, system, or function).
	Role string `json:"role"`

	// Content contains the actual message text.
	Content string `json:"content"`

	// ToolUse is present when an AI requests to use a tool.
	ToolUse *ToolUse `json:"tool_use,omitempty"`

	// ToolResult is present when including the result of a tool execution.
	ToolResult *ToolResult `json:"tool_result,omitempty"`
}

// InputSchema defines the structure of tool inputs following a simplified JSON Schema format.
// It specifies the parameters a tool accepts, their types, and which ones are required.
type InputSchema struct {
	// Type is usually "object" for tool inputs.
	Type string `json:"type"`

	// Description explains what the tool does, helping the AI decide when to use it.
	Description string `json:"description"`

	// Properties maps parameter names to their type definitions.
	Properties map[string]Property `json:"properties"`

	// Required lists which parameters must be provided.
	Required []string `json:"required,omitempty"`

	// AdditionalProperties indicates whether additional properties are allowed.
	AdditionalProperties *bool `json:"additionalProperties,omitempty"`
	
	// MinProperties specifies the minimum number of properties in an object.
	MinProperties *int `json:"minProperties,omitempty"`
	
	// MaxProperties specifies the maximum number of properties in an object.
	MaxProperties *int `json:"maxProperties,omitempty"`
	
	// OneOf specifies a set of possible schemas, of which one must match.
	OneOf []InputSchema `json:"oneOf,omitempty"`
	
	// AnyOf specifies a set of possible schemas, of which at least one must match.
	AnyOf []InputSchema `json:"anyOf,omitempty"`
	
	// AllOf specifies a set of schemas, all of which must match.
	AllOf []InputSchema `json:"allOf,omitempty"`
	
	// Not specifies a schema which must not match.
	Not *InputSchema `json:"not,omitempty"`
}

// Property defines a single parameter in a tool's input schema.
// It specifies the parameter's type and provides a description.
type Property struct {
	// Type is the JSON type of this property (string, number, boolean, etc.).
	Type string `json:"type"`

	// Description explains what this parameter is used for.
	Description string `json:"description"`

	// Properties is a nested schema for nested objects.
	Properties map[string]Property `json:"properties,omitempty"`

	// Format specifies the format of a string value (e.g., "email", "date-time").
	Format string `json:"format,omitempty"`

	// Enum lists the allowed values for this property.
	Enum []any `json:"enum,omitempty"`

	// Minimum specifies the minimum value for a numeric property.
	Minimum *float64 `json:"minimum,omitempty"`

	// Maximum specifies the maximum value for a numeric property.
	Maximum *float64 `json:"maximum,omitempty"`

	// MinLength specifies the minimum length for a string property.
	MinLength *int `json:"minLength,omitempty"`

	// MaxLength specifies the maximum length for a string property.
	MaxLength *int `json:"maxLength,omitempty"`

	// MinItems specifies the minimum number of items for an array property.
	MinItems *int `json:"minItems,omitempty"`

	// MaxItems specifies the maximum number of items for an array property.
	MaxItems *int `json:"maxItems,omitempty"`

	// Default specifies the default value for this property.
	Default any `json:"default,omitempty"`

	// OneOf specifies a set of possible schemas, of which one must match.
	OneOf []Property `json:"oneOf,omitempty"`

	// Items specifies the schema for array items.
	Items *Property `json:"items,omitempty"`

	// Examples provides example values for this property.
	Examples []any `json:"examples,omitempty"`

	// Required lists which properties must be provided (for nested objects).
	Required []string `json:"required,omitempty"`

	// AdditionalProperties indicates whether additional properties are allowed.
	AdditionalProperties *bool `json:"additionalProperties,omitempty"`
	
	// Pattern specifies a regex pattern for string validation.
	Pattern string `json:"pattern,omitempty"`
	
	// MultipleOf specifies that a number must be a multiple of this value.
	MultipleOf *float64 `json:"multipleOf,omitempty"`
	
	// ExclusiveMinimum indicates whether the minimum value is exclusive.
	ExclusiveMinimum *bool `json:"exclusiveMinimum,omitempty"`
	
	// ExclusiveMaximum indicates whether the maximum value is exclusive.
	ExclusiveMaximum *bool `json:"exclusiveMaximum,omitempty"`
	
	// UniqueItems indicates whether array items must be unique.
	UniqueItems *bool `json:"uniqueItems,omitempty"`
	
	// MaxProperties specifies the maximum number of properties in an object.
	MaxProperties *int `json:"maxProperties,omitempty"`
	
	// MinProperties specifies the minimum number of properties in an object.
	MinProperties *int `json:"minProperties,omitempty"`
	
	// AnyOf specifies a set of possible schemas, of which at least one must match.
	AnyOf []Property `json:"anyOf,omitempty"`
	
	// AllOf specifies a set of schemas, all of which must match.
	AllOf []Property `json:"allOf,omitempty"`
	
	// Not specifies a schema which must not match.
	Not *Property `json:"not,omitempty"`
}
