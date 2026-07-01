package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OllamaProvider struct {
	BaseURL string 
	Model   string // e.g. "llama3.1", "codellama", "mistral"
	client  *http.Client
}

func NewOllamaProvider(baseURL, model string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434" 
	}
	if model == "" {
		model = "gemma4"
	}
	return &OllamaProvider{
		BaseURL: baseURL,
		Model:   model,
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (p *OllamaProvider) Name() string {
	return "ollama-" + p.Model
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	PromptEvalCount int `json:"prompt_eval_count"`
	EvalCount       int `json:"eval_count"`
}

func (p *OllamaProvider) Generate(prompt string) (*AnswerResult, error) {
	start := time.Now()

	reqBody := ollamaRequest{
		Model:  p.Model,
		Prompt: prompt,
		Stream: false, 
	}

	bodyBytes, _ := json.Marshal(reqBody)

	url := p.BaseURL + "/api/generate"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed (is `ollama serve` running?): %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama error %d: %s", resp.StatusCode, string(body))
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to parse ollama response: %w", err)
	}

	return &AnswerResult{
		Answer:       ollamaResp.Response,
		InputTokens:  ollamaResp.PromptEvalCount,
		OutputTokens: ollamaResp.EvalCount,
		LatencyMs:    time.Since(start).Milliseconds(),
		Provider:     p.Name(),
	}, nil
}