package embeddings

type EmbeddingProvider interface{
	EmbedTexts (texts []string)(vectors [][]float32 , totalTokens int , err error)

	Name()string

	Dimensions()int
}