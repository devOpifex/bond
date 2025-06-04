package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/devOpifex/bond/mcp"
)

func main() {
	// Create a new MCP instance with mcpOrchestra command
	// Replace "mcpOrchestra" with your actual MCP-compatible server command
	mcpInstance := mcp.New("mcpbrowser", nil)

	// Set a timeout for requests
	mcpInstance.SetDefaultTimeout(5 * time.Second)

	fmt.Println("Starting MCP client...")

	// Start the MCP process
	if err := mcpInstance.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting MCP: %v\n", err)
		os.Exit(1)
	}

	// Get server capabilities
	capabilities, err := mcpInstance.GetCapabilities()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: couldn't get server capabilities: %v\n", err)
	} else {
		fmt.Println("Server capabilities:")
		capabilitiesJSON, _ := json.MarshalIndent(capabilities, "", "  ")
		fmt.Println(string(capabilitiesJSON))
	}

	fmt.Println("\nFetching available tools...")

	// List available tools
	toolList, err := mcpInstance.ListTools("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing tools: %v\n", err)
		os.Exit(1)
	}

	// Print the tool list
	fmt.Printf("Found %d tools:\n", len(toolList.Tools))
	for i, tool := range toolList.Tools {
		fmt.Printf("%d. %s - %s\n", i+1, tool.Name, tool.Description)
	}

	// If we have at least one tool, try to call it
	if len(toolList.Tools) > 0 {
		exampleTool := toolList.Tools[0]
		fmt.Printf("\nCalling tool: %s\n", exampleTool.Name)

		// Parse the input schema to understand required arguments
		var schemaObj map[string]interface{}
		if err := json.Unmarshal(exampleTool.InputSchema, &schemaObj); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing input schema: %v\n", err)
		}

		// For demo purposes, we'll print the schema
		fmt.Println("Tool input schema:")
		schemaJSON, _ := json.MarshalIndent(schemaObj, "", "  ")
		fmt.Println(string(schemaJSON))

		// This is a very simple example that would need to be adapted to the actual tool
		// For demo purposes, we're just creating an empty arguments map
		// In a real application, you'd extract the required parameters from the schema
		// and provide appropriate values
		arguments := map[string]any{
			"url": "https://www.opifex.org",
		}

		// Comment out the actual tool call since we don't have real arguments
		// Uncomment and adapt this code when you have a real MCP server and know the required arguments
		result, err := mcpInstance.CallTool(exampleTool.Name, arguments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error calling tool: %v\n", err)
		} else {
			fmt.Println("Tool result:")
			resultJSON, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(resultJSON))
		}
	}

	fmt.Println("\nStopping MCP client...")

	// Stop the MCP process
	if err := mcpInstance.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "Error stopping MCP: %v\n", err)
	}

	fmt.Println("MCP client stopped.")
}
