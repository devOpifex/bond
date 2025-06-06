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
	claude.SetTemperature(0.8)

	ctx := context.Background()

	response, err := claude.SendMessageWithTools(ctx, models.Message{
		Role:    models.RoleUser,
		Content: "Does the core_dpp study have the ARM variable in its codelist?",
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(response)
}
