package handlers

import (
	"codebuddy/embeddings"
	"codebuddy/ingestion"
	"codebuddy/models"
	"codebuddy/vectorstore"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type IngestRequest struct {
	RepoURL string `json:"repo_url" binding:"required"`
}

type IngestResponse struct {
	RepoURL    string        `json:"repo_url"`
	TotalFiles int           `json:"total_files"`
	TotalChunks int          `json:"total_chunks"`
	// Chunks     []models.Chunk `json:"chunks"`
	TotalTokens    int           `json:"total_tokens_used"`
	CollectionID   string        `json:"collection_id"`
	IngestionTimeS float64       `json:"ingestion_time_seconds"` 
	EmbeddingProvider string 	 `json:"embedding_provider"`
}

type ingestConfig interface{
	GetOpenApiKey() string
	GetChromaURL() string
	GetEmbeddingProvider() string
	GetOllamaURL() string
	GetOllamaEmbedModel() string
}

func IngestRepo(cfg ingestConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now() // track ingestion time

		var req IngestRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "repo_url is required"})
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

		if err != nil {
			c.JSON(http.StatusInternalServerError , gin.H{
				"error":"embedding provider setup failed: " + err.Error(),
			})
		}

		clonePath, err := ingestion.CloneRepo(req.RepoURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "clone failed: " + err.Error()})
			return
		}

		files, err := ingestion.WalkRepo(clonePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "walk failed: " + err.Error()})
			return
		}

		var allChunks []models.Chunk
		for _, file := range files {
			chunks, err := ingestion.ChunkFile(file, req.RepoURL)
			if err != nil {
				continue
			}
			allChunks = append(allChunks, chunks...)
		}

		fmt.Printf("Found %d files → %d chunks\n", len(files), len(allChunks))

		// 4. Embed
		embeddedChunks, totalTokens, err := embeddings.BatchEmbed(embedProvider, allChunks)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "embedding failed: " + err.Error()})
			return
		}

	
		chroma := vectorstore.NewChromaClient(cfg.GetChromaURL())

	
		repoName := ingestion.ExtractRepoName(req.RepoURL)

		collectionName := repoName + "-" + embedProvider.Name()

		collectionID, err := chroma.CreateCollection(collectionName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "chroma collection failed: " + err.Error()})
			return
		}

		if err := chroma.UpsertChunks(collectionID, embeddedChunks); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "chroma upsert failed: " + err.Error()})
			return
		}

		elapsed := time.Since(start).Seconds()

		c.JSON(http.StatusOK, IngestResponse{
			RepoURL:        req.RepoURL,
			TotalFiles:     len(files),
			TotalChunks:    len(allChunks),
			TotalTokens:    totalTokens,
			CollectionID:   collectionID,
			IngestionTimeS: elapsed,
			EmbeddingProvider: embedProvider.Name(),
		})
	}
}