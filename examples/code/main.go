package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/devOpifex/bond/agent"
	"github.com/devOpifex/bond/models"
	"github.com/devOpifex/bond/providers/claude"
	"github.com/devOpifex/bond/reasoning"
	"github.com/devOpifex/bond/tools"
)

// LanguageCodeGenerator is an agent that generates code for a specific language
type LanguageCodeGenerator struct {
	Language string
}

func (c *LanguageCodeGenerator) Process(ctx context.Context, input string) (string, error) {
	switch c.Language {
	case "python":
		return "<PYTHON CODE>", nil
	case "javascript":
		return "<JAVASCRIPT CODE>", nil
	case "go":
		return "<GO CODE>", nil
	default:
		return fmt.Sprintf("Sorry, I don't know how to generate %s code yet.", c.Language), nil
	}
}

// CodeExplainer is an agent that explains code
type CodeExplainer struct{}

func (c *CodeExplainer) Process(ctx context.Context, input string) (string, error) {
	// In a real implementation, this would use a specialized model
	// Here we're just doing some simple pattern matching

	if strings.Contains(input, "fibonacci") {
		return "This is fibonacci code", nil
	}

	return "This is code", nil
}

// BenchmarkAnalyzer is an agent that analyzes code performance
type BenchmarkAnalyzer struct{}

func (b *BenchmarkAnalyzer) Process(ctx context.Context, input string) (string, error) {
	// In a real implementation, this would use actual benchmarking tools
	// Here we're just simulating an analysis

	if strings.Contains(input, "fibonacci") && strings.Contains(input, "python") {
		return "This is fibonacci Python code, people like it", nil
	} else if strings.Contains(input, "fibonacci") && strings.Contains(input, "go") {
		return "This is fibonacci Go code, people love it", nil
	}

	return "This is code", nil
}

// CodeImprover is an agent that suggests improvements to code
type CodeImprover struct{}

func (i *CodeImprover) Process(ctx context.Context, input string) (string, error) {
	// In a real implementation, this would use a specialized model
	// Here we're just doing some simple pattern matching

	if strings.Contains(input, "fibonacci") {
		return "This is improved fibonacci code", nil
	}

	return "This is improved code", nil
}

// extractCodeLanguage extracts the programming language from the input
func extractCodeLanguage(input string) string {
	pythonPattern := regexp.MustCompile(`(?i)python`)
	goPattern := regexp.MustCompile(`(?i)\bgo\b`)
	jsPattern := regexp.MustCompile(`(?i)javascript|js`)

	if pythonPattern.MatchString(input) {
		return "python"
	} else if goPattern.MatchString(input) {
		return "go"
	} else if jsPattern.MatchString(input) {
		return "javascript"
	}

	return "python" // Default to Python if no language detected
}

// extractCodeTask extracts the task from the input
func extractCodeTask(input string) string {
	// Remove language references to isolate the task
	input = regexp.MustCompile(`(?i)python|javascript|js|\bgo\b`).ReplaceAllString(input, "")

	// Look for keywords like "to", "that", "which", "for"
	taskPattern := regexp.MustCompile(`(?i)(create|write|implement|generate|code for|function for|program for|to|that|which)\s+(.+)`)
	matches := taskPattern.FindStringSubmatch(input)

	if len(matches) > 2 {
		return matches[2]
	}

	// If no clear task is found, return the whole input
	return input
}

func main() {
	// Create a provider (Claude in this example)
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: ANTHROPIC_API_KEY environment variable not set")
		os.Exit(1)
	}

	provider := claude.NewClient(apiKey)

	// Configure the provider
	provider.SetModel("claude-3-sonnet-20240229")
	provider.SetMaxTokens(1000)

	// Create an agent manager
	agentManager := agent.NewAgentManager()

	// Register language-specific code generators
	agentManager.RegisterAgent("python-code", &LanguageCodeGenerator{Language: "python"})
	agentManager.RegisterAgent("javascript-code", &LanguageCodeGenerator{Language: "javascript"})
	agentManager.RegisterAgent("go-code", &LanguageCodeGenerator{Language: "go"})

	// Register specialized agents for code understanding
	agentManager.RegisterAgent("code-explainer", &CodeExplainer{})
	agentManager.RegisterAgent("benchmark-analyzer", &BenchmarkAnalyzer{})
	agentManager.RegisterAgent("code-improver", &CodeImprover{})

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

	// Example usage of multi-step reasoning
	ctx := context.Background()

	// This is our input query from the user
	userQuery := "Create a fibonacci function in Go and tell me how I could improve it"

	fmt.Println("User query:", userQuery)
	fmt.Println("-----------------------------------")
	fmt.Println("Starting multi-step reasoning workflow...")

	// Create a workflow for complex multi-step processing
	workflow := reasoning.NewWorkflow()

	// Step 1: Extract code language and task
	workflow.AddStep(reasoning.ProcessorStep(
		"Extract language and task from query",
		"Extracts programming language and task description from the user query",
		func(ctx context.Context, input string, memory *reasoning.Memory) (string, map[string]interface{}, error) {
			language := extractCodeLanguage(input)
			task := extractCodeTask(input)

			memory.Set("language", language)
			memory.Set("task", task)

			return fmt.Sprintf("Language: %s\nTask: %s", language, task),
				map[string]interface{}{
					"language": language,
					"task":     task,
				}, nil
		},
	)).
	// Step 2: Generate code based on extracted info
	Then(&reasoning.Step{
		Name:        "Generate Code",
		Description: "Generates code in the requested language",
		Execute: func(ctx context.Context, _ string, memory *reasoning.Memory) (reasoning.StepResult, error) {
			language, _ := memory.GetString("language")
			task, _ := memory.GetString("task")

			capability := fmt.Sprintf("%s-code", language)

			code, err := agentManager.ProcessWithBestAgent(ctx, capability, task)
			if err != nil {
				return reasoning.StepResult{Error: err}, err
			}

			return reasoning.StepResult{
				Output: code,
				Metadata: map[string]interface{}{
					"language": language,
				},
			}, nil
		},
	}).
	// Step 3: Explain the generated code
	Then(&reasoning.Step{
		Name:        "Explain Code",
		Description: "Explains the generated code",
		Execute: func(ctx context.Context, input string, memory *reasoning.Memory) (reasoning.StepResult, error) {
			explanation, err := agentManager.ProcessWithBestAgent(ctx, "code-explainer", input)
			if err != nil {
				return reasoning.StepResult{Error: err}, err
			}

			return reasoning.StepResult{
				Output: explanation,
			}, nil
		},
	}).
	// Step 4: Analyze performance
	Then(&reasoning.Step{
		Name:        "Analyze Performance",
		Description: "Analyzes code performance",
		Execute: func(ctx context.Context, input string, memory *reasoning.Memory) (reasoning.StepResult, error) {
			language, _ := memory.GetString("language")

			analysisInput := fmt.Sprintf("%s\nLanguage: %s", input, language)
			analysis, err := agentManager.ProcessWithBestAgent(ctx, "benchmark-analyzer", analysisInput)
			if err != nil {
				return reasoning.StepResult{Error: err}, err
			}

			return reasoning.StepResult{
				Output: analysis,
			}, nil
		},
	}).
	// Step 5: Suggest improvements
	Then(&reasoning.Step{
		Name:        "Suggest Improvements",
		Description: "Suggests improvements to the code",
		Execute: func(ctx context.Context, input string, memory *reasoning.Memory) (reasoning.StepResult, error) {
			// We can access outputs from step indexes
			codeOutput, _ := memory.GetString("step.step_1.output")
			performanceAnalysis, _ := memory.GetString("step.step_3.output")

			improvementInput := fmt.Sprintf("%s\n\nPerformance Analysis:\n%s", codeOutput, performanceAnalysis)
			improvements, err := agentManager.ProcessWithBestAgent(ctx, "code-improver", improvementInput)
			if err != nil {
				return reasoning.StepResult{Error: err}, err
			}

			return reasoning.StepResult{
				Output: improvements,
			}, nil
		},
	}).
	// Step 6: Generate final report
	Then(&reasoning.Step{
		Name:        "Generate Report",
		Description: "Generates final comprehensive report",
		Execute: func(ctx context.Context, input string, memory *reasoning.Memory) (reasoning.StepResult, error) {
			language, _ := memory.GetString("language")
			task, _ := memory.GetString("task")
			code, _ := memory.GetString("step.step_1.output")
			explanation, _ := memory.GetString("step.step_2.output")
			performance, _ := memory.GetString("step.step_3.output")
			improvements, _ := memory.GetString("step.step_4.output")

			report := fmt.Sprintf(
				"# Code Solution Report\n\n"+
					"## Task\n%s\n\n"+
					"## %s Implementation\n%s\n\n"+
					"## Explanation\n%s\n\n"+
					"## Performance Analysis\n%s\n\n"+
					"## Suggested Improvements\n%s\n\n",
				task, strings.Title(language), code, explanation, performance, improvements,
			)

			return reasoning.StepResult{
				Output: report,
				Metadata: map[string]interface{}{
					"timestamp": time.Now().Format(time.RFC3339),
				},
			}, nil
		},
	})

	// Execute the workflow
	result, err := workflow.Execute(ctx, userQuery)
	if err != nil {
		fmt.Printf("Workflow error: %v\n", err)
		return
	}

	fmt.Println(result)

	// For comparison, use the traditional tool approach
	fmt.Println("-----------------------------------")
	fmt.Println("Now trying the same with traditional tool approach...")

	// Traditional tool-based approach
	response, err := provider.SendMessageWithTools(ctx, models.Message{
		Role:    models.RoleUser,
		Content: userQuery,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("-----------------------------------")
	fmt.Printf("Claude's response:\n%s\n", response)
}