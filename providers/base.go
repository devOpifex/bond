package providers

// Type represents the type of AI provider
type Type string

const (
	// Claude represents the Anthropic Claude provider
	Claude Type = "claude"
	// OpenAI represents the OpenAI provider
	OpenAI Type = "openai"
	// Add more providers here as they're implemented
	// Gemini Type = "gemini"
)