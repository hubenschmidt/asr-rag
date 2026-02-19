# ASR-RAG: Go Jargon Correction via RAG

Teaching PoC that demonstrates how a RAG pipeline can improve ASR transcription quality for Go programming terminology. Whisper often mishears Go-specific jargon (e.g., "go routines" instead of "goroutines"). By retrieving relevant terminology from a vector database and feeding it to an LLM, we correct these errors post-transcription.

## Architecture

```mermaid
flowchart LR
    WAV["WAV File"]:::input -- "file path" --> ASR["Whisper Server"]:::asr
    ASR -- "raw transcript" --> EMB["Ollama Embed"]:::embed
    EMB -- "query vector" --> QD["Qdrant Search"]:::vectordb
    QD -- "top-5 Go terms" --> LLM["Ollama LLM"]:::llm
    ASR -. "raw transcript" .-> LLM
    LLM -- "corrected text" --> OUT["Output"]:::output

    classDef input fill:#3b82f6,stroke:#1e40af,color:#fff
    classDef asr fill:#f59e0b,stroke:#b45309,color:#fff
    classDef embed fill:#6366f1,stroke:#4338ca,color:#fff
    classDef vectordb fill:#ef4444,stroke:#b91c1c,color:#fff
    classDef llm fill:#10b981,stroke:#047857,color:#fff
    classDef output fill:#64748b,stroke:#334155,color:#fff
```

## RAG Pipeline Detail

```mermaid
flowchart TB
    subgraph Seed ["go run . seed"]
        CJ["corpus.json<br/>~60 Go terms"]:::input --> SE["Embed each term"]:::embed
        SE --> UP["Upsert to Qdrant"]:::vectordb
    end

    subgraph Transcribe ["go run . transcribe file.wav"]
        WAV["Read WAV"]:::input --> WH["Whisper /inference<br/>multipart POST"]:::asr
        WH --> RT["Raw Transcript"]:::asr
        RT --> QE["Embed transcript"]:::embed
        QE --> QS["Qdrant top-5 search"]:::vectordb
        QS --> SYS["Build system prompt<br/>with retrieved terms"]:::llm
        RT --> SYS
        SYS --> LLM["Ollama /api/chat"]:::llm
        LLM --> CT["Corrected Transcript"]:::output
    end

    classDef input fill:#3b82f6,stroke:#1e40af,color:#fff
    classDef asr fill:#f59e0b,stroke:#b45309,color:#fff
    classDef embed fill:#6366f1,stroke:#4338ca,color:#fff
    classDef vectordb fill:#ef4444,stroke:#b91c1c,color:#fff
    classDef llm fill:#10b981,stroke:#047857,color:#fff
    classDef output fill:#64748b,stroke:#334155,color:#fff
```

## Stack

| Component | Tool | Port |
|-----------|------|------|
| ASR | whisper-server (whisper.cpp) | :8178 |
| Embeddings | Ollama `nomic-embed-text` (768-dim) | :11434 |
| Vector DB | Qdrant (Docker, gRPC) | :6334 |
| LLM | Ollama `llama3.2:3b` | :11434 |

## Color Legend

| Color | Component |
|-------|-----------|
| **Blue** | Input (WAV, corpus) |
| **Amber** | ASR — Whisper |
| **Indigo** | Embeddings — Ollama |
| **Red** | Vector DB — Qdrant |
| **Green** | LLM — Ollama |
| **Gray** | Output |

## Usage

```bash
# Start Qdrant
docker compose up -d

# Start whisper-server
~/.local/src/whisper.cpp/build/bin/whisper-server \
  -m ~/.local/src/whisper.cpp/models/ggml-medium.en.bin --port 8178

# Seed the vector DB with Go terminology
go run . seed

# Search for similar terms
go run . search "go routines"

# Transcribe and correct a WAV file
go run . transcribe sample.wav

# Record from mic and correct (default 5 seconds)
go run . record 5
```
