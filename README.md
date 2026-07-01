# CodeBuddy 🤖

> Ask questions about any GitHub repository and get grounded answers with source citations.

Built with **Go + Gin**, **RAG (Retrieval-Augmented Generation)**, **ChromaDB**, and a
pluggable strategy layer supporting both **OpenAI** and **Ollama (local/free)**.

---

## The Problem

Reading a new codebase is slow. You clone a repo, grep around, jump between files, and
still struggle to understand how things connect. CodeBuddy lets you just *ask* — and get
answers grounded in actual source code with file + line citations.

---

## Demo

```bash
# 1. Ingest a repo
curl -X POST http://localhost:8080/api/v1/ingest \
  -H "Content-Type: application/json" \
  -d '{"repo_url": "https://github.com/joho/godotenv"}'

# 2. Ask a question
curl -X POST http://localhost:8080/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "repo_url": "https://github.com/joho/godotenv",
    "query": "how does the parser handle quoted strings?",
    "provider": "ollama"
  }'
```

**Response:**
```json
{
  "answer": "Quoted strings are handled in the parseValue function [1].
             Single and double quotes are both supported. The parser
             strips surrounding quotes and processes escape sequences
             like \\n inside double-quoted values [2].",
  "citations": [
    { "index": 1, "file_path": "godotenv.go", "start_line": 124, "end_line": 174, "similarity": 0.66 },
    { "index": 2, "file_path": "godotenv.go", "start_line": 175, "end_line": 225, "similarity": 0.61 }
  ],
  "retrieval_time_ms": 147,
  "llm_time_ms": 48003,
  "total_time_ms": 48150,
  "embedding_provider": "ollama-nomic-embed-text",
  "provider": "ollama-gemma4"
}
```

---

## Architecture

```
POST /api/v1/ingest                    POST /api/v1/query
       │                                      │
       ▼                                      ▼
  Clone Repo (git --depth=1)        Embed Question
       │                            (EmbeddingProvider)
       ▼                                      │
  Walk + Filter Files                         ▼
       │                             ChromaDB Vector Search
       ▼                                      │
  Chunk Code                                  ▼
  (50 lines, 10 overlap)            Top-K Chunks + Similarity
       │                                      │
       ▼                                      ▼
  Embed Chunks                      LLM Generation
  (EmbeddingProvider)               (LLMProvider)
       │                                      │
       ▼                                      ▼
  Store in ChromaDB               Answer + Citations JSON
```

### Strategy Pattern — Pluggable Providers

Both embedding and LLM layers sit behind Go interfaces.
Swap providers with a single `.env` change — zero code changes.

| Layer | Interface | OpenAI | Ollama |
|---|---|---|---|
| Embeddings | `EmbeddingProvider` | `text-embedding-3-small` (1536-dim) | `nomic-embed-text` (768-dim) |
| LLM | `LLMProvider` | `gpt-4o` | `llama3.1` / `gemma4` |

---

## Benchmark Results

Evaluated on **20 test queries** across 4 categories against `joho/godotenv`.
Run fully locally using **Ollama — zero API cost.**

### Retrieval Metrics (Ollama — nomic-embed-text)

| Metric | Value |
|---|---|
| Overall Precision@5 | **60%** (12/20) |
| Avg Retrieval Latency | **147ms** |
| Avg Similarity Score | **0.545** |

### Precision by Category

| Category | Score | Result |
|---|---|---|
| Bug investigation | 4/5 | ████░ 80% |
| Architecture | 3/5 | ███░░ 60% |
| Fuzzy / vague queries | 3/5 | ███░░ 60% |
| Feature questions | 2/5 | ██░░░ 40% |

### LLM Metrics (Ollama — gemma4)

| Metric | Value |
|---|---|
| Avg Generation Latency | 48,003ms |
| Avg Tokens per Query | 3,113 |
| Total Pipeline Latency | ~48s |
| Cost per Query | **$0.00** |

> Generation latency is high due to CPU-only local inference.
> Swap `LLM_PROVIDER=openai` for ~2s responses at ~$0.003/query.

### Per-Question Breakdown

| # | Category | Query | Result | Latency | Similarity |
|---|---|---|---|---|---|
| 1 | feature | how are .env files parsed? | ✅ | 60ms | 0.66 |
| 2 | feature | how do you load environment variables? | ❌ | 229ms | 0.59 |
| 3 | feature | how does the Read function work? | ✅ | 152ms | 0.58 |
| 4 | feature | how are quoted strings handled? | ❌ | 84ms | 0.65 |
| 5 | feature | what happens when a variable is already set? | ❌ | 184ms | 0.53 |
| 6 | bug | where could parsing fail? | ✅ | 328ms | 0.61 |
| 7 | bug | how are errors returned to the caller? | ✅ | 96ms | 0.51 |
| 8 | bug | what happens with malformed env file lines? | ✅ | 120ms | 0.61 |
| 9 | bug | how are comments handled? | ❌ | 119ms | 0.49 |
| 10 | bug | what happens if the file does not exist? | ✅ | 84ms | 0.46 |
| 11 | architecture | what are the main exported functions? | ❌ | 214ms | 0.55 |
| 12 | architecture | how is the package structured? | ❌ | 125ms | 0.51 |
| 13 | architecture | how does Overload differ from Load? | ✅ | 140ms | 0.58 |
| 14 | architecture | how does the Write function work? | ✅ | 188ms | 0.58 |
| 15 | architecture | how are multiple env files handled? | ✅ | 111ms | 0.62 |
| 16 | fuzzy | export | ❌ | 148ms | 0.49 |
| 17 | fuzzy | parse environment | ✅ | 245ms | 0.58 |
| 18 | fuzzy | file not found | ✅ | 141ms | 0.46 |
| 19 | fuzzy | overwrite existing | ❌ | 88ms | 0.42 |
| 20 | fuzzy | write back to disk | ✅ | 93ms | 0.42 |

---

## Tech Stack

| Component | Choice | Why |
|---|---|---|
| Language | Go + Gin | Fast, low memory, strong typing |
| Vector DB | ChromaDB | Local, persistent, simple HTTP API |
| Embeddings | OpenAI / Ollama | Pluggable via Strategy pattern |
| LLM | GPT-4o / Ollama | Swappable via env var |
| Chunking | 50-line overlapping windows | Language-agnostic, overlap preserves context |
| Similarity | Cosine distance | Robust to chunk length variation |

---

## Running Locally

**Prerequisites:** Go 1.21+, Docker, Git

```bash
# 1. Clone
git clone https://github.com/harshsantoshi-tech/codebuddy
cd codebuddy

# 2. Setup env
cp .env.example .env
# Add OPENAI_API_KEY if using OpenAI, otherwise leave blank for Ollama

# 3. Start ChromaDB
docker run -d --name chromadb -p 8000:8000 \
  -v chroma-data:/chroma/chroma chromadb/chroma:latest

# 4. Start Ollama (for local inference)
brew install ollama
ollama serve
ollama pull nomic-embed-text
ollama pull gemma4          # or llama3.1

# 5. Configure .env for fully local setup
echo "EMBEDDING_PROVIDER=ollama" >> .env
echo "LLM_PROVIDER=ollama" >> .env
echo "OLLAMA_EMBED_MODEL=nomic-embed-text" >> .env
echo "OLLAMA_MODEL=gemma4" >> .env

# 6. Start server
go run main.go

# 7. Ingest a repo
curl -X POST http://localhost:8080/api/v1/ingest \
  -H "Content-Type: application/json" \
  -d '{"repo_url": "https://github.com/joho/godotenv"}'

# 8. Query it
curl -X POST http://localhost:8080/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "repo_url": "https://github.com/joho/godotenv",
    "query": "how does Load work?"
  }'

# 9. Run benchmark
go run cmd/benchmark/main.go \
  -repo https://github.com/joho/godotenv \
  -embed ollama -llm ollama \
  -topk 5 -llm-eval
```

---

## API Reference

### `POST /api/v1/ingest`

| Field | Type | Required | Description |
|---|---|---|---|
| `repo_url` | string | ✅ | GitHub repo URL to index |

### `POST /api/v1/query`

| Field | Type | Required | Description |
|---|---|---|---|
| `repo_url` | string | ✅ | Same repo as ingested |
| `query` | string | ✅ | Natural language question |
| `top_k` | int | ❌ | Chunks to retrieve (default: 5) |
| `provider` | string | ❌ | Override LLM: `openai` or `ollama` |

### `GET /health`

Returns server status and timestamp.

---

## Design Decisions

**Overlapping chunks over AST parsing** — 50-line windows with 10-line overlap is
language-agnostic and handles function boundaries without needing a parser per language.

**Shallow clone (`--depth=1`)** — only latest snapshot, 10-50x faster than full clone.
We don't need git history for Q&A.

**Provider-isolated ChromaDB collections** — OpenAI (1536-dim) and Ollama (768-dim)
vectors live in separate collections to prevent vector space collision.

**Strategy pattern over if/else** — adding a new provider means implementing one
interface and one factory entry. Zero changes to handlers or business logic.