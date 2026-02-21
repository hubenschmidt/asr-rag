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
func (c *Client) Transcribe(path string) (string, error) {

	// Read the WAV file from disk
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read WAV file: %w", err)
	}

	// build form body
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return "", fmt.Errorf("failed to write WAV file: %w", err)
	}
	part.Write(data)
	writer.Close()

	// post to whisper server
	resp, err := http.Post(c.url+"/inference", writer.FormDataContentType(), &body)
	if err != nil {
		return "", fmt.Errorf("could not post to whisper /inference endpoint: %w", err)
	}
	defer resp.Body.Close()

	// check for non-200 status
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("whisper status %d: %s", resp.StatusCode, respBody)
	}

	// decode JSON response
	var result whisperResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding whisper response: %w", err)
	}

	return result.Text, nil
}
