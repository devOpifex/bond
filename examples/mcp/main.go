package main

import (
	"context"
	"fmt"
	"os"

	"github.com/devOpifex/bond/models"
	"github.com/devOpifex/bond/providers/claude"
)

func main() {
	claude := claude.New(os.Getenv("ANTHROPIC_API_KEY"))

	claude.RegisterMCP("mcpOrchestra", nil)

	ctx := context.Background()

	response, err := claude.SendMessageWithTools(ctx, models.Message{
		Role:    models.RoleUser,
		Content: "Find John's email address using mcpOrchestra:search_email",
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(response)
}
