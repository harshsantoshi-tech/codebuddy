package embeddings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	openAiEmbedURL = "https://api.openai.com/v1/embeddings"
	EmbedModel = "text-embedding-3-small"
)

type EmbedRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type EmbedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`    
	} `json:"data"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

func EmbedTexts(apiKey string , texts []string)([][]float32 , int , error){

	reqBody := EmbedRequest{
		Input: texts,
		Model: EmbedModel,
	}
	reqJ , _ := json.Marshal(reqBody)

	req , err := http.NewRequest("POST" , openAiEmbedURL , bytes.NewBuffer(reqJ))
	if err != nil {
		return nil , 0 , err
	}

	req.Header.Set("Content-Type" , "application/json")
	req.Header.Set("Authorization" , "Bearer "+ apiKey)


	client := &http.Client{}

	res , err := client.Do(req)
	if err != nil {
		return nil , 0 , fmt.Errorf("openai request failed : %w" , err)
	}
	defer res.Body.Close()


	body , _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK{
		return nil , 0 , fmt.Errorf("openai error %d : %s " , res.StatusCode , string(body))

	}

	var embedRes EmbedResponse
	if err := json.Unmarshal(body , &embedRes); err != nil {
		return nil , 0 , err
	}

	vectors := make([][]float32 , len(texts))

	for _ , item := range embedRes.Data{
		vectors[item.Index] = item.Embedding
	}
	return vectors , embedRes.Usage.TotalTokens , nil
}
