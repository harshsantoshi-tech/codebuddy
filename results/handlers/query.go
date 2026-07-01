package handlers

import (
	"codebuddy/embeddings"
	"codebuddy/llm"
	"codebuddy/retrieval"
	"codebuddy/vectorstore"
	"net/http"

	"github.com/gin-gonic/gin"
)

type QueryRequest struct {
	RepoURL  string `json:"repo_url" binding:"required"`
	Query    string `json:"query"    binding:"required"`
	TopK     int    `json:"top_k"`
	Provider string `json:"provider"`
}

type QueryResponse struct {
	Query           string         `json:"query"`
	RepoURL         string         `json:"repo_url"`
	Answer          string         `json:"answer"`
	Provider        string         `json:"provider"`
	Citations       []llm.Citation `json:"citations"`
	RetrievalTimeMs int64          `json:"retrieval_time_ms"`
	LLMTimeMs       int64          `json:"llm_time_ms"`
	TotalTimeMs     int64          `json:"total_time_ms"`
	TokensUsed      int            `json:"tokens_used"`
	EmbeddingProvider string 	   `json:"embedding_provider"`
}

type queryConfig interface {
	GetOpenApiKey() string
	GetChromaURL() string
	GetLLMProvider() string
	GetOllamaURL() string
	GetOllamaModel() string
	GetEmbeddingProvider() string
	GetOllamaEmbedModel() string
}

func QueryRepo(cfg queryConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req QueryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "repo_url and query are required"})
			return
		}

		embedProvider , err := embeddings.NewEmbeddingProvider(
			embeddings.ProviderType(cfg.GetEmbeddingProvider()),
			embeddings.ProviderConfig{
				OpenAIKey: cfg.GetOpenApiKey(),
				OllamaURL: cfg.GetOllamaURL(),
				OllamaModel: cfg.GetOllamaEmbedModel(),
			},
		)
		if err != nil{
			c.JSON(http.StatusInternalServerError , gin.H{
				"error" : "embedding provider setup failed : " + err.Error(),
			})
		}

		chroma := vectorstore.NewChromaClient(cfg.GetChromaURL())
		result, err := retrieval.Retrieve(embedProvider, chroma, req.RepoURL, req.Query, req.TopK)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		prompt := llm.BuildPrompt(req.Query, result.Chunks)

		providerType := cfg.GetLLMProvider()
		if req.Provider != "" {
			providerType = req.Provider
		}

		provider, err := llm.NewProvider(
			llm.ProviderType(providerType),
			llm.ProviderConfig{
				OpenAIKey:   cfg.GetOpenApiKey(),
				OllamaURL:   cfg.GetOllamaURL(),
				OllamaModel: cfg.GetOllamaModel(),
			},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "provider setup failed: " + err.Error()})
			return
		}

		answer, err := provider.Generate(prompt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "LLM call failed: " + err.Error()})
			return
		}

		citations := llm.BuildCitationSummary(result.Chunks)

		c.JSON(http.StatusOK, QueryResponse{
			Query:           req.Query,
			RepoURL:         req.RepoURL,
			Answer:          answer.Answer,
			Provider:        answer.Provider,
			EmbeddingProvider: result.EmbeddingProvider,
			Citations:       citations,
			RetrievalTimeMs: result.RetrievalTimeMs,
			LLMTimeMs:       answer.LatencyMs,
			TotalTimeMs:     result.RetrievalTimeMs + answer.LatencyMs,
			TokensUsed:      answer.InputTokens + answer.OutputTokens,
		})
	}
}