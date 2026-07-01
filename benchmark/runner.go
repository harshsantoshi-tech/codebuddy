package benchmark

import (
	"codebuddy/embeddings"
	"codebuddy/llm"
	"codebuddy/retrieval"
	"codebuddy/vectorstore"
	"fmt"
	"strings"
	"time"
)

// Result holds metrics for one test case
type Result struct {
	TestCase
	// Retrieval metrics
	TopFiles        []string // files that appeared in top-5
	PrecisionHit    bool     // did any ExpectedFile appear in top-5?
	RetrievalTimeMs int64
	AvgSimilarity   float32

	// LLM metrics
	Answer    string
	LLMTimeMs int64
	Tokens    int
	Provider  string

	// Overall
	TotalTimeMs int64
	Error       string
}

// BenchmarkConfig holds everything the runner needs
type BenchmarkConfig struct {
	RepoURL           string
	EmbedProvider     embeddings.EmbeddingProvider
	LLMProvider       llm.LLMProvider
	ChromaClient      *vectorstore.ChromaClient
	TopK              int
	RunLLM            bool // set false to benchmark retrieval only (faster)
}

// Run executes all test cases and returns results
func Run(cfg BenchmarkConfig) []Result {
	results := make([]Result, len(TestSuite))

	for i, tc := range TestSuite {
		fmt.Printf("\n[%d/%d] %s: \"%s\"\n", i+1, len(TestSuite), tc.Category, tc.Query)

		result := Result{TestCase: tc}
		start := time.Now()

		// --- Retrieval ---
		retrieved, err := retrieval.Retrieve(
			cfg.EmbedProvider,
			cfg.ChromaClient,
			cfg.RepoURL,
			tc.Query,
			cfg.TopK,
		)
		if err != nil {
			result.Error = err.Error()
			results[i] = result
			fmt.Println("  ❌ retrieval error:", err)
			continue
		}

		result.RetrievalTimeMs = retrieved.RetrievalTimeMs

		// Collect top file paths and avg similarity
		var totalSim float32
		for _, chunk := range retrieved.Chunks {
			result.TopFiles = append(result.TopFiles, chunk.FilePath)
			totalSim += chunk.Similarity
		}
		if len(retrieved.Chunks) > 0 {
			result.AvgSimilarity = totalSim / float32(len(retrieved.Chunks))
		}

		// Check precision — did any expected file appear in top-K?
		result.PrecisionHit = checkPrecision(result.TopFiles, tc.ExpectedFiles)
		hitStr := "✅"
		if !result.PrecisionHit {
			hitStr = "❌"
		}
		fmt.Printf("  %s retrieval %dms | sim=%.2f | files=%v\n",
			hitStr, result.RetrievalTimeMs, result.AvgSimilarity, uniqueFiles(result.TopFiles))

		// --- LLM (optional) ---
		if cfg.RunLLM {
			prompt := llm.BuildPrompt(tc.Query, retrieved.Chunks)
			answer, err := cfg.LLMProvider.Generate(prompt)
			if err != nil {
				result.Error = "llm: " + err.Error()
				fmt.Println("  ❌ LLM error:", err)
			} else {
				result.Answer = answer.Answer
				result.LLMTimeMs = answer.LatencyMs
				result.Tokens = answer.InputTokens + answer.OutputTokens
				result.Provider = answer.Provider
				fmt.Printf("  💬 LLM %dms | tokens=%d\n", result.LLMTimeMs, result.Tokens)
			}
		}

		result.TotalTimeMs = time.Since(start).Milliseconds()
		results[i] = result
	}

	return results
}

// checkPrecision returns true if any retrieved file matches any expected file
func checkPrecision(retrieved []string, expected []string) bool {
	for _, r := range retrieved {
		for _, e := range expected {
			// Use Contains so "handlers/godotenv.go" matches "godotenv.go"
			if strings.Contains(r, e) {
				return true
			}
		}
	}
	return false
}

func uniqueFiles(files []string) []string {
	seen := map[string]bool{}
	var unique []string
	for _, f := range files {
		// shorten path for display: "some/long/path/file.go" -> "file.go"
		parts := strings.Split(f, "/")
		short := parts[len(parts)-1]
		if !seen[short] {
			seen[short] = true
			unique = append(unique, short)
		}
	}
	return unique
}