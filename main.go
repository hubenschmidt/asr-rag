package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// config holds service URLs loaded from config.json.
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

func main() {
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

	commands := buildCommands(cfg, cps)

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
		fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
		os.Exit(1)
	}
}
