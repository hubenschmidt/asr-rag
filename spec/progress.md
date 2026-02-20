# ASR-RAG Progress

| Phase | Description               | Status  |
| ----- | ------------------------- | ------- |
| 1     | Project scaffold + config | done    |
| 2     | Data model + corpus       | done    |
| 3     | Embedder client           | done    |
| 4     | Vector DB client          | pending |
| 5     | Search command            | pending |
| 6     | ASR client                | pending |
| 7     | LLM corrector             | pending |
| 8     | Full pipeline wiring      | pending |
| 9     | Mic recorder              | pending |

## Log

- **Phase 1** — `go.mod`, `.env`, `docker-compose.yml`, `main.go` with map dispatch + stubs. Compiles, CLI routes correctly.
- **Phase 2** — `corpus.json` with 61 Go terms, `corpusEntry` struct, `loadCorpus()` using `os.ReadFile`.
- **Phase 3** — `internal/embedder/ollama.go` — Ollama embed client, verified 768-dim vectors via `go run . seed`.
