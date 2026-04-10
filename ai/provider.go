// ABOUTME: AI provider interface and factory for vision-based metadata extraction.
// ABOUTME: Returns the correct provider implementation based on configuration.
package ai

import (
	"context"
	"fmt"
)

// AIProvider abstracts a vision model that can analyze an image and return text.
type AIProvider interface {
	AnalyzeImage(ctx context.Context, imageBase64 string, prompt string) (string, error)
}

// AIConfig holds the configuration for an AI provider.
type AIConfig struct {
	Provider string
	APIKey   string
	Model    string
	BaseURL  string
}

// NewAIProvider returns the appropriate provider implementation based on config.
// Supported providers: "anthropic", "openai-compatible".
func NewAIProvider(cfg AIConfig) (AIProvider, error) {
	switch cfg.Provider {
	case "anthropic":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("AI_API_KEY is required for anthropic provider")
		}
		if cfg.Model == "" {
			return nil, fmt.Errorf("AI_MODEL is required for anthropic provider")
		}
		return &AnthropicProvider{apiKey: cfg.APIKey, model: cfg.Model}, nil
	case "openai-compatible":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("AI_API_KEY is required for openai-compatible provider")
		}
		if cfg.Model == "" {
			return nil, fmt.Errorf("AI_MODEL is required for openai-compatible provider")
		}
		if cfg.BaseURL == "" {
			return nil, fmt.Errorf("AI_BASE_URL is required for openai-compatible provider")
		}
		return &OpenAICompatProvider{apiKey: cfg.APIKey, model: cfg.Model, baseURL: cfg.BaseURL}, nil
	default:
		return nil, fmt.Errorf("unknown AI provider: %q (supported: anthropic, openai-compatible)", cfg.Provider)
	}
}
