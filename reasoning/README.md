# Bond Reasoning

The `reasoning` package extends Bond with multi-step reasoning capabilities, allowing agents to break down complex tasks into a series of steps with state management.

## Key Components

### Memory

Provides a thread-safe key-value store for maintaining state between reasoning steps:

```go
memory := reasoning.NewMemory()
memory.Set("key", "value")
value, exists := memory.GetString("key")
```

### Step

Represents a single reasoning step:

```go
step := &reasoning.Step{
    Name:        "Step Name",
    Description: "Step description",
    Execute: func(ctx context.Context, input string, memory *reasoning.Memory) (reasoning.StepResult, error) {
        // Process input and return results
        return reasoning.StepResult{
            Output: "Step output",
            Metadata: map[string]interface{}{
                "key": "value",
            },
        }, nil
    },
}
```

### Chain

A linear sequence of steps executed in order, where each step's output becomes the input to the next step:

```go
chain := reasoning.NewChain()
chain.AddStep(step1).Then(step2)  // Method chaining supported
result, err := chain.Execute(ctx, "initial input")
```

### Workflow

A simplified sequence of steps that is equivalent to a Chain:

```go
workflow := reasoning.NewWorkflow()
workflow.AddStep(step1).Then(step2)  // Method chaining supported
result, err := workflow.Execute(ctx, "initial input")
```

### ReActAgent

A Reasoning + Acting agent that follows a loop of reasoning about what to do next, taking actions using tools, and incorporating observations to achieve a goal:

```go
provider := claude.NewClaudeProvider("<API_KEY>")
reactAgent := reasoning.NewReActAgent(provider)
reactAgent.RegisterTool(someTool)
result, err := reactAgent.Process(ctx, "Solve this problem...")
```

## Adapters

The package provides adapters to integrate with Bond agents and providers:

### Agent Adapter

```go
step := reasoning.AgentStep(
    "Step Name",
    "Step description",
    "agent-capability",
    agentManager,
)
```

### Provider Adapter

```go
step := reasoning.ProviderStep(
    "Step Name",
    "Step description",
    "Prompt template with %s placeholder",
    provider,
)
```

### Processor Adapter

```go
step := reasoning.ProcessorStep(
    "Step Name",
    "Step description",
    func(ctx context.Context, input string, memory *reasoning.Memory) (string, map[string]interface{}, error) {
        // Custom processing logic
        return "output", map[string]interface{}{"key": "value"}, nil
    },
)
```

## Example Usage

See the `/examples/react/main.go` for a complete example of using the ReAct agent with tools.

### Accessing Step Outputs

You can access step outputs using automatically generated step IDs:

```go
// Access the output of the first step (index 0)
firstStepOutput, _ := memory.GetString("step.step_0.output")

// Access the output of the second step (index 1)
secondStepOutput, _ := memory.GetString("step.step_1.output")
```

Key benefits of the multi-step approach:

1. **State Management**: Pass information between steps
2. **Sequential Execution**: Clear, readable flow of operations
3. **Result Reuse**: Access outputs from any previous step
4. **Composability**: Build complex workflows from simpler steps
5. **Fluent API**: Method chaining for more readable code

## Best Practices

1. Keep each step focused on a single responsibility
2. Store intermediate results in memory for sharing between steps
3. Use metadata to capture additional information beyond the primary output
4. Handle errors appropriately at each step
5. Use the fluent API with method chaining for more readable code