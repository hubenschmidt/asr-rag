package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hubenschmidt/asr-rag/internal/embedder"
)

type config struct {
	WhisperURL string `json:"whisper_url"`
	OllamaURL  string `json:"ollama_url"`
	QdrantURL  string `json:"qdrant_url"`
}

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

type corpusEntry struct {
	Term       string `json:"term"`
	Definition string `json:"definition"`
}

func loadCorpus(path string) (*[]corpusEntry, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		fmt.Println("Failed to load corpus")
	}

	var entries []corpusEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	return &entries, nil // nil is returned in error interface
}

type command struct {
	usage string
	run   func(args []string) error
}

func main() {
	cfg, err := loadConfig("config.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	cps, err := loadCorpus("corpus.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpus: %v\n", err)
	}

	commands := map[string]command{ // e.g. command["seed"]run // this is like Python dict[str, Command] or typescript Record<string, Command>
		"seed": {
			usage: "seed -- embed corpus and upsert to Qdrant",
			run: func(args []string) error {
				emb := embedder.New(cfg.OllamaURL, "nomic-embed-text")

				for _, entry := range *cps {
					vec, err := emb.Embed(entry.Term + ": " + entry.Definition)
					if err != nil {
						return err
					}
					fmt.Println(entry.Term, "dims =", len(vec))
				}

				return nil
			},
		},
		"transcribe": {
			usage: "transcribe <file.wav> -- transcribe and correcta WAV file",
			run: func(args []string) error {
				fmt.Println("transcribe: not yet implemented")
				return nil
			},
		},
		"record": {
			usage: "record [seconds] -- record from mic, transcribe, and correct",
			run: func(args []string) error {
				fmt.Println("record: not yet implemented")
				return nil
			},
		},
		"search": {
			usage: "search <query> -- search Qdrant for similar Go terms",
			run: func(args []string) error {
				fmt.Println("search: not yet implemented")
				return nil
			},
		},
	}

	// if insufficient args, exit program
	if len(os.Args) < 2 {
		printUsage(commands)
		os.Exit(1)
	}

	name := os.Args[1]
	cmd, ok := commands[name]
	if !ok {
		fmt.Println("unknown command: " + name)
		printUsage(commands)
		os.Exit(1)
	}
	if err := cmd.run(os.Args[2:]); err != nil {
		// fmt.Println(name + ": " + err.Error()) // Println does not unpack .Error() so it must be done explictly
		fmt.Fprintf(os.Stderr, "%s: %v\n", name, err) // Fprintf uses format verbs
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
