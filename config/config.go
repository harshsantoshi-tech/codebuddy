package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	OpenAIKey    string
	ChromaURL    string
	LLMProvider  string // "openai" or "ollama" — NEW
	EmbeddingProvider string
	OllamaURL    string 
	OllamaEmbedModel string
	OllamaModel  string 
	AnthropicKey string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment directly")
	}

	return &Config{
		Port:        getEnv("PORT", "8080"),
		OpenAIKey:   getEnv("OPENAI_API_KEY", ""),
		ChromaURL:   getEnv("CHROMA_URL", "http://localhost:8000"),
		LLMProvider: getEnv("LLM_PROVIDER", "ollama"), // default to ollama
		EmbeddingProvider: getEnv("EMBEDDING_PROVIDER" ,"ollama"),
		OllamaURL:   getEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaEmbedModel: getEnv("OLLAMA_EMBED_MODEL" , "nomic-embed-text"),
		OllamaModel: getEnv("OLLAMA_MODEL", "gemma4"),
		AnthropicKey : getEnv("ANTHROPIC_KEY",""),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// Getters used by handlers
func (c *Config)GetPort()string { return c.Port}
func (c *Config) GetOpenApiKey() string   { return c.OpenAIKey }
func (c *Config) GetChromaURL() string   { return c.ChromaURL }
func (c *Config) GetLLMProvider() string { return c.LLMProvider }
func (c *Config)GetEmbeddingProvider() string{ return c.EmbeddingProvider}
func (c *Config)GetOllamaEmbedModel() string { return c.OllamaEmbedModel}
func (c *Config) GetOllamaURL() string   { return c.OllamaURL }
func (c *Config) GetOllamaModel() string { return c.OllamaModel }
func (c *Config) GetAnthropicKey() string { return c.AnthropicKey}