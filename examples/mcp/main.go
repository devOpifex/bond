package main

import (
	"fmt"
	"os"

	"github.com/devOpifex/bond/mcp"
)

func main() {
	mcpInstance := mcp.New("mcpOrchestra", []string{})

	// Start the MCP, which will execute the command
	if err := mcpInstance.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

