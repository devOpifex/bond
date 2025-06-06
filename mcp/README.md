# Bond MCP

The `mcp` package implements support for the Model Context Protocol (MCP), enabling Bond to integrate with external tool servers.

## Key Components

### MCP Client

The main client for interacting with MCP servers:

```go
// Create a new MCP client
mcp := mcp.New("command", []string{"arg1", "arg2"})

// With custom IO
mcpCustom := mcp.NewMCP(stdin, stdout, stderr, "command", []string{"args"})
```

### JSON-RPC Communication

The package uses JSON-RPC 2.0 for communication with MCP servers:

```go
// Create a new request
request := mcp.NewRequest("method", params, id)

// Send a request to the MCP server
response, err := mcpClient.Call("method", params)
```

### Tool Management

The MCP client can discover and call tools on MCP servers:

```go
// List available tools
toolList, err := mcpClient.ListTools()

// Call a tool
result, err := mcpClient.CallTool("tool_name", map[string]any{
    "param1": "value1",
    "param2": 42,
})
```

## MCP Server Capabilities

The MCP client can query a server's capabilities:

```go
// Get server capabilities
capabilities, err := mcpClient.GetCapabilities()

// Check if the server supports tool list change notifications
if capabilities.Tools.ListChanged {
    // Handle tool list changes
}
```

## Handler Registration

Register handlers for specific MCP methods:

```go
// Register a handler for tool list change notifications
mcpClient.RegisterHandler("notifications/tools/list_changed", func(response *mcp.Response) {
    // Refresh tool list
    toolList, _ := mcpClient.ListTools()
    fmt.Printf("Tool list updated, %d tools available\n", len(toolList.Tools))
})
```

## Example Usage

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/devOpifex/bond/mcp"
)

func main() {
    // Create an MCP client
    mcpClient := mcp.New("orchestra", nil)

    // Set a default timeout for requests
    mcpClient.SetDefaultTimeout(30 * time.Second)

    // Initialize the MCP client
    capabilities, err := mcpClient.Initialise()
    if err != nil {
        log.Fatalf("Failed to initialize MCP: %v", err)
    }

    // List available tools
    toolList, err := mcpClient.ListTools()
    if err != nil {
        log.Fatalf("Failed to list tools: %v", err)
    }

    fmt.Printf("Found %d tools\n", len(toolList.Tools))
    for _, tool := range toolList.Tools {
        fmt.Printf("- %s: %s\n", tool.Name, tool.Description)
    }

    // Call a tool
    result, err := mcpClient.CallTool("get_weather", map[string]any{
        "location": "London",
    })
    if err != nil {
        log.Fatalf("Failed to call tool: %v", err)
    }

    fmt.Printf("Tool result: %s\n", result.Result)
    
    // Stop the MCP client when done
    mcpClient.Stop()
}
```

## Integration with Providers

The MCP package is designed to be integrated with Bond providers:

```go
// Create a provider
provider := claude.NewClient(os.Getenv("ANTHROPIC_API_KEY"))

// Register an MCP server with the provider
provider.RegisterMCP("orchestra", nil)

// Use the provider as usual, with access to MCP tools
response, err := provider.SendMessageWithTools(ctx, models.Message{
    Role:    models.RoleUser,
    Content: "Please check the weather in London using the weather tool.",
})
```