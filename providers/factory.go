package providers

import (
	"fmt"

	"github.com/devOpifex/bond/models"
	"github.com/devOpifex/bond/providers/claude"
	"github.com/devOpifex/bond/providers/openai"
)

// NewProvider creates a new provider of the specified type
func NewProvider(providerType Type, apiKey string) (models.Provider, error) {
	switch providerType {
	case Claude:
		return claude.NewClient(apiKey), nil
	case OpenAI:
		return openai.NewClient(apiKey), nil
	// Add cases for other providers as they're implemented
	// case Gemini:
	//     return gemini.NewClient(apiKey), nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}