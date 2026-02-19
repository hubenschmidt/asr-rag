// package main_ex

// import (
// 	"encoding/json"
// 	"fmt"
// 	"os"
// )

// type command struct {
// 	usage string
// 	run   func(args []string) error
// }

// func main_ex() {
// 	cfg, err := loadConfig("config.json")
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "config: %v\n", err)
// 		os.Exit(1)
// 	}
// 	_ = cfg // used in later phases

// 	commands := map[string]command{
// 		"seed": {
// 			usage: "seed — embed corpus and upsert to Qdrant",
// 			run:   func(args []string) error { fmt.Println("seed: not yet implemented"); return nil },
// 		},
// 		"transcribe": {
// 			usage: "transcribe <file.wav> — transcribe and correct a WAV file",
// 			run:   func(args []string) error { fmt.Println("transcribe: not yet implemented"); return nil },
// 		},
// 		"record": {
// 			usage: "record [seconds] — record from mic, transcribe, and correct",
// 			run:   func(args []string) error { fmt.Println("record: not yet implemented"); return nil },
// 		},
// 		"search": {
// 			usage: "search <query> — search Qdrant for similar Go terms",
// 			run:   func(args []string) error { fmt.Println("search: not yet implemented"); return nil },
// 		},
// 	}

// 	if len(os.Args) < 2 {
// 		printUsage(commands)
// 		os.Exit(1)
// 	}

// 	name := os.Args[1]
// 	cmd, ok := commands[name]
// 	if !ok {
// 		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", name)
// 		printUsage(commands)
// 		os.Exit(1)
// 	}

// 	if err := cmd.run(os.Args[2:]); err != nil {
// 		fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
// 		os.Exit(1)
// 	}
// }

// func printUsage(commands map[string]command) {
// 	fmt.Fprintln(os.Stderr, "usage: asr-rag <command> [args]")
// 	fmt.Fprintln(os.Stderr, "\ncommands:")
// 	for _, cmd := range commands {
// 		fmt.Fprintf(os.Stderr, "  %s\n", cmd.usage)
// 	}
// }

// type config struct {
// 	WhisperURL string `json:"whisper_url"`
// 	OllamaURL  string `json:"ollama_url"`
// 	QdrantURL  string `json:"qdrant_url"`
// }

// func loadConfig(path string) (*config, error) {
// 	data, err := os.ReadFile(path)
// 	if err != nil {
// 		return nil, fmt.Errorf("read %s: %w", path, err)
// 	}
// 	var cfg config
// 	if err := json.Unmarshal(data, &cfg); err != nil {
// 		return nil, fmt.Errorf("parse %s: %w", path, err)
// 	}
// 	return &cfg, nil
// }
