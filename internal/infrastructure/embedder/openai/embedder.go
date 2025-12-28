// Package openai provides an Embedder implementation using OpenAI.
package openai

import (
	"context"
	"errors"
	"fmt"

	"github.com/sashabaranov/go-openai"

	"github.com/ersonp/lore-core/internal/infrastructure/config"
)

// VectorSize is the dimension of text-embedding-3-small vectors.
const VectorSize = 1536

// Embedder implements the Embedder interface using OpenAI.
type Embedder struct {
	client *openai.Client
	model  openai.EmbeddingModel
}

// NewEmbedder creates a new OpenAI embedder.
func NewEmbedder(cfg config.EmbedderConfig) (*Embedder, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("OpenAI API key is required")
	}

	client := openai.NewClient(cfg.APIKey)

	model := openai.SmallEmbedding3
	if cfg.Model != "" {
		model = openai.EmbeddingModel(cfg.Model)
	}

	return &Embedder{
		client: client,
		model:  model,
	}, nil
}

// Embed generates a vector embedding for the given text.
func (e *Embedder) Embed(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := e.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	if len(embeddings) == 0 {
		return nil, errors.New("no embeddings returned")
	}

	return embeddings[0], nil
}

// EmbedBatch generates vector embeddings for multiple texts.
func (e *Embedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	resp, err := e.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model: e.model,
		Input: texts,
	})
	if err != nil {
		return nil, fmt.Errorf("creating embeddings: %w", err)
	}

	embeddings := make([][]float32, len(resp.Data))
	for i, data := range resp.Data {
		embeddings[i] = data.Embedding
	}

	return embeddings, nil
}
