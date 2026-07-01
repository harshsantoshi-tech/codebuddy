package llm

import "fmt"

type ProviderType string

const (
	ProviderOpenAI ProviderType = "openai"
	ProviderOllama ProviderType = "ollama"
)

type ProviderConfig struct {
	OpenAIKey   string
	OllamaURL   string
	OllamaModel string
}

func NewProvider(providerType ProviderType, cfg ProviderConfig) (LLMProvider, error) {
	switch providerType {
	case ProviderOpenAI:
		if cfg.OpenAIKey == "" {
			return nil, fmt.Errorf("openai provider requires OPENAI_API_KEY")
		}
		return NewOpenAIProvider(cfg.OpenAIKey), nil

	case ProviderOllama:
		return NewOllamaProvider(cfg.OllamaURL, cfg.OllamaModel), nil

	default:
		return nil, fmt.Errorf("unknown LLM provider: %s (use 'openai' or 'ollama')", providerType)
	}
}