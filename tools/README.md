# Bond Tools

The `tools` package provides a framework for creating and managing tools that can be used by AI providers. Tools allow the AI to interact with external systems and perform specific functions, greatly expanding its capabilities.

## Key Components

### ToolExecutor Interface

The central interface that all tools must implement:

```go
type ToolExecutor interface {
	IsNamespaced() bool
	Namespace(string)
	GetName() string
	GetDescription() string
	GetSchema() models.InputSchema
	Execute(input json.RawMessage) (string, error)
}
```

### Base Tool Implementation

The package provides a base implementation of the ToolExecutor interface:

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

### Namespaced Tools

Tools can be namespaced to avoid name collisions and to indicate their source:

```go
// Create a tool
weatherTool := tools.NewTool(...)

// Namespace the tool
weatherTool.Namespace("weather")

// The tool name is now "weather:get_weather"
name := weatherTool.GetName() // Returns "weather:get_weather"
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

### Integration with MCP

Tools can be integrated with the Model Context Protocol:

```go
// Create an MCP client
mcpClient := mcp.New("orchestra", nil)

// Initialize the MCP
mcpClient.Initialise()

// Get tools from the MCP server
toolList, _ := mcpClient.ListTools()

// Register MCP tools with a provider
provider := claude.NewClient(apiKey)
provider.RegisterMCP("orchestra", nil)

// The provider can now use tools from the MCP server
response, err := provider.SendMessageWithTools(ctx, models.Message{
    Role:    models.RoleUser,
    Content: "Use the orchestra:get_data tool to retrieve information.",
})
```