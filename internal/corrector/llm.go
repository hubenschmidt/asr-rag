package corrector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client talks to Ollama's /api/chat endpoint to correct transcripts.
// Fields: url string, model string
type Client struct {
	url   string
	model string
}

// New creates a corrector client.
// Takes: ollama base URL, model name (e.g. "llama3.2:3b")
func New(url string, model string) *Client {
	return &Client{url: url, model: model}
}

// chatRequest is the JSON body sent to Ollama /api/chat.
// JSON shape:
//
//	{
//	  "model": "llama3.2:3b",
//	  "messages": [
//	    {"role": "system", "content": "..."},
//	    {"role": "user", "content": "..."}
//	  ],
//	  "stream": false
//	}
type chatRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse is the JSON returned by Ollama.
type chatResponse struct {
	Message message `json:"message"`
}

// Term represents a Go term retrieved from Qdrant.
// Passed in by the caller so this package doesn't depend on vectordb.
type Term struct {
	Name       string
	Definition string
}

// Correct takes a raw transcript and relevant Go terms, asks the LLM to fix misheard jargon.
func (c *Client) Correct(transcript string, terms []Term) (string, error) {
	// Build system prompt with retrieved Go terms as context
	prompt := "Fix misheard Go terms in this transcript using the provided reference. Return only the corrected text.\n\nReference terms:\n"
	for _, t := range terms {
		prompt += "- " + t.Name + ": " + t.Definition + "\n"
	}

	// Build the chat request with system + user messages
	body, err := json.Marshal(chatRequest{
		Model:  c.model,
		Stream: false,
		Messages: []message{
			{Role: "system", Content: prompt},
			{Role: "user", Content: transcript},
		},
	})
	if err != nil {
		return "", fmt.Errorf("marshal chat request: %w", err)
	}

	// POST to Ollama /api/chat
	resp, err := http.Post(c.url+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("chat request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("chat status %d: %s", resp.StatusCode, respBody)
	}

	// Decode the LLM response
	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode chat response: %w", err)
	}

	return result.Message.Content, nil
}
