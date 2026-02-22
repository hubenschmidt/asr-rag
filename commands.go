package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/hubenschmidt/asr-rag/internal/asr"
	"github.com/hubenschmidt/asr-rag/internal/corrector"
	"github.com/hubenschmidt/asr-rag/internal/embedder"
	"github.com/hubenschmidt/asr-rag/internal/recorder"
	"github.com/hubenschmidt/asr-rag/internal/vectordb"
)

// command maps a CLI subcommand name to its usage text and run function.
// run returns error so main can handle failures uniformly.
type command struct {
	usage string
	run   func(args []string) error
}

func printUsage(commands map[string]command) {
	fmt.Println("usage: asr-rag <command> [args]")
	fmt.Println("\n commands:")
	for _, cmd := range commands {
		fmt.Println(" " + cmd.usage)
	}
}

func buildCommands(cfg *config, cps *[]corpusEntry) map[string]command {
	return map[string]command{
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

				results, err := db.Search(context.Background(), vec, 5)
				if err != nil {
					return err
				}

				// Convert search results to corrector terms
				terms := make([]corrector.Term, len(results))
				for i, r := range results {
					terms[i] = corrector.Term{Name: r.Term, Definition: r.Definition}
				}

				fmt.Println("context terms:")
				for _, t := range terms {
					fmt.Printf("  • %s — %s\n", t.Name, t.Definition)
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

				results, err := db.Search(context.Background(), vec, 5)
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
}
