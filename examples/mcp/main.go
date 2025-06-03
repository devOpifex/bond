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
	mcpInstance := mcp.New("mcpOrchestra", []string{})

	// Set a timeout for requests
	mcpInstance.SetDefaultTimeout(5 * time.Second)

	fmt.Println("Starting MCP client...")

	// Start the MCP process
	if err := mcpInstance.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting MCP: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("MCP client started. Sending tools/list request...")

	// Send a JSON-RPC request for tools/list method
	response, err := mcpInstance.Call("tools/list", nil)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error calling tools/list method: %v\n", err)
		os.Exit(1)
	}

	// Print the response in a formatted way
	fmt.Println("Received response from tools/list:")
	if response.Error != nil {
		fmt.Printf("Error: %s (code: %d)\n", response.Error.Message, response.Error.Code)
		if response.Error.Data != nil {
			fmt.Printf("Error data: %v\n", response.Error.Data)
		}
	} else {
		// Pretty print the result
		resultJSON, err := json.MarshalIndent(response.Result, "", "  ")
		if err != nil {
			fmt.Printf("Result: %v\n", response.Result)
		} else {
			fmt.Printf("%s\n", resultJSON)
		}
	}

	fmt.Println("Stopping MCP client...")

	// Stop the MCP process
	if err := mcpInstance.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "Error stopping MCP: %v\n", err)
	}

	fmt.Println("MCP client stopped.")
}
