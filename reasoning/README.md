# Bond Reasoning

The `reasoning` package extends Bond with multi-step reasoning capabilities, allowing agents to break down complex tasks into a series of steps.

## Key Components

### Step

Represents a single reasoning step:

```go
step := &reasoning.Step{
    Name:        "Step Name",
    Description: "Step description",
    Execute: func(ctx context.Context, input string) (string, error) {
        // Process input and return results
        return "Step output", nil
    },
}
```

### Chain

A linear sequence of steps executed in order, where each step's output becomes the input to the next step:

```go
chain := reasoning.NewChain()
chain.Add(step1).Then(step2)  // Method chaining supported
result, err := chain.Execute(ctx, "initial input")
```

### ReActAgent

A Reasoning + Acting agent that follows a loop of reasoning about what to do next, taking actions using tools, and incorporating observations to achieve a goal:

```go
provider := claude.NewClaudeProvider("<API_KEY>")
reactAgent := reasoning.NewReActAgent(provider)
reactAgent.RegisterTool(someTool)
result, err := reactAgent.Process(ctx, "Solve this problem...")
```

## Step Creators

The package provides factory functions to easily create steps:

### WithAgent

```go
step := reasoning.WithAgent(
    "Step Name",
    "Step description",
    "agent-capability",
    agentManager,
)
```

### WithProvider

```go
step := reasoning.WithProvider(
    "Step Name",
    "Step description",
    "Prompt template with %s placeholder",
    provider,
)
```

### WithProcessor

```go
step := reasoning.WithProcessor(
    "Step Name",
    "Step description",
    func(ctx context.Context, input string) (string, error) {
        // Custom processing logic
        return "processed output", nil
    },
)
```

## Example Usage

See the `/examples/react_in_chain/main.go` for a complete example of using the simplified Chain API.

```go
// Create a chain
chain := reasoning.NewChain()

// Add steps using the simplified API
chain.Add(reasoning.WithProcessor(
    "Extract entities", 
    "Extracts entities from text",
    func(ctx context.Context, input string) (string, error) {
        // Process input
        return "Extracted entities: X, Y, Z", nil
    },
)).
Then(reasoning.WithProvider(
    "Analyze sentiment",
    "Analyzes the sentiment of the input",
    "Analyze the sentiment of this text: %s",
    provider,
))

// Execute the chain
result, err := chain.Execute(context.Background(), "Your input text here")
```

Key benefits of the simplified approach:

1. **Simplicity**: Focused on the core task of chaining operations
2. **Clarity**: Each step has a clear input and output
3. **Composability**: Build complex workflows from simpler steps
4. **Fluent API**: Method chaining for more readable code

## Best Practices

1. Keep each step focused on a single responsibility
2. Create reusable steps for common operations
3. Handle errors appropriately at each step
4. Use the fluent API with method chaining for more readable code