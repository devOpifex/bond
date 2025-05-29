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

### Provider Interface

`Provider` defines the interface that all AI service providers must implement:

```go
type Provider interface {
	SendMessage(ctx context.Context, message Message) (string, error)
	SendMessageWithHistory(ctx context.Context, messages []Message) (string, error)
	SendMessageWithTools(ctx context.Context, message Message) (string, error)
	RegisterTool(tool tools.Tool)
}
```

This interface allows Bond to work with multiple AI providers (like OpenAI, Claude, etc.) through a consistent API.

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
response, err := provider.SendMessageWithHistory(ctx, []models.Message{
    systemMessage,
    userMessage,
})
```