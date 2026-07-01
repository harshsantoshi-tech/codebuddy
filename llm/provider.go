package llm

type LLMProvider interface{
	Generate(prompt string )(*AnswerResult , error)

	Name() string
}

type AnswerResult struct{
	Answer string `json:"answer"`
	InputTokens int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	LatencyMs int64 `json:"latency_ms"`
	Provider string `json:"provider"`
}