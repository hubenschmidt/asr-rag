# ASR-RAG PoC: Go Jargon Correction via RAG

## Pipeline

```
WAV → whisper-server → raw transcript → embed → Qdrant search → LLM correction → corrected output
```

## Stack (all local, no API keys)

| Component | Tool | Notes |
|-----------|------|-------|
| ASR | whisper-server HTTP API | `~/.local/src/whisper.cpp/build/bin/whisper-server` |
| Embeddings | Ollama `nomic-embed-text` (768-dim) | Already installed |
| Vector DB | Qdrant (Docker, gRPC on :6334) | `github.com/qdrant/go-client` |
| LLM | Ollama `llama3.2:3b` | Already installed |

## Project Structure

```
asr-rag/
├── main.go                    # CLI: seed | transcribe <file.wav> | record [seconds] | search <query>
├── corpus.json                # ~60 Go term/definition pairs (go:embed)
├── internal/
│   ├── asr/whisper.go         # whisper-server multipart POST client
│   ├── recorder/mic.go        # PortAudio mic capture → WAV bytes
│   ├── embedder/ollama.go     # Ollama /api/embed client
│   ├── vectordb/qdrant.go     # Qdrant gRPC: EnsureCollection, Upsert, Search
│   └── corrector/llm.go       # Ollama /api/chat for transcript correction
├── docker-compose.yml         # Qdrant only
└── .env                       # WHISPER_URL, OLLAMA_URL, QDRANT_URL
```

## Build Phases (bottom-up, one at a time)

Each phase is a reviewable unit. We pause after each for review before continuing.

### Phase 1 — Project scaffold + config
- `go mod init`, `.env`, `docker-compose.yml`
- `main.go` with map-based command dispatch (stubs only)
- **Deliverables:** project compiles, `docker compose up -d` starts Qdrant

### Phase 2 — Data model + corpus
- `corpus.json`: ~60 Go term/definition pairs
- Go struct for corpus entries, `go:embed` to bake into binary
- **Deliverables:** `corpus.json` exists, struct loads and parses at startup

### Phase 3 — Embedder client (`internal/embedder/ollama.go`)
- POST to `{OLLAMA_URL}/api/embed`
- Single-text and batch embed functions returning `[]float32`
- **Deliverables:** `go run . seed` can embed one test string (print dim count)

### Phase 4 — Vector DB client (`internal/vectordb/qdrant.go`)
- Qdrant gRPC client: `EnsureCollection` (768-dim, cosine), `Upsert`, `Search`
- **Deliverables:** `go run . seed` embeds all corpus terms → upserts to Qdrant

### Phase 5 — Search command
- `go run . search <query>` — embed query → Qdrant top-5 → print results
- **Deliverables:** `go run . search "go routines"` returns goroutine-related hits

### Phase 6 — ASR client (`internal/asr/whisper.go`)
- Multipart WAV POST to whisper-server `/inference`, returns text
- Simplified from reference: takes file path, reads bytes, posts directly
- **Deliverables:** `go run . transcribe sample.wav` prints raw transcript

### Phase 7 — LLM corrector (`internal/corrector/llm.go`)
- POST to `{OLLAMA_URL}/api/chat` with system prompt containing retrieved terms
- **Deliverables:** corrector function callable with (transcript, []terms) → corrected text

### Phase 8 — Full pipeline wiring
- `transcribe` command: WAV → whisper → embed → search → LLM correct → print raw vs corrected
- **Deliverables:** end-to-end demo working

### Phase 9 — Mic recorder (`internal/recorder/mic.go`)
- PortAudio 16kHz mono capture → WAV bytes
- `record [seconds]` command feeds into same pipeline as `transcribe`
- **Deliverables:** `go run . record 5` captures and corrects live speech

## Key Patterns
- Guard clauses / early returns (no nested conditionals)
- Map dispatch instead of switch/case
- Multipart WAV upload pattern from reference at `~/asr-llm-tts-poc/services/gateway/internal/pipeline/asr.go:127-153`

## Verification
1. `docker compose up -d` — Qdrant dashboard at `localhost:6333/dashboard`
2. Start whisper-server: `~/.local/src/whisper.cpp/build/bin/whisper-server -m ~/.local/src/whisper.cpp/models/ggml-medium.en.bin --port 8178`
3. `go run . seed` — N terms embedded and upserted
4. `go run . search "go routines"` — goroutine-related results
5. `go run . transcribe sample.wav` — raw vs corrected side-by-side
