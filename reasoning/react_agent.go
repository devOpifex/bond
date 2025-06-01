// Package reasoning implements high-level reasoning patterns and workflows for AI agents.
// It provides structured approaches for complex AI behaviors like step-by-step reasoning,
// tool usage, and multi-step reasoning chains. This package builds upon the provider
// layer to enable more sophisticated agent behaviors.
package reasoning

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/devOpifex/bond/models"
)

// ReActAgent implements the Reasoning + Acting (ReAct) pattern for AI agents.
// This pattern involves alternating between reasoning about a problem and taking
// actions (using tools) to gather information or make progress toward a solution.
// The ReActAgent manages the cycle of reasoning, tool selection, and tool execution.
type ReActAgent struct {
	// provider is the LLM provider that handles communication with the AI model
	provider models.Provider
	
	// tools is a registry of available tools that the agent can use
	tools map[string]models.ToolExecutor
	
	// maxIterations limits the number of reasoning-action cycles to prevent infinite loops
	maxIterations int
	
	// systemPrompt contains instructions that guide the AI model's behavior
	systemPrompt string
	
	// messages stores the conversation history for context
	messages []models.Message
}

// NewReActAgent creates a new ReAct agent with the specified provider.
// It initializes the agent with default settings that can be customized
// through the SetMaxIterations and SetSystemPrompt methods.
func NewReActAgent(provider models.Provider) *ReActAgent {
	return &ReActAgent{
		provider:      provider,
		tools:         make(map[string]models.ToolExecutor),
		maxIterations: 10,
		messages:      []models.Message{},
		systemPrompt:  defaultReActPrompt,
	}
}

// RegisterTool adds a tool to the agent's available tools.
// The tool is stored in the agent's tool registry and will be
// available for the AI model to use during reasoning.
func (ra *ReActAgent) RegisterTool(tool models.ToolExecutor) {
	ra.tools[tool.GetName()] = tool
}

// SetMaxIterations configures the maximum number of reasoning-action cycles.
// This prevents the agent from getting stuck in infinite loops by limiting
// how many times it can go through the reasoning-action cycle.
func (ra *ReActAgent) SetMaxIterations(iterations int) {
	ra.maxIterations = iterations
}

// SetSystemPrompt overrides the default system prompt with a custom one.
// The system prompt provides instructions to the AI model about how to
// behave and how to structure its responses.
func (ra *ReActAgent) SetSystemPrompt(prompt string) {
	ra.systemPrompt = prompt
}

// Process implements the Agent interface and can be used as a step in a Chain.
// It executes the ReAct pattern, alternating between model reasoning and tool execution
// until a final response is reached or the maximum iterations limit is hit.
// This method handles the entire conversation flow, tool execution, and context management.
func (ra *ReActAgent) Process(ctx context.Context, input string) (string, error) {
	// Reset messages for this new conversation
	ra.messages = []models.Message{
		// Don't include system message in the messages array
		// System prompt is set separately on the provider
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

// AsStep returns the ReActAgent as a Chain Step for easy integration into workflows.
// This allows the ReAct agent to be used as a component in a larger reasoning pipeline.
// The name and description parameters are used to identify the step in the chain.
func (ra *ReActAgent) AsStep(name string, description string) *Step {
	return &Step{
		Name:        name,
		Description: description,
		Execute:     ra.Process,
	}
}

// parseResponse extracts tool use information from an LLM response.
// It returns:
// - A ToolUse pointer if the response contains a tool call (nil otherwise)
// - The text content from the response (thought text or final answer)
// - A boolean indicating if this is a final response (true) or if it requires tool use (false)
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
