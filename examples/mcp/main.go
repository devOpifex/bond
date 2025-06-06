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

	claude.RegisterMCP("orchestra", nil)

	ctx := context.Background()

	response, err := claude.SendMessageWithTools(ctx, models.Message{
		Role:    models.RoleUser,
		Content: "Get the codelist for the core_dpp. use orchestra:get_codelists",
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(response)
}
