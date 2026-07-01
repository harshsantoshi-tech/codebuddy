package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	openAIChatURL   = "https://api.openai.com/v1/chat/completions"
	openAIChatModel = "gpt-4o"
	maxTokens       = 1024
)

// OpenAIProvider implements LLMProvider using OpenAI's chat completions API
type OpenAIProvider struct {
	APIKey string
	client *http.Client
}

func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{
		APIKey: apiKey,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *OpenAIProvider) Name() string {
	return "openai-" + openAIChatModel
}

type openAIChatRequest struct {
	Model     string        `json:"model"`
	MaxTokens int           `json:"max_tokens"`
	Messages  []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (p *OpenAIProvider) Generate(prompt string) (*AnswerResult, error) {
	start := time.Now()

	reqBody := openAIChatRequest{
		Model:     openAIChatModel,
		MaxTokens: maxTokens,
		Messages: []chatMessage{
			{Role: "system", Content: "You are an expert code assistant. Answer using only the provided context. Always cite sources."},
			{Role: "user", Content: prompt},
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", openAIChatURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai error %d: %s", resp.StatusCode, string(body))
	}

	var chatResp openAIChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	if chatResp.Error != nil {
		return nil, fmt.Errorf("openai API error: %s", chatResp.Error.Message)
	}
	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from openai")
	}

	return &AnswerResult{
		Answer:       chatResp.Choices[0].Message.Content,
		InputTokens:  chatResp.Usage.PromptTokens,
		OutputTokens: chatResp.Usage.CompletionTokens,
		LatencyMs:    time.Since(start).Milliseconds(),
		Provider:     p.Name(),
	}, nil
}