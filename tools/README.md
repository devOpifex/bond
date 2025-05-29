# Bond Tools

The `tools` package provides a framework for creating and managing tools that can be used by AI providers. Tools allow the AI to interact with external systems and perform specific functions, greatly expanding its capabilities.

## Key Components

### Tool Interface

The central interface that all tools must implement:

```go
type Tool interface {
	GetName() string
	GetDescription() string
	GetInputSchema() models.InputSchema
	Execute(params map[string]interface{}) (string, error)
}
```

### Base Tool Implementation

The package provides a base implementation of the Tool interface:

```go
tool := tools.NewTool(
    "calculator",
    "Performs basic arithmetic operations",
    models.InputSchema{
        Type: "object",
        Properties: map[string]models.Property{
            "operation": {
                Type:        "string",
                Description: "The operation to perform (add, subtract, multiply, divide)",
                Enum:        []string{"add", "subtract", "multiply", "divide"},
            },
            "a": {
                Type:        "number",
                Description: "The first number",
            },
            "b": {
                Type:        "number",
                Description: "The second number",
            },
        },
        Required: []string{"operation", "a", "b"},
    },
    func(params map[string]interface{}) (string, error) {
        // Implementation of the calculator logic
        // ...
    },
)
```

### Tool Registry

Allows registration and lookup of tools:

```go
registry := tools.NewRegistry()
registry.Register(calculatorTool)
registry.Register(weatherTool)

// Look up a tool by name
tool, exists := registry.Get("calculator")
```

### Agent Tools

The package supports creating tools that delegate to agents:

```go
// Create an agent tool
agentTool := tools.NewAgentTool(
    "code_generator",
    "Generates code in various programming languages",
    models.InputSchema{...},
    agentManager,
    "code-generation-capability"
)
```

## Example Usage

### Creating a Weather Tool

```go
weatherTool := tools.NewTool(
    "get_weather",
    "Gets the current weather for a location",
    models.InputSchema{
        Type: "object",
        Properties: map[string]models.Property{
            "location": {
                Type:        "string",
                Description: "The city and country (e.g., 'London, UK')",
            },
        },
        Required: []string{"location"},
    },
    func(params map[string]interface{}) (string, error) {
        location, _ := params["location"].(string)
        // Call a weather API to get the weather for the location
        weather := fetchWeatherData(location)
        return fmt.Sprintf("The weather in %s is %s with a temperature of %.1f°C", 
            location, weather.Condition, weather.Temperature), nil
    },
)
```

### Registering Tools with a Provider

```go
// Create tools
calculatorTool := tools.NewTool(...)
weatherTool := tools.NewTool(...)

// Create a provider
provider := claude.NewClient(apiKey)

// Register tools with the provider
provider.RegisterTool(calculatorTool)
provider.RegisterTool(weatherTool)

// Now when the AI is asked about weather or calculations,
// it can use these tools to provide accurate responses
response, err := provider.SendMessageWithTools(ctx, models.Message{
    Role:    models.RoleUser,
    Content: "What's the weather in Paris? Also, what's 1342 * 5938?",
})
```

### Using the Registry

```go
// Create a registry
registry := tools.NewRegistry()

// Register multiple tools
registry.Register(calculatorTool)
registry.Register(weatherTool)
registry.Register(translationTool)

// Look up a tool by name
tool, exists := registry.Get("calculator")
if exists {
    result, err := tool.Execute(map[string]interface{}{
        "operation": "multiply",
        "a": 5,
        "b": 7,
    })
    // result will be "35"
}
```