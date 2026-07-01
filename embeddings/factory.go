package embeddings

import "fmt"

type ProviderType string

const (
	ProviderOpenAI ProviderType = "openai"
	ProviderOllama ProviderType = "ollama"
)

type ProviderConfig struct {
	OpenAIKey   string
	OllamaURL   string
	OllamaModel string // embedding model, e.g. "nomic-embed-text"
}

// NewEmbeddingProvider is the factory — returns the correct EmbeddingProvider
// based on the type string from config. Single place that knows about all providers.
func NewEmbeddingProvider(providerType ProviderType, cfg ProviderConfig) (EmbeddingProvider, error) {
	switch providerType {
	case ProviderOpenAI:
		if cfg.OpenAIKey == "" {
			return nil, fmt.Errorf("openai embedding provider requires OPENAI_API_KEY")
		}
		return NewOpenAIEmbeddingProvider(cfg.OpenAIKey), nil

	case ProviderOllama:
		return NewOllamaEmbeddingProvider(cfg.OllamaURL, cfg.OllamaModel), nil

	default:
		return nil, fmt.Errorf("unknown embedding provider: %s (use 'openai' or 'ollama')", providerType)
	}
}