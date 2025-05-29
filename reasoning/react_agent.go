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

// Process implements the Agent interface and can be used as a step in a Chain
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
		// Get the last message to send to the provider
		lastMessage := ra.messages[len(ra.messages)-1]
		
		// Create a context that includes the full message history
		ctxWithHistory := context.WithValue(ctx, "message_history", ra.messages)
		
		// Get next thought from the model
		response, err := ra.provider.SendMessageWithTools(ctxWithHistory, lastMessage)
		if err != nil {
			return "", fmt.Errorf("provider error: %w", err)
		}

		// Add the assistant's response to the message history
		assistantMessage := models.Message{
			Role:    models.RoleAssistant,
			Content: response,
		}
		ra.messages = append(ra.messages, assistantMessage)

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
			
			// Store the assistant message for context
			_ = context.WithValue(ctx, "original_message", assistantMessage)
			
			// Add the tool result to the message history
			functionMessage := models.Message{
				Role:       models.RoleFunction,
				Content:    result,
				ToolResult: &models.ToolResult{ToolName: toolUse.Name, Result: result},
			}
			ra.messages = append(ra.messages, functionMessage)
		}
	}

	return finalResponse, nil
}

// AsStep returns the ReActAgent as a Chain Step for easy integration
func (ra *ReActAgent) AsStep(name string, description string) *Step {
	return &Step{
		Name:        name,
		Description: description,
		Execute:     ra.Process,
	}
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
		
		// Parse the tool use JSON manually to handle different input formats
		var rawToolUse map[string]interface{}
		if err := json.Unmarshal([]byte(toolJSON), &rawToolUse); err != nil {
			return nil, response, false
		}
		
		// Extract name and input
		name, ok := rawToolUse["name"].(string)
		if !ok {
			return nil, response, false
		}
		
		// Handle input which can be a string or an object
		var inputStr string
		if inputObj, ok := rawToolUse["input"].(map[string]interface{}); ok {
			// Input is an object, convert it to a JSON string
			inputJSON, err := json.Marshal(inputObj)
			if err != nil {
				return nil, response, false
			}
			inputStr = string(inputJSON)
		} else if inputStr, ok = rawToolUse["input"].(string); !ok {
			// Try marshaling whatever input is
			inputJSON, err := json.Marshal(rawToolUse["input"])
			if err != nil {
				return nil, response, false
			}
			inputStr = string(inputJSON)
		}
		
		toolUse := models.ToolUse{
			Name: name,
			Input: inputStr,
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
