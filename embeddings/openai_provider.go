package embeddings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	openAIEmbedURL   = "https://api.openai.com/v1/embeddings"
	openAIEmbedModel = "text-embedding-3-small"
	openAIDimensions = 1536
)

// OpenAIEmbeddingProvider implements EmbeddingProvider using OpenAI's API
type OpenAIEmbeddingProvider struct {
	APIKey string
	client *http.Client
}

func NewOpenAIEmbeddingProvider(apiKey string) *OpenAIEmbeddingProvider {
	return &OpenAIEmbeddingProvider{
		APIKey: apiKey,
		client: &http.Client{},
	}
}

func (p *OpenAIEmbeddingProvider) Name() string       { return "openai-" + openAIEmbedModel }
func (p *OpenAIEmbeddingProvider) Dimensions() int    { return openAIDimensions }

type openAIEmbedRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type openAIEmbedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

func (p *OpenAIEmbeddingProvider) EmbedTexts(texts []string) ([][]float32, int, error) {
	reqBody := openAIEmbedRequest{Input: texts, Model: openAIEmbedModel}
	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", openAIEmbedURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("openai embed request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("openai embed error %d: %s", resp.StatusCode, string(body))
	}

	var embedResp openAIEmbedResponse
	if err := json.Unmarshal(body, &embedResp); err != nil {
		return nil, 0, err
	}

	// Results come back index-ordered but let's be safe
	vectors := make([][]float32, len(texts))
	for _, item := range embedResp.Data {
		vectors[item.Index] = item.Embedding
	}

	return vectors, embedResp.Usage.TotalTokens, nil
}