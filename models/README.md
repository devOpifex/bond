# Bond Models

The `models` package defines core data structures and interfaces used throughout the Bond framework.

## Key Components

### Message

`Message` represents a single message in a conversation:

```go
type Message struct {
	Role    string
	Content string
}
```

The `Role` field can be one of the predefined constants:
- `RoleUser`: Messages from the end user
- `RoleAssistant`: Messages from the AI assistant
- `RoleSystem`: System instructions for the AI
- `RoleFunction`: Messages from function/tool calls

### Provider Interface

`Provider` defines the interface that all AI service providers must implement:

```go
type Provider interface {
	SendMessage(ctx context.Context, message Message) (string, error)
	SendMessageWithTools(ctx context.Context, message Message) (string, error)
	RegisterTool(tool ToolExecutor)
	SetSystemPrompt(prompt string)
	SetModel(model string)
	SetMaxTokens(tokens int)
	SetTemperature(temperature float64)
	RegisterMCP(command string, args []string) error
}
```

This interface allows Bond to work with multiple AI providers (like OpenAI, Claude, etc.) through a consistent API. The `RegisterMCP` method supports integration with Model Context Protocol servers.

### ToolExecutor Interface

`ToolExecutor` defines the interface for tools that can be executed by AI models:

```go
type ToolExecutor interface {
	IsNamespaced() bool
	Namespace(string)
	GetName() string
	GetDescription() string
	GetSchema() InputSchema
	Execute(input json.RawMessage) (string, error)
}
```

### Agent Interface

`Agent` defines the interface that all AI agents must implement:

```go
type Agent interface {
	Process(ctx context.Context, input string) (string, error)
}
```

### Input Schema

`InputSchema` defines the JSON schema structure for tool parameters:

```go
type InputSchema struct {
	Type       string
	Properties map[string]Property
	Required   []string
}

type Property struct {
	Type        string
	Description string
	Enum        []string
}
```

This schema is used to validate and document tool parameters when registering tools with providers.

### Tool Result

`ToolResult` represents the result of executing a tool:

```go
type ToolResult struct {
	Name    string
	Result  string
	IsError bool
	Content []ContentItem
}

type ContentItem struct {
	Type     string
	Text     string
	Data     string
	MimeType string
	Resource *Resource
}

type Resource struct {
	URI      string
	MimeType string
	Text     string
}
```

## Example Usage

```go
// Creating messages
systemMessage := models.Message{
    Role:    models.RoleSystem,
    Content: "You are a helpful assistant.",
}

userMessage := models.Message{
    Role:    models.RoleUser,
    Content: "What's the weather like today?",
}

// Using with a provider
response, err := provider.SendMessageWithTools(ctx, userMessage)
```