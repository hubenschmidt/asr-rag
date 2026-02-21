# ASR-RAG Progress

| Phase | Description               | Status  |
| ----- | ------------------------- | ------- |
| 1     | Project scaffold + config | done    |
| 2     | Data model + corpus       | done    |
| 3     | Embedder client           | done    |
| 4     | Vector DB client          | done    |
| 5     | Search command            | done    |
| 6     | ASR client                | done    |
| 7     | LLM corrector             | done    |
| 8     | Full pipeline wiring      | pending |
| 9     | Mic recorder              | pending |

## Log

- **Phase 1** — `go.mod`, `.env`, `docker-compose.yml`, `main.go` with map dispatch + stubs. Compiles, CLI routes correctly.
- **Phase 2** — `corpus.json` with 61 Go terms, `corpusEntry` struct, `loadCorpus()` using `os.ReadFile`.
- **Phase 3** — `internal/embedder/ollama.go` — Ollama embed client, verified 768-dim vectors via `go run . seed`.
- **Phase 4** — `internal/vectordb/qdrant.go` — Qdrant gRPC client with EnsureCollection, Upsert, Search. Seed command wired end-to-end.
- **Phase 5** — Search command wired: embed query → Qdrant top-5 → print scored results.
- **Phase 6** — `internal/asr/whisper.go` — multipart WAV POST to whisper-server, transcribe command wired.
- **Phase 7** — `internal/corrector/llm.go` — Ollama /api/chat client for transcript correction with Go term context.
