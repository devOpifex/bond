package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/devOpifex/bond/mcp"
	"github.com/devOpifex/bond/models"
	"github.com/devOpifex/bond/providers"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable not set")
	}

	// Create a provider
	provider, err := providers.NewProvider(providers.Claude, apiKey)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// Configure the provider
	provider.SetModel("claude-3-sonnet-20240229")
	provider.SetMaxTokens(1000)

	// Create and register a stdio MCP
	stdioMCPConfig := models.MCPConfig{
		ID:      "orchestra-mcp",
		Type:    models.MCPTypeStdio,
		Command: "mcpOrchestra",
	}

	stdioMCP, err := mcp.NewMCPProvider(stdioMCPConfig)
	if err != nil {
		log.Fatalf("Failed to create stdio MCP: %v", err)
	}

	fmt.Println("Initializing and registering the MCP...")
	err = provider.RegisterMCP(stdioMCP)
	if err != nil {
		log.Fatalf("Failed to register stdio MCP: %v", err)
	}
	fmt.Println("Successfully registered MCP")

	// Display example message
	exampleMessage := models.Message{
		Role:    models.RoleUser,
		Content: "Can you retrieve John's email address?",
	}

	ctx := context.Background()

	// Send the message with tools
	response, err := provider.SendMessageWithTools(ctx, exampleMessage)
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	fmt.Println("\nResponse:", response)

	// Clean up
	fmt.Println("\nShutting down the MCP...")
	if err := stdioMCP.Shutdown(); err != nil {
		log.Printf("Warning: Failed to shut down MCP: %v", err)
	}

	fmt.Println("Example completed successfully")
}
