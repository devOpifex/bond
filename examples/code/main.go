package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/devOpifex/bond/agent"
	"github.com/devOpifex/bond/models"
	"github.com/devOpifex/bond/providers"
	"github.com/devOpifex/bond/tools"
)

// LanguageCodeGenerator is an agent that generates code for a specific language
type LanguageCodeGenerator struct {
	Language string
}

func (c *LanguageCodeGenerator) Process(ctx context.Context, input string) (string, error) {
	// In a real implementation, this would use a specialized model or service
	// tailored to the specific language
	switch c.Language {
	case "python":
		return generatePythonCode(input), nil
	case "javascript":
		return generateJavaScriptCode(input), nil
	case "go":
		return generateGoCode(input), nil
	default:
		return fmt.Sprintf("Sorry, I don't know how to generate %s code yet.", c.Language), nil
	}
}

// Mock implementations of language-specific code generators
func generatePythonCode(task string) string {
	if strings.Contains(task, "fibonacci") {
		return "```python\ndef fibonacci(n):\n    if n <= 1:\n        return n\n    return fibonacci(n-1) + fibonacci(n-2)\n\n# Example usage\nfor i in range(10):\n    print(fibonacci(i))\n```"
	}
	return "```python\n# Python implementation for: " + task + "\ndef solution():\n    print('Implementation would go here')\n```"
}

func generateJavaScriptCode(task string) string {
	if strings.Contains(task, "fibonacci") {
		return "```javascript\nfunction fibonacci(n) {\n    if (n <= 1) return n;\n    return fibonacci(n-1) + fibonacci(n-2);\n}\n\n// Example usage\nfor (let i = 0; i < 10; i++) {\n    console.log(fibonacci(i));\n}\n```"
	}
	return "```javascript\n// JavaScript implementation for: " + task + "\nfunction solution() {\n    console.log('Implementation would go here');\n}\n```"
}

func generateGoCode(task string) string {
	if strings.Contains(task, "fibonacci") {
		return "```go\npackage main\n\nimport \"fmt\"\n\nfunc fibonacci(n int) int {\n    if n <= 1 {\n        return n\n    }\n    return fibonacci(n-1) + fibonacci(n-2)\n}\n\nfunc main() {\n    for i := 0; i < 10; i++ {\n        fmt.Println(fibonacci(i))\n    }\n}\n```"
	}
	return "```go\npackage main\n\nimport \"fmt\"\n\n// Go implementation for: " + task + "\nfunc solution() {\n    fmt.Println(\"Implementation would go here\")\n}\n\nfunc main() {\n    solution()\n}\n```"
}

func main() {
	// Create a provider (Claude in this example)
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: ANTHROPIC_API_KEY environment variable not set")
		os.Exit(1)
	}

	provider, err := providers.NewProvider(providers.Claude, apiKey)
	if err != nil {
		fmt.Printf("Error creating provider: %v\n", err)
		return
	}

	// Configure the provider
	provider.SetModel("claude-3-sonnet-20240229")
	provider.SetMaxTokens(1000)

	// Create an agent manager
	agentManager := agent.NewAgentManager()

	// Register language-specific code generators
	agentManager.RegisterAgent("python-code", &LanguageCodeGenerator{Language: "python"})
	agentManager.RegisterAgent("javascript-code", &LanguageCodeGenerator{Language: "javascript"})
	agentManager.RegisterAgent("go-code", &LanguageCodeGenerator{Language: "go"})

	// Create a code generation tool that routes to the appropriate agent
	codeGenTool := tools.NewTool(
		"generate_code",
		"Generate code in a specified programming language",
		models.InputSchema{
			Type: "object",
			Properties: map[string]models.Property{
				"language": {
					Type:        "string",
					Description: "The programming language (python, javascript, go)",
				},
				"task": {
					Type:        "string",
					Description: "Description of what the code should do",
				},
			},
			Required: []string{"language", "task"},
		},
		func(params map[string]interface{}) (string, error) {
			language, _ := params["language"].(string)
			task, _ := params["task"].(string)

			// Determine which agent to use based on language
			capability := fmt.Sprintf("%s-code", language)

			return agentManager.ProcessWithBestAgent(
				context.Background(),
				capability,
				task,
			)
		},
	)

	// Register the tool with the provider
	provider.RegisterTool(codeGenTool)

	// Example usage
	ctx := context.Background()
	prompt := "I need to calculate Fibonacci numbers. Can you write this in Python and Go for comparison?"

	fmt.Println("Sending prompt to Claude:", prompt)
	fmt.Println("-----------------------------------")

	response, err := provider.SendMessageWithTools(ctx, prompt)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Claude's response:\n%s\n", response)
}

