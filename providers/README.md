# Bond Providers

The `providers` package implements connections to various AI model providers such as Claude and OpenAI. This package allows Bond to interface with different AI services through a consistent API.

## Key Components

### Base Provider

The base provider implements common functionality shared across all providers:

```go
type BaseProvider struct {
	tools []tools.Tool
}
```

### Provider Implementations

The package includes implementations for specific AI providers:

#### Claude Provider

```go
client := claude.NewClient("your-api-key")
client.SetModel("claude-3-sonnet-20240229")
client.SetMaxTokens(1000)
```

#### OpenAI Provider

```go
client := openai.NewClient("your-api-key")
client.SetModel("gpt-4")
client.SetMaxTokens(1000)
```

### Common HTTP Client

The `common` sub-package provides a shared HTTP client with proper configuration for API calls:

```go
httpClient := common.NewClient()
```

## Provider Capabilities

All providers implement the `models.Provider` interface:

1. **Basic Message Sending**:
   ```go
   response, err := provider.SendMessage(ctx, models.Message{
       Role:    models.RoleUser,
       Content: "Hello, how can you help me today?",
   })
   ```

2. **Conversation History**:
   ```go
   response, err := provider.SendMessageWithHistory(ctx, []models.Message{
       {Role: models.RoleSystem, Content: "You are a helpful assistant."},
       {Role: models.RoleUser, Content: "What's the weather like?"},
       {Role: models.RoleAssistant, Content: "I don't have access to current weather data."},
       {Role: models.RoleUser, Content: "What can you help me with then?"},
   })
   ```

3. **Function Calling / Tools**:
   ```go
   // Register a tool
   provider.RegisterTool(weatherTool)
   
   // Send a message that might trigger tool use
   response, err := provider.SendMessageWithTools(ctx, models.Message{
       Role:    models.RoleUser, 
       Content: "What's the weather in New York?",
   })
   ```

## Provider Configuration

Providers typically support configuration options:

```go
// Set the model to use
provider.SetModel("claude-3-opus-20240229")

// Set maximum output tokens
provider.SetMaxTokens(2000)

// Set temperature (randomness)
provider.SetTemperature(0.7)

// Set system prompt
provider.SetSystemPrompt("You are a specialized assistant for weather forecasting.")
```

## Example Usage

```go
// Create a provider
apiKey := os.Getenv("ANTHROPIC_API_KEY")
provider := claude.NewClient(apiKey)
provider.SetModel("claude-3-sonnet-20240229")

// Register a tool
weatherTool := tools.NewTool(...)
provider.RegisterTool(weatherTool)

// Send a message
response, err := provider.SendMessageWithTools(ctx, models.Message{
    Role:    models.RoleUser,
    Content: "What's the weather in London?",
})
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Println(response)
```