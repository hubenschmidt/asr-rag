package embedder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client talks to Ollama's /api/embed endpoint to convert text into vectors.
// The vectors are 4096-dimensional floats (for qwen3-embedding:8b).
// These vectors capture semantic meaning — similar text produces similar vectors.
type Client struct {
	url   string // Ollama base URL, e.g. "http://localhost:11434"
	model string // embedding model name, e.g. "qwen3-embedding:8b"
}

// New creates an embedder client. This is Go's constructor pattern —
// a standalone function because no Client instance exists yet to call a method on.
func New(url string, model string) *Client {
	return &Client{url: url, model: model}
}

// embedRequest is the JSON body sent to Ollama.
// {"model": "qwen3-embedding:8b", "input": "goroutine: A lightweight thread..."}
type embedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// embedResponse is the JSON body returned by Ollama.
// {"embeddings": [[0.1, 0.2, ...]]} — array of arrays because Ollama supports batch,
// but we only send one input at a time so we use [0].
type embedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

// Embed sends text to Ollama and returns its vector representation.
// Used in two places:
//   - seed: embed each corpus entry, then store the vector in Qdrant
//   - search/transcribe: embed the query text, then find similar vectors in Qdrant

func (c *Client) Embed(text string) ([]float32, error) {
	// 1. Build the JSON request body
	embedReq := embedRequest{Model: c.model, Input: text}
	body, err := json.Marshal(embedReq)
	if err != nil {
		return nil, fmt.Errorf("error marshalling embed request: %w", err)
	}

	// 2. POST to Ollama's embed endpoint
	// here http.Post method expects body as 'io.Reader' which should implement Read() method.
	// So, bytes package will take care of that.
	b := bytes.NewReader(body)
	postURL := c.url + "/api/embed"
	resp, err := http.Post(postURL, "application/json", b)

	if err != nil {
		return nil, fmt.Errorf("error posting to %s/api/embed:", err)
	}

	// Ensure the response body is closed after use to prevent resource leaks
	defer resp.Body.Close()

	// 3. Check for non-200 status
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embed status %d: %s", resp.StatusCode, respBody)
	}

	// 4. Decode the JSON response into our struct
	var result embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %w", err)
	}

	// 5. Sanity check — make sure we got at least one vector back
	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("embed returned no vectors")
	}

	// 6. Return the first (and only) embedding vector
	return result.Embeddings[0], nil
}
