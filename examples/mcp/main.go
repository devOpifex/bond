package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/devOpifex/bond/mcp"
)

func main() {
	// Create a new MCP instance with mcpbrowser command
	mcpInstance := mcp.New("mcpbrowser", nil)

	// Set a timeout for requests
	mcpInstance.SetDefaultTimeout(5 * time.Second)

	fmt.Println("Initializing MCP client with capabilities...")

	// Initialize capabilities (starts MCP if not running)
	capabilities, err := mcpInstance.Initialise()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing capabilities: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Println("Server capabilities:")
		capabilitiesJSON, _ := json.MarshalIndent(capabilities, "", "  ")
		fmt.Println(string(capabilitiesJSON))
	}

	fmt.Println("\nFetching available tools...")

	// List available tools
	toolList, err := mcpInstance.ListTools()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing tools: %v\n", err)
		os.Exit(1)
	}

	// Print the tool list
	fmt.Printf("Found %d tools:\n", len(toolList.Tools))
	for i, tool := range toolList.Tools {
		fmt.Printf("%d. %s - %s\n", i+1, tool.Name, tool.Description)
	}

	// Try to find and call the browser tool if available
	for _, tool := range toolList.Tools {
		if tool.Name == "browser" {
			fmt.Printf("\nCalling browser tool: %s\n", tool.Name)

			// Call the browser tool
			arguments := map[string]any{
				"url": "https://www.opifex.org",
			}

			result, err := mcpInstance.CallTool(tool.Name, arguments)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error calling browser tool: %v\n", err)
			} else {
				fmt.Println("Tool result:")
				resultJSON, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(resultJSON))
			}
			break
		}
	}

	fmt.Println("\nStopping MCP client...")

	// Stop the MCP process
	if err := mcpInstance.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "Error stopping MCP: %v\n", err)
	}

	fmt.Println("MCP client stopped.")
}
