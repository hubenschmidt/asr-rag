package main

import (
	"context" // qdrant will not compile without `context`
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/hubenschmidt/asr-rag/internal/asr"
	"github.com/hubenschmidt/asr-rag/internal/corrector"
	"github.com/hubenschmidt/asr-rag/internal/embedder"
	"github.com/hubenschmidt/asr-rag/internal/recorder"
	"github.com/hubenschmidt/asr-rag/internal/vectordb"
)

// config holds service URLs loaded from config.json.
// Struct tags map JSON keys to Go fields.
type config struct {
	WhisperURL string `json:"whisper_url"`
	OllamaURL  string `json:"ollama_url"`
	QdrantURL  string `json:"qdrant_url"`
}

// loadConfig reads a JSON file and returns a pointer to a populated config.
func loadConfig(path string) (*config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file at %s: %w", path, err)
	}

	var cfg config

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error unmarshalling: %w", err)
	}

	return &cfg, nil
}

// corpusEntry represents a single Go term and its definition from corpus.json.
type corpusEntry struct {
	Term       string `json:"term"`
	Definition string `json:"definition"`
}

// loadCorpus reads corpus.json and returns a slice of term/definition pairs.
func loadCorpus(path string) (*[]corpusEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var entries []corpusEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	return &entries, nil
}

// command maps a CLI subcommand name to its usage text and run function.
// run returns error so main can handle failures uniformly.
type command struct {
	usage string
	run   func(args []string) error
}

func main() {
	// Load config and corpus at startup — both are needed by all commands.
	cfg, err := loadConfig("config.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	cps, err := loadCorpus("corpus.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpus: %v\n", err)
		os.Exit(1)
	}

	// Command map — keys are CLI subcommand names, values define behavior.
	// Map dispatch avoids nested if/switch (guard clause pattern).
	commands := map[string]command{
		"seed": {
			usage: "seed -- embed corpus and upsert to Qdrant",
			run: func(args []string) error {
				// Create embedder and vector DB clients from config URLs
				emb := embedder.New(cfg.OllamaURL, "qwen3-embedding:8b")

				db, err := vectordb.New(cfg.QdrantURL)
				if err != nil {
					return err
				}
				defer db.Close() // clean up gRPC connection when seed finishes

				// Create collection if it doesn't exist (idempotent)
				if err := db.EnsureCollection(context.Background()); err != nil {
					return err
				}

				// Embed each corpus entry and store in Qdrant
				for i, entry := range *cps {
					vec, err := emb.Embed(entry.Term + ": " + entry.Definition)
					if err != nil {
						return err
					}
					if err := db.Upsert(context.Background(), uint64(i), vec, entry.Term, entry.Definition); err != nil {
						return err
					}
					fmt.Println("seeded:", entry.Term)
				}

				fmt.Println("done:", len(*cps), "terms seeded")
				return nil
			},
		},
		"transcribe": {
			usage: "transcribe <file.wav> -- transcribe and correct a WAV file",
			run: func(args []string) error {
				if len(args) < 1 {
					return fmt.Errorf("usage: transcribe <file.wav>")
				}

				// WAV → whisper → raw transcript
				whisper := asr.New(cfg.WhisperURL)
				text, err := whisper.Transcribe(args[0])
				if err != nil {
					return err
				}
				fmt.Println("raw:", text)

				// Embed the raw transcript
				emb := embedder.New(cfg.OllamaURL, "qwen3-embedding:8b")
				vec, err := emb.Embed(text)
				if err != nil {
					return err
				}

				// Search Qdrant for the 5 most similar Go terms
				db, err := vectordb.New(cfg.QdrantURL)
				if err != nil {
					return err
				}
				defer db.Close()

				results, err := db.Search(context.Background(), vec, 10)
				if err != nil {
					return err
				}

				// Convert search results to corrector terms
				terms := make([]corrector.Term, len(results))
				for i, r := range results {
					terms[i] = corrector.Term{Name: r.Term, Definition: r.Definition}
				}

				// LLM correction using retrieved terms as context
				llm := corrector.New(cfg.OllamaURL, "llama3.2:3b")
				corrected, err := llm.Correct(text, terms)
				if err != nil {
					return err
				}

				fmt.Println("corrected:", corrected)
				return nil
			},
		},
		"record": {
			usage: "record [seconds] -- record from mic, transcribe, and correct",
			run: func(args []string) error {
				// Parse optional seconds arg, default 5
				seconds := 5
				if len(args) > 0 {
					n, err := strconv.Atoi(args[0])
					if err != nil {
						return fmt.Errorf("invalid seconds: %w", err)
					}
					seconds = n
				}

				// Record from mic → WAV file
				tmp := "recording.wav"
				if err := recorder.Record(seconds, tmp); err != nil {
					return err
				}
				defer os.Remove(tmp)

				// WAV → whisper → raw transcript
				whisper := asr.New(cfg.WhisperURL)
				text, err := whisper.Transcribe(tmp)
				if err != nil {
					return err
				}
				fmt.Println("raw:", text)

				// Embed the raw transcript
				emb := embedder.New(cfg.OllamaURL, "qwen3-embedding:8b")
				vec, err := emb.Embed(text)
				if err != nil {
					return err
				}

				// Search Qdrant for the 5 most similar Go terms
				db, err := vectordb.New(cfg.QdrantURL)
				if err != nil {
					return err
				}
				defer db.Close()

				results, err := db.Search(context.Background(), vec, 10)
				if err != nil {
					return err
				}

				// Convert search results to corrector terms
				terms := make([]corrector.Term, len(results))
				for i, r := range results {
					terms[i] = corrector.Term{Name: r.Term, Definition: r.Definition}
				}

				fmt.Println("corrector search terms: ", terms)

				// LLM correction using retrieved terms as context
				llm := corrector.New(cfg.OllamaURL, "llama3.2:3b")
				corrected, err := llm.Correct(text, terms)
				if err != nil {
					return err
				}

				fmt.Println("corrected:", corrected)
				return nil
			},
		},
		"search": {
			usage: "search <query> -- search Qdrant for similar Go terms",
			run: func(args []string) error {
				if len(args) < 1 {
					return fmt.Errorf("usage: search <query>")
				}
				query := args[0]

				// Embed the query text into a vector
				emb := embedder.New(cfg.OllamaURL, "qwen3-embedding:8b")
				vec, err := emb.Embed(query)
				if err != nil {
					return err
				}

				// Search Qdrant for the 5 most similar terms
				db, err := vectordb.New(cfg.QdrantURL)
				if err != nil {
					return err
				}
				defer db.Close()

				results, err := db.Search(context.Background(), vec, 10)
				if err != nil {
					return err
				}

				// Print results
				for _, r := range results {
					fmt.Printf("%.4f  %s — %s\n", r.Score, r.Term, r.Definition)
				}
				return nil
			},
		},
	}

	// Validate CLI args — need at least a subcommand name
	if len(os.Args) < 2 {
		printUsage(commands)
		os.Exit(1)
	}

	// Look up command by name, run it, handle errors
	name := os.Args[1]
	cmd, ok := commands[name]
	if !ok {
		fmt.Println("unknown command: " + name)
		printUsage(commands)
		os.Exit(1)
	}
	if err := cmd.run(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
		os.Exit(1)
	}
}

func printUsage(commands map[string]command) {
	fmt.Println("usage: asr-rag <command> [args]")
	fmt.Println("\n commands:")
	for _, cmd := range commands {
		fmt.Println("  " + cmd.usage)
	}
}
