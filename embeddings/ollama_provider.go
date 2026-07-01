package embeddings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Best open-source embedding models available in Ollama:
// - nomic-embed-text  → 768 dims, fast, great quality
// - mxbai-embed-large → 1024 dims, slower, higher quality
// Pull with: ollama pull nomic-embed-text

const (
	ollamaEmbedDefaultModel = "nomic-embed-text"
	ollamaEmbedDimensions   = 768 // nomic-embed-text output size
)

// OllamaEmbeddingProvider implements EmbeddingProvider using local Ollama
type OllamaEmbeddingProvider struct {
	BaseURL string
	Model   string
	dims    int
	client  *http.Client
}

func NewOllamaEmbeddingProvider(baseURL, model string) *OllamaEmbeddingProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = ollamaEmbedDefaultModel
	}

	// Map model → known dimensions
	// If you add a new model, add it here
	dims := ollamaEmbedDimensions
	if model == "mxbai-embed-large" {
		dims = 1024
	}

	return &OllamaEmbeddingProvider{
		BaseURL: baseURL,
		Model:   model,
		dims:    dims,
		client:  &http.Client{},
	}
}

func (p *OllamaEmbeddingProvider) Name() string    { return "ollama-" + p.Model }
func (p *OllamaEmbeddingProvider) Dimensions() int { return p.dims }

type ollamaEmbedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"` // Ollama embeds one string at a time
}

type ollamaEmbedResponse struct {
	Embedding []float32 `json:"embedding"`
}

// EmbedTexts calls Ollama once per text (no batch API in Ollama).
// Tradeoff vs OpenAI: more HTTP calls, but zero cost and fully local.
func (p *OllamaEmbeddingProvider) EmbedTexts(texts []string) ([][]float32, int, error) {
	vectors := make([][]float32, len(texts))

	for i, text := range texts {
		reqBody := ollamaEmbedRequest{
			Model:  p.Model,
			Prompt: text,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		resp, err := p.client.Post(
			p.BaseURL+"/api/embeddings",
			"application/json",
			bytes.NewBuffer(bodyBytes),
		)
		if err != nil {
			return nil, 0, fmt.Errorf("ollama embed request %d failed (is ollama running?): %w", i, err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			return nil, 0, fmt.Errorf("ollama embed error %d: %s", resp.StatusCode, string(body))
		}

		var embedResp ollamaEmbedResponse
		if err := json.Unmarshal(body, &embedResp); err != nil {
			return nil, 0, fmt.Errorf("failed to parse ollama embed response: %w", err)
		}

		vectors[i] = embedResp.Embedding
	}

	// Ollama doesn't report token counts — return 0
	// This means token cost tracking only works for OpenAI
	return vectors, 0, nil
}