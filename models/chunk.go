package models

type Chunk struct {
	ID string `json:"id"`
	RepoURL string `json:"repo_url"`
	FilePath string `json:"file_path"`
	Language string `json:"language"`
	Content string `json:"content"`
	StartLine int `json:"start_line"`
	EndLine int `json:"end_line"`
}