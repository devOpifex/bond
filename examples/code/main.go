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
	"github.com/devOpifex/bond/providers"
	"github.com/devOpifex/bond/reasoning"
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

// CodeExplainer is an agent that explains code
type CodeExplainer struct{}

func (c *CodeExplainer) Process(ctx context.Context, input string) (string, error) {
	// In a real implementation, this would use a specialized model
	// Here we're just doing some simple pattern matching

	if strings.Contains(input, "fibonacci") {
		return "This code implements the Fibonacci sequence, where each number is the sum of the two preceding ones. " +
			"It uses recursion to calculate each number. The time complexity is O(2^n) which is inefficient for large values.", nil
	}

	return "This code implements a solution for the specified task. It includes the necessary structure and placeholders for implementation.", nil
}

// BenchmarkAnalyzer is an agent that analyzes code performance
type BenchmarkAnalyzer struct{}

func (b *BenchmarkAnalyzer) Process(ctx context.Context, input string) (string, error) {
	// In a real implementation, this would use actual benchmarking tools
	// Here we're just simulating an analysis

	if strings.Contains(input, "fibonacci") && strings.Contains(input, "python") {
		return "Python Fibonacci (recursive):\n" +
			"- Time complexity: O(2^n)\n" +
			"- Space complexity: O(n) due to call stack\n" +
			"- Benchmark: ~15ms for fibonacci(20)", nil
	} else if strings.Contains(input, "fibonacci") && strings.Contains(input, "go") {
		return "Go Fibonacci (recursive):\n" +
			"- Time complexity: O(2^n)\n" +
			"- Space complexity: O(n) due to call stack\n" +
			"- Benchmark: ~5ms for fibonacci(20)", nil
	}

	return "Performance analysis:\n" +
		"- Time complexity: Depends on implementation\n" +
		"- Space complexity: Minimal memory usage\n" +
		"- Consider optimizing for your specific use case", nil
}

// CodeImprover is an agent that suggests improvements to code
type CodeImprover struct{}

func (i *CodeImprover) Process(ctx context.Context, input string) (string, error) {
	// In a real implementation, this would use a specialized model
	// Here we're just doing some simple pattern matching

	if strings.Contains(input, "fibonacci") {
		return "Improvement suggestions:\n\n" +
			"1. Use memoization to avoid redundant calculations:\n\n" +
			"```python\ndef fibonacci_memo(n, memo={}):\n" +
			"    if n in memo:\n" +
			"        return memo[n]\n" +
			"    if n <= 1:\n" +
			"        return n\n" +
			"    memo[n] = fibonacci_memo(n-1, memo) + fibonacci_memo(n-2, memo)\n" +
			"    return memo[n]\n```\n\n" +
			"2. Or use an iterative approach which is even more efficient:\n\n" +
			"```python\ndef fibonacci_iter(n):\n" +
			"    if n <= 1:\n" +
			"        return n\n" +
			"    a, b = 0, 1\n" +
			"    for _ in range(2, n+1):\n" +
			"        a, b = b, a+b\n" +
			"    return b\n```", nil
	}

	return "Consider adding error handling, parameter validation, and documentation to make your code more robust.", nil
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
	userQuery := "Create a Fibonacci function in Python and tell me how I could improve it"

	fmt.Println("User query:", userQuery)
	fmt.Println("-----------------------------------")
	fmt.Println("Starting multi-step reasoning workflow...")

	// Create a workflow for complex multi-step processing
	workflow := reasoning.NewWorkflow()

	// Step 1: Extract code language and task
	workflow.AddStep(reasoning.ProcessorStep(
		"extract-info",
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
	))

	// Step 2: Generate code based on extracted info
	workflow.AddStep(&reasoning.Step{
		ID:          "generate-code",
		Name:        "Generate Code",
		Description: "Generates code in the requested language",
		DependsOn:   []string{"extract-info"},
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
	})

	// Step 3: Explain the generated code
	workflow.AddStep(&reasoning.Step{
		ID:          "explain-code",
		Name:        "Explain Code",
		Description: "Explains the generated code",
		DependsOn:   []string{"generate-code"},
		Execute: func(ctx context.Context, input string, memory *reasoning.Memory) (reasoning.StepResult, error) {
			explanation, err := agentManager.ProcessWithBestAgent(ctx, "code-explainer", input)
			if err != nil {
				return reasoning.StepResult{Error: err}, err
			}

			return reasoning.StepResult{
				Output: explanation,
			}, nil
		},
	})

	// Step 4: Analyze performance
	workflow.AddStep(&reasoning.Step{
		ID:          "analyze-performance",
		Name:        "Analyze Performance",
		Description: "Analyzes code performance",
		DependsOn:   []string{"generate-code"},
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
	})

	// Step 5: Suggest improvements
	workflow.AddStep(&reasoning.Step{
		ID:          "suggest-improvements",
		Name:        "Suggest Improvements",
		Description: "Suggests improvements to the code",
		DependsOn:   []string{"generate-code", "analyze-performance"},
		Execute: func(ctx context.Context, input string, memory *reasoning.Memory) (reasoning.StepResult, error) {
			codeOutput, _ := memory.GetString("step.generate-code.output")
			performanceAnalysis, _ := memory.GetString("step.analyze-performance.output")

			improvementInput := fmt.Sprintf("%s\n\nPerformance Analysis:\n%s", codeOutput, performanceAnalysis)
			improvements, err := agentManager.ProcessWithBestAgent(ctx, "code-improver", improvementInput)
			if err != nil {
				return reasoning.StepResult{Error: err}, err
			}

			return reasoning.StepResult{
				Output: improvements,
			}, nil
		},
	})

	// Step 6: Generate final report (simplified dependencies)
	workflow.AddStep(&reasoning.Step{
		ID:          "generate-report",
		Name:        "Generate Report",
		Description: "Generates final comprehensive report",
		// Only depend on suggest-improvements and explain-code to avoid potential cycles
		DependsOn:   []string{"suggest-improvements", "explain-code"},
		Execute: func(ctx context.Context, input string, memory *reasoning.Memory) (reasoning.StepResult, error) {
			language, _ := memory.GetString("language")
			task, _ := memory.GetString("task")
			code, _ := memory.GetString("step.generate-code.output")
			explanation, _ := memory.GetString("step.explain-code.output")
			performance, _ := memory.GetString("step.analyze-performance.output")
			improvements, _ := memory.GetString("step.suggest-improvements.output")

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
	start := time.Now()
	result, err := workflow.Execute(ctx, userQuery, "generate-report")
	if err != nil {
		fmt.Printf("Workflow error: %v\n", err)
		return
	}
	elapsed := time.Since(start)

	fmt.Println("-----------------------------------")
	fmt.Printf("Workflow completed in %v\n", elapsed)
	fmt.Println("-----------------------------------")
	fmt.Println(result)

	// For comparison, use the traditional tool approach
	fmt.Println("-----------------------------------")
	fmt.Println("Now trying the same with traditional tool approach...")
	traditionalStart := time.Now()

	// Traditional tool-based approach
	prompt := "I need to calculate Fibonacci numbers in Python and suggest improvements"
	response, err := provider.SendMessageWithTools(ctx, prompt)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	traditionalElapsed := time.Since(traditionalStart)
	fmt.Println("-----------------------------------")
	fmt.Printf("Traditional approach completed in %v\n", traditionalElapsed)
	fmt.Println("-----------------------------------")
	fmt.Printf("Claude's response:\n%s\n", response)
}

