package embeddings

import (
	"codebuddy/models"
	"fmt"
	"time"
)

const BatchSize = 100

type EmbeddedChunk struct{
	models.Chunk
	Vector []float32
	EmbeddingProvider string //which provider produced this vector
}

func BatchEmbed(provider EmbeddingProvider , chunks []models.Chunk)([]EmbeddedChunk , int  , error){
	var results []EmbeddedChunk

	totalTokens := 0

	for i := 0 ; i < len(chunks) ; i+= BatchSize{
		end := i + BatchSize
		end = min(end , len(chunks))

		batch := chunks[i:end]

		fmt.Printf("[%s] Embedding batch %d-%d of %d chunks ...\n" ,provider.Name(), i + 1 , end , len(chunks))

		texts := make([]string , len(batch))

		for j , chunk := range batch{
			texts[j] = fmt.Sprintf("%s %s\n%s" , chunk.Language , chunk.FilePath  ,chunk.Content)

		}

		vectors , tokens , err := provider.EmbedTexts(texts)
		if err != nil {
			return nil , totalTokens , fmt.Errorf("batch %d failed : %w",i/BatchSize , err)
		}

		totalTokens += tokens

		for j , chunk := range batch{
			results = append(results , EmbeddedChunk{
				Chunk: chunk,
				Vector: vectors[j],
				EmbeddingProvider: provider.Name(),
			})
		}

		//pause between batches for rate limit protection
		//change it to 20 sec if want to use free one
		if end < len(chunks){
			time.Sleep(200 * time.Millisecond)
		}
	}
	fmt.Printf("[%s] Embedded %d chunks using %d tokens\n" ,provider.Name(), len(results) , totalTokens)
	return results , totalTokens , nil
}