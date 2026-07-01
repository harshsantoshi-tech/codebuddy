package main

import (
	"codebuddy/benchmark"
	"codebuddy/config"
	"codebuddy/embeddings"
	"codebuddy/llm"
	"codebuddy/vectorstore"
	"flag"
	"fmt"
	"log"
)

func main() {
	// CLI flags so you can run different combinations easily
	repoURL    := flag.String("repo",    "https://github.com/joho/godotenv", "repo to benchmark")
	embedType  := flag.String("embed",   "openai", "embedding provider: openai or ollama")
	llmType    := flag.String("llm",     "openai", "llm provider: openai or ollama")
	topK       := flag.Int("topk",       5,        "number of chunks to retrieve")
	runLLM     := flag.Bool("llm-eval",  false,    "also run LLM and measure answer quality")
	flag.Parse()

	fmt.Printf("🔍 CodeBuddy Benchmark\n")
	fmt.Printf("   Repo:      %s\n", *repoURL)
	fmt.Printf("   Embedding: %s\n", *embedType)
	fmt.Printf("   LLM:       %s\n", *llmType)
	fmt.Printf("   TopK:      %d\n", *topK)
	fmt.Printf("   Run LLM:   %v\n\n", *runLLM)

	// Load config (.env)
	cfg := config.Load()

	// Build embedding provider
	embedProvider, err := embeddings.NewEmbeddingProvider(
		embeddings.ProviderType(*embedType),
		embeddings.ProviderConfig{
			OpenAIKey:   cfg.GetOpenApiKey(),
			OllamaURL:   cfg.GetOllamaURL(),
			OllamaModel: cfg.GetOllamaEmbedModel(),
		},
	)
	if err != nil {
		log.Fatal("embedding provider error:", err)
	}

	// Build LLM provider
	llmProvider, err := llm.NewProvider(
		llm.ProviderType(*llmType),
		llm.ProviderConfig{
			OpenAIKey:   cfg.GetOpenApiKey(),
			OllamaURL:   cfg.GetOllamaURL(),
			OllamaModel: cfg.GetOllamaModel(),
		},
	)
	if err != nil {
		log.Fatal("llm provider error:", err)
	}

	// Run benchmark
	results := benchmark.Run(benchmark.BenchmarkConfig{
		RepoURL:       *repoURL,
		EmbedProvider: embedProvider,
		LLMProvider:   llmProvider,
		ChromaClient:  vectorstore.NewChromaClient(cfg.GetChromaURL()),
		TopK:          *topK,
		RunLLM:        *runLLM,
	})

	// Print report
	benchmark.PrintReport(results, benchmark.BenchmarkConfig{
		RepoURL:       *repoURL,
		EmbedProvider: embedProvider,
		LLMProvider:   llmProvider,
		TopK:          *topK,
		RunLLM:        *runLLM,
	})
}