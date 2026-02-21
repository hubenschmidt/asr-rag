package vectordb

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/qdrant/go-client/qdrant"
)

const collectionName = "go_terms"

// Client wraps the Qdrant gRPC client.
type Client struct {
	conn *qdrant.Client
}

// New connects to Qdrant at "host:port".
func New(addr string) (*Client, error) {
	host, port, err := parseAddr(addr)
	if err != nil {
		return nil, fmt.Errorf("could not parse host and post address: %w", err)
	}
	conn, err := qdrant.NewClient(&qdrant.Config{Host: host, Port: port})
	if err != nil {
		return nil, fmt.Errorf("could not connect to qdrant: %w", err)
	}

	return &Client{conn: conn}, nil
}

// Close shuts down the gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// EnsureCollection creates the "go_terms" collection if it doesn't already exist.
// Makes seed idempotent — safe to run multiple times.
// Uses 768 dimensions (nomic-embed-text) with cosine similarity.
func (c *Client) EnsureCollection(ctx context.Context) error {
	collections, err := c.conn.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("could not list qdrant collections: %w", err)
	}

	for _, collection := range collections {
		if collection == collectionName {
			return nil
		}
	}

	if err := c.conn.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     768,
			Distance: qdrant.Distance_Cosine,
		}),
	}); err != nil {
		return fmt.Errorf("error creating collection: %w", err)
	}

	return nil
}

// Upsert stores a single point (vector + metadata) in Qdrant.
// Blocks until the write is confirmed.
func (c *Client) Upsert(ctx context.Context, id uint64, vec []float32, term string, definition string) error {
	wait := true
	_, err := c.conn.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Wait:           &wait,
		Points: []*qdrant.PointStruct{
			{
				Id:      qdrant.NewIDNum(id),
				Vectors: qdrant.NewVectorsDense(vec),
				Payload: qdrant.NewValueMap(map[string]any{
					"term":       term,
					"definition": definition,
				}),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("upsert point %d: %w", id, err)
	}
	return nil
}

// SearchResult holds a single search hit with its score and payload.
type SearchResult struct {
	Term       string
	Definition string
	Score      float32
}

// Search finds the top-k most similar vectors to the query vector.
// Returns scored results with term, definition, and similarity score.
func (c *Client) Search(ctx context.Context, vec []float32, limit uint64) ([]SearchResult, error) {
	points, err := c.conn.Query(ctx, &qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          qdrant.NewQueryDense(vec),
		Limit:          &limit,
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	var results []SearchResult
	for _, el := range points { // _ is index
		results = append(results, SearchResult{
			Term:       el.Payload["term"].GetStringValue(),
			Definition: el.Payload["definition"].GetStringValue(),
			Score:      el.Score,
		})
	}
	return results, nil
}

// parseAddr splits "host:port" into separate values.
func parseAddr(addr string) (string, int, error) {
	hostStr, portStr, ok := strings.Cut(addr, ":")
	if !ok {
		return "", 0, fmt.Errorf("invalid addr %s", addr)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("invalid port %s", portStr)
	}
	return hostStr, port, nil
}
