package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	cfg, err := loadConfig("configg.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(*cfg) // *cfg dereferences and prints {http://localhost:8178}
}

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
