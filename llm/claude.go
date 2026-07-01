package llm

import (
	"strings"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

)

const (
	claudeAPIURL = "https://api.antropic.com/v1/messages"
	claudeModel = "claude-sonnet-4-6"
)

type ClaudeClient struct{
	APIKey string
	client *http.Client
}

func NewClaudeClient(apiKey string)*ClaudeClient{
	return &ClaudeClient{
		APIKey: apiKey,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type claudeRequest struct{
	Model string `json:"model"`
	MaxTokens int `json:"max_tokens"`
	Messages []message `json:"messages"`
}

type message struct{
	Role string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct{
	Content []struct{
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func ( c *ClaudeClient)Generate(prompt string)(*AnswerResult , error){

	start := time.Now()

	reqBody := claudeRequest{
		Model : claudeModel,
		MaxTokens: maxTokens,
		Messages: []message{
			{Role :"User" , Content: prompt},
		},
	}

	bodyBytes , err := json.Marshal(reqBody)
	if err != nil {
		return nil , fmt.Errorf("failed to marshal request : %w" , err)
	}

	req , err := http.NewRequest("POST" , claudeAPIURL , bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil , err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key" , c.APIKey)

	req.Header.Set("anthropic-version","2023-06-01")

	resp , err := c.client.Do(req)
	if err != nil {
		return nil , fmt.Errorf("claude request failed: %w" , err)
	}

	defer resp.Body.Close()

	body , _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK{
		return nil , fmt.Errorf("claude error %d: %s" , resp.StatusCode , string(body))
	}

	var claudeRes claudeResponse

	if err := json.Unmarshal(body , &claudeRes) ; err != nil {
		return nil , fmt.Errorf("failed to parse claude response : %w",err)
	}

	if claudeRes.Error != nil {
		return nil , fmt.Errorf("claude API error : %s" , claudeRes.Error.Message)
	}


	var answerText strings.Builder

	for _ , block := range claudeRes.Content{
		if block.Text == "text"{
			answerText .WriteString(block.Text)
		}
	}

	latency := time.Since(start).Milliseconds()

	return &AnswerResult{
		Answer: answerText.String(),
		InputTokens: claudeRes.Usage.InputTokens,
		OutputTokens: claudeRes.Usage.OutputTokens,
		LatencyMs: latency,
	},nil
}