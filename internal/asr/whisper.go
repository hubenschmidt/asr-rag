package asr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// Client talks to whisper-server's /inference endpoint.
type Client struct {
	url string
}

// New creates a whisper ASR client.
func New(url string) *Client {
	return &Client{url: url}
}

// whisperResponse is the JSON returned by whisper-server: {"text": "..."}
type whisperResponse struct {
	Text string `json:"text"`
}

// Transcribe reads a WAV file and sends it to whisper-server.
// Returns the raw transcript text.
//
// This is different from the embedder — whisper expects a multipart file upload
// (like a browser form with a file input), not a JSON body.
//
// Step 1: Read the WAV file from disk
// Step 2: Build a multipart form body
// Step 3: POST to whisper-server
// Step 4: Check for non-200 status
// Step 5: Decode JSON response into whisperResponse
func (c *Client) Transcribe(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read WAV file: %w", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return "", fmt.Errorf("failed to write WAV file: %w", err)
	}
	part.Write(data)
	writer.Close()

	resp, err := http.Post(c.url+"/inference", writer.FormDataContentType(), &body)
	if err != nil {
		return "", fmt.Errorf("could not post to whisper /inference endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("whisper status %d: %s", resp.StatusCode, respBody)
	}

	var result whisperResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding whisper response: %w", err)
	}

	return result.Text, nil
}
