package ingestion

import (
	"codebuddy/models"
	"fmt"
	"os"
	"strings"
)

const (
	ChunkLines = 50 
	OverlapLines = 10
)

func ChunkFile(file FileInfo , repoURL string)([]models.Chunk ,error ){

	data , err := os.ReadFile(file.AbsPath)

	if err != nil {
		return nil , fmt.Errorf("failed to read %s: %w", file.AbsPath , err)
	}

	lines := strings.Split(string(data) , "\n")

	var chunks []models.Chunk
	chunkIndex := 0

	step := ChunkLines - OverlapLines

	for start := 0 ; start < len(lines) ; start += step{

		end := start + ChunkLines

		end = min(end , len(lines))

		chunkLines := lines[start : end]
		content := strings.Join(chunkLines , "\n")

		if strings.TrimSpace(content) == ""{
			continue
		}

		chunkId := fmt.Sprintf("%s__%s__%d",
			ExtractRepoName(repoURL),
			strings.ReplaceAll(file.RelPath , "/" , "_"),
			chunkIndex,
		)	
		chunkIndex ++ 

		chunks = append(chunks , models.Chunk{
			ID: chunkId,
			RepoURL: repoURL,
			FilePath: file.RelPath,
			Language: file.Language,
			Content: content,
			StartLine: start + 1,
			EndLine: end,
		})
	}
	return chunks , nil
}