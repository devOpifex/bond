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

Represents a single reasoning step with dependencies:

```go
step := &reasoning.Step{
    ID:          "step-id",
    Name:        "Step Name",
    Description: "Step description",
    DependsOn:   []string{"dependency-step-id"},
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
chain.AddStep(step1)
chain.AddStep(step2)
result, err := chain.Execute(ctx, "initial input")
```

### Workflow

A directed graph of steps that can have complex dependencies. Steps are executed only after all their dependencies have completed:

```go
workflow := reasoning.NewWorkflow()
workflow.AddStep(step1)
workflow.AddStep(step2)
result, err := workflow.Execute(ctx, "initial input", "entry-point-step-id")
```

## Adapters

The package provides adapters to integrate with Bond agents and providers:

### Agent Adapter

```go
step := reasoning.AgentStep(
    "step-id",
    "Step Name",
    "Step description",
    "agent-capability",
    agentManager,
)
```

### Provider Adapter

```go
step := reasoning.ProviderStep(
    "step-id",
    "Step Name",
    "Step description",
    "Prompt template with %s placeholder",
    provider,
)
```

### Processor Adapter

```go
step := reasoning.ProcessorStep(
    "step-id",
    "Step Name",
    "Step description",
    func(ctx context.Context, input string, memory *reasoning.Memory) (string, map[string]interface{}, error) {
        // Custom processing logic
        return "output", map[string]interface{}{"key": "value"}, nil
    },
)
```

## Example Usage

See the `/examples/code/main.go` for a complete example of using multi-step reasoning for code generation and analysis.

Key benefits of the multi-step approach:

1. **State Management**: Pass information between steps
2. **Parallel Execution**: Independent steps can run concurrently
3. **Complex Dependencies**: Define step prerequisites for proper sequencing
4. **Result Reuse**: Access outputs from any previous step
5. **Composability**: Build complex workflows from simpler steps

## Best Practices

1. Use meaningful step IDs for better debugging
2. Keep each step focused on a single responsibility
3. Store intermediate results in memory for sharing between steps
4. Use metadata to capture additional information beyond the primary output
5. Handle errors appropriately at each step