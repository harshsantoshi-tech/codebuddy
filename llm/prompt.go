package llm

import (
	"codebuddy/retrieval"
	"fmt"
	"strings"
)


type Citation struct{
	Index int `json:"index"`
	FilePath   string  `json:"file_path"`
	StartLine  int     `json:"start_line"`
	EndLine    int     `json:"end_line"`
	Similarity float32 `json:"similarity"`
}

func BuildPrompt(query string , chunks []retrieval.RetrievedChunk)string{
	var sb strings.Builder

	sb.WriteString("You are an expert code assistant analyzing a GitHub repository.\n")
	sb.WriteString("Answer the user's question using ONLY the code context provided below.\n")
	sb.WriteString("Cite sources using their number [1], [2], etc.\n")
	sb.WriteString("If the context doesn't contain enough information, say so clearly.\n")
	sb.WriteString("Be concise, technical, and precise.\n\n")

	sb.WriteString("--- RETRIEVED CODE CONTEXT ---\n\n")
	for i, chunk := range chunks {
		sb.WriteString(fmt.Sprintf(
			"[%d] File: %s | Lines: %d-%d | Language: %s | Relevance: %.2f\n",
			i+1,
			chunk.FilePath,
			chunk.StartLine,
			chunk.EndLine,
			chunk.Language,
			chunk.Similarity,
		))
		sb.WriteString("```" + chunk.Language + "\n")
		sb.WriteString(chunk.Content)
		sb.WriteString("\n```\n\n")
	}
	sb.WriteString("--- END CONTEXT ---\n\n")

	sb.WriteString(fmt.Sprintf("QUESTION: %s\n\n", query))
	sb.WriteString("ANSWER (cite sources as [1], [2], etc.):\n")

	return sb.String()

}

func BuildCitationSummary(chunks []retrieval.RetrievedChunk)[]Citation{
	citations := make([]Citation , len(chunks))

	for i , chunk := range chunks{
		citations[i] = Citation{
			Index: i + 1,
			FilePath: chunk.FilePath,
			StartLine: chunk.StartLine,
			EndLine: chunk.EndLine,
			Similarity: chunk.Similarity,
		}
	}
	return citations
}