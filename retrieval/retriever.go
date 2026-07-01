package retrieval

import (
	"codebuddy/embeddings"
	"codebuddy/ingestion"
	"codebuddy/vectorstore"
	"fmt"
	"time"
)

const DefaultTopK = 5

type RetrievedChunk struct {
	FilePath   string  `json:"file_path"`
	Language   string  `json:"language"`
	StartLine  int     `json:"start_line"`
	EndLine    int     `json:"end_line"`
	Content    string  `json:"content"`
	Similarity float32 `json:"similarity"`
}

type RetrievalResult struct {
	Query          string           `json:"query"`
	Chunks         []RetrievedChunk `json:"chunks"`
	RetrievalTimeMs int64           `json:"retrieval_time_ms"` // resume metric
	EmbeddingProvider string 		`json:"embedding_provider"`
}

func Retrieve(embedProvider embeddings.EmbeddingProvider , chromaClient *vectorstore.ChromaClient , repoURL string , query string , topK int)(*RetrievalResult , error){
	start := time.Now()

	if topK <= 0{
		topK = DefaultTopK
	}

	fmt.Printf("[%s] Embedding query : %s\n" ,embedProvider.Name(), query)

	vectors , _ , err := embedProvider.EmbedTexts([]string{query})
	if err != nil{
		return nil , fmt.Errorf("failed to embed query: %w" , err)
	}

	queryVector := vectors[0]

	repoName := ingestion.ExtractRepoName(repoURL)

	collectionName := collectionName(repoName , embedProvider.Name())

	collectionID , err := chromaClient.GetCollectionID(collectionName)
	if err != nil {
		return nil , fmt.Errorf("repo not indexed yet - run/injest first : %w" , err)
	}

	fmt.Printf("Searching collection %s for top-%d chunks ... \n" , repoName , topK)

	rawResults , err := chromaClient.QueryCollection(collectionID , queryVector , topK)
	if err != nil {
		return nil , fmt.Errorf("chroma query failed : %w" , err)

	}

	var chunks []RetrievedChunk

	for _ , r := range rawResults{
		similarity := float32(1) - r.Distance

		filePath , _ := r.Metadata["file_path"].(string)
		language , _ := r.Metadata["language"].(string)

		startLine := 0
		endLine := 0

		if sl , ok := r.Metadata["start_line"].(float64) ; ok{
			startLine = int(sl)
		}

		if el , ok := r.Metadata["end_line"].(float64) ; ok{
			endLine = int(el)
		}

		chunks = append(chunks , RetrievedChunk{
			FilePath: filePath,
			Language: language,
			StartLine: startLine,
			EndLine: endLine,
			Content: r.Document,
			Similarity: similarity,
		})
	}

	retrievalTime := time.Since(start).Milliseconds()

	fmt.Printf("Retrived %d chunks in %dms \n" , len(chunks) , retrievalTime)
	
	return &RetrievalResult{
		Query : query,
		Chunks: chunks,
		RetrievalTimeMs: retrievalTime,
		EmbeddingProvider: embedProvider.Name(),
	}, nil

}

func collectionName(repoName , providerName string)string{
	return repoName + "-" + providerName
}