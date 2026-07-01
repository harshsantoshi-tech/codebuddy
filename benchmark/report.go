package benchmark

import (
	"fmt"
	"strings"
)

// PrintReport prints a full benchmark report to stdout
// Copy-paste this into your README
func PrintReport(results []Result, cfg BenchmarkConfig) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("           CODEBUDDY BENCHMARK REPORT")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Repo:               %s\n", cfg.RepoURL)
	fmt.Printf("Embedding Provider: %s\n", cfg.EmbedProvider.Name())
	if cfg.RunLLM {
		fmt.Printf("LLM Provider:       %s\n", cfg.LLMProvider.Name())
	}
	fmt.Printf("Top-K:              %d\n", cfg.TopK)
	fmt.Printf("Total Questions:    %d\n", len(results))
	fmt.Println(strings.Repeat("-", 60))

	// --- Retrieval metrics ---
	hits := 0
	var totalRetMs int64
	var totalSim float32
	validCount := 0

	categoryHits := map[string]int{}
	categoryTotal := map[string]int{}

	for _, r := range results {
		if r.Error != "" {
			continue
		}
		validCount++
		if r.PrecisionHit {
			hits++
			categoryHits[r.Category]++
		}
		categoryTotal[r.Category]++
		totalRetMs += r.RetrievalTimeMs
		totalSim += r.AvgSimilarity
	}

	var precision float64
	var avgRetMs int64
	var avgSim float32

	if validCount > 0 {
		precision = float64(hits) / float64(validCount) * 100
		avgRetMs = totalRetMs / int64(validCount)
		avgSim = totalSim / float32(validCount)
	} else {
		fmt.Println("\n⚠️  All queries failed — make sure the repo is indexed with the correct embedding provider.")
		fmt.Printf("    Expected collection: <reponame>-%s\n", cfg.EmbedProvider.Name())
		fmt.Println("    Run: POST /api/v1/ingest with matching EMBEDDING_PROVIDER in .env")
		return
	}

	fmt.Printf("\n📊 RETRIEVAL METRICS\n")
	fmt.Printf("  Overall Precision@%d:  %.1f%% (%d/%d)\n", cfg.TopK, precision, hits, validCount)
	fmt.Printf("  Avg Retrieval Latency: %dms\n", avgRetMs)
	fmt.Printf("  Avg Similarity Score:  %.3f\n", avgSim)

	fmt.Printf("\n  Precision by category:\n")
	for _, cat := range []string{"feature", "bug", "architecture", "fuzzy"} {
		total := categoryTotal[cat]
		if total == 0 {
			continue
		}
		catHits := categoryHits[cat]
		bar := strings.Repeat("█", catHits) + strings.Repeat("░", total-catHits)
		fmt.Printf("    %-14s %s  %d/%d (%.0f%%)\n",
			cat+":", bar, catHits, total, float64(catHits)/float64(total)*100)
	}

	// --- LLM metrics ---
	if cfg.RunLLM {
		var totalLLMMs int64
		var totalTokens int
		llmCount := 0
		for _, r := range results {
			if r.Error == "" && r.LLMTimeMs > 0 {
				totalLLMMs += r.LLMTimeMs
				totalTokens += r.Tokens
				llmCount++
			}
		}
		if llmCount > 0 {
			fmt.Printf("\n💬 LLM METRICS\n")
			fmt.Printf("  Avg LLM Latency:    %dms\n", totalLLMMs/int64(llmCount))
			fmt.Printf("  Avg Tokens/Query:   %d\n", totalTokens/llmCount)
			fmt.Printf("  Avg Total Latency:  %dms\n", (totalRetMs+totalLLMMs)/int64(llmCount))
			fmt.Printf("  Est. Cost/Query:    $%.5f (OpenAI gpt-4o pricing)\n",
				float64(totalTokens/llmCount)*0.000005)
		}
	}

	// --- Per question breakdown ---
	fmt.Printf("\n📋 PER-QUESTION BREAKDOWN\n")
	fmt.Printf("%-3s %-14s %-42s %s\n", "#", "Category", "Query", "Result")
	fmt.Println(strings.Repeat("-", 80))

	for i, r := range results {
		status := "✅"
		if !r.PrecisionHit {
			status = "❌"
		}
		if r.Error != "" {
			status = "💥"
		}

		query := r.Query
		if len(query) > 40 {
			query = query[:37] + "..."
		}
		detail := fmt.Sprintf("%dms sim=%.2f", r.RetrievalTimeMs, r.AvgSimilarity)
		if r.Error != "" {
			detail = "ERROR: " + r.Error
		}
		fmt.Printf("%-3d %-14s %-42s %s %s\n", i+1, r.Category, query, status, detail)
	}

	// --- Resume summary ---
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  RESUME / README SUMMARY")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf(`
- Retrieval Precision@%d : %.1f%% across %d test queries
- Avg Retrieval Latency  : %dms
- Avg Similarity Score   : %.3f
- Embedding Provider     : %s
`,
		cfg.TopK, precision, validCount,
		avgRetMs,
		avgSim,
		cfg.EmbedProvider.Name(),
	)
	if cfg.RunLLM {
		fmt.Printf("• LLM Provider           : %s\n", cfg.LLMProvider.Name())
	}
	fmt.Println(strings.Repeat("=", 60))
}