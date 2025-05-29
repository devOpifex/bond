package reasoning

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/devOpifex/bond/models"
)

// ReActAgent implements a Reasoning + Acting agent
type ReActAgent struct {
	provider      models.Provider
	tools         map[string]models.ToolExecutor
	maxIterations int
	systemPrompt  string
	messages      []models.Message
}

// NewReActAgent creates a new ReAct agent
func NewReActAgent(provider models.Provider) *ReActAgent {
	return &ReActAgent{
		provider:      provider,
		tools:         make(map[string]models.ToolExecutor),
		maxIterations: 10,
		messages:      []models.Message{},
		systemPrompt:  defaultReActPrompt,
	}
}

// RegisterTool adds a tool to the agent
func (ra *ReActAgent) RegisterTool(tool models.ToolExecutor) {
	ra.tools[tool.GetName()] = tool
}

// SetMaxIterations configures the maximum number of reasoning-action cycles
func (ra *ReActAgent) SetMaxIterations(iterations int) {
	ra.maxIterations = iterations
}

// SetSystemPrompt overrides the default system prompt
func (ra *ReActAgent) SetSystemPrompt(prompt string) {
	ra.systemPrompt = prompt
}

// Process implements the Agent interface
func (ra *ReActAgent) Process(ctx context.Context, input string) (string, error) {
	// Reset messages for this new conversation
	ra.messages = []models.Message{
		{
			Role:    models.RoleSystem,
			Content: ra.systemPrompt,
		},
		{
			Role:    models.RoleUser,
			Content: input,
		},
	}

	// Register all tools with the provider
	for _, tool := range ra.tools {
		ra.provider.RegisterTool(tool)
	}

	// Set the system prompt for the provider
	ra.provider.SetSystemPrompt(ra.systemPrompt)

	var finalResponse string
	
	// Main ReAct loop
	for i := 0; i < ra.maxIterations; i++ {
		// Get next thought from the model
		response, err := ra.provider.SendMessageWithTools(ctx, ra.messages[len(ra.messages)-1])
		if err != nil {
			return "", fmt.Errorf("provider error: %w", err)
		}

		// Add the assistant's response to the message history
		ra.messages = append(ra.messages, models.Message{
			Role:    models.RoleAssistant,
			Content: response,
		})

		// Parse response to extract tool calls
		toolUse, actionText, isFinalResponse := parseResponse(response)
		
		// If this is a final response with no tool use, we're done
		if isFinalResponse {
			finalResponse = actionText
			break
		}

		// If there's a tool to use, execute it
		if toolUse != nil {
			// Find the tool
			tool, exists := ra.tools[toolUse.Name]
			if !exists {
				toolResult := fmt.Sprintf("Error: Tool '%s' not found", toolUse.Name)
				ra.messages = append(ra.messages, models.Message{
					Role:       models.RoleFunction,
					Content:    toolResult,
					ToolResult: &models.ToolResult{ToolName: toolUse.Name, Result: toolResult},
				})
				continue
			}

			// Parse the input
			var inputJSON json.RawMessage
			if err := json.Unmarshal([]byte(toolUse.Input), &inputJSON); err != nil {
				toolResult := fmt.Sprintf("Error: Invalid tool input JSON: %v", err)
				ra.messages = append(ra.messages, models.Message{
					Role:       models.RoleFunction,
					Content:    toolResult,
					ToolResult: &models.ToolResult{ToolName: toolUse.Name, Result: toolResult},
				})
				continue
			}

			// Execute the tool
			result, err := tool.Execute(inputJSON)
			if err != nil {
				toolResult := fmt.Sprintf("Error executing tool: %v", err)
				ra.messages = append(ra.messages, models.Message{
					Role:       models.RoleFunction,
					Content:    toolResult,
					ToolResult: &models.ToolResult{ToolName: toolUse.Name, Result: toolResult},
				})
				continue
			}

			// Add the tool result to the message history
			ra.messages = append(ra.messages, models.Message{
				Role:       models.RoleFunction,
				Content:    result,
				ToolResult: &models.ToolResult{ToolName: toolUse.Name, Result: result},
			})
		}
	}

	return finalResponse, nil
}

// Helper function to parse the LLM response to extract tool use
func parseResponse(response string) (*models.ToolUse, string, bool) {
	// Simple parsing - in a real implementation, this would be more robust
	if strings.Contains(response, "```json") && strings.Contains(response, "\"name\":") {
		// Extract the JSON between the markers
		startMarker := "```json"
		endMarker := "```"
		
		startIdx := strings.Index(response, startMarker) + len(startMarker)
		endIdx := strings.Index(response[startIdx:], endMarker)
		if endIdx == -1 {
			return nil, response, false
		}
		
		toolJSON := strings.TrimSpace(response[startIdx : startIdx+endIdx])
		
		var toolUse models.ToolUse
		if err := json.Unmarshal([]byte(toolJSON), &toolUse); err != nil {
			return nil, response, false
		}
		
		// Extract the thought text before the tool use
		thoughtText := strings.TrimSpace(response[:strings.Index(response, "```json")])
		return &toolUse, thoughtText, false
	}
	
	// If no tool use is detected, this is a final response
	return nil, response, true
}

// Default system prompt for ReAct agents
const defaultReActPrompt = `You are a ReAct agent that can use tools to solve problems.
For each step:
1. Think about what to do next
2. If you need to use a tool, format your response like this:

<thought>
Your detailed reasoning about what to do next...
</thought>

` + "```json" + `
{
  "name": "tool_name",
  "input": "tool input in JSON format"
}
` + "```" + `

3. If you have the final answer, just respond directly without using tools.

Examples of tool usage:
<thought>
I need to search for information about Python.
</thought>

` + "```json" + `
{
  "name": "search",
  "input": "Python programming language"
}
` + "```" + `

Always think carefully before deciding which tool to use.`
