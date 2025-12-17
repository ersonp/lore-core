package services

import (
	"context"
	"fmt"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/ports"
)

// DefaultSearchLimit is the default number of results to return.
const DefaultSearchLimit = 10

// QueryService handles fact querying and search.
type QueryService struct {
	embedder ports.Embedder
	vectorDB ports.VectorDB
}

// NewQueryService creates a new query service.
func NewQueryService(embedder ports.Embedder, vectorDB ports.VectorDB) *QueryService {
	return &QueryService{
		embedder: embedder,
		vectorDB: vectorDB,
	}
}

// Search finds facts semantically similar to the query.
func (s *QueryService) Search(ctx context.Context, query string, limit int) ([]entities.Fact, error) {
	if limit <= 0 {
		limit = DefaultSearchLimit
	}

	embedding, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("generating query embedding: %w", err)
	}

	facts, err := s.vectorDB.Search(ctx, embedding, limit)
	if err != nil {
		return nil, fmt.Errorf("searching facts: %w", err)
	}

	return facts, nil
}

// SearchByType finds facts filtered by type.
func (s *QueryService) SearchByType(ctx context.Context, query string, factType entities.FactType, limit int) ([]entities.Fact, error) {
	if limit <= 0 {
		limit = DefaultSearchLimit
	}

	embedding, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("generating query embedding: %w", err)
	}

	facts, err := s.vectorDB.SearchByType(ctx, embedding, factType, limit)
	if err != nil {
		return nil, fmt.Errorf("searching facts by type: %w", err)
	}

	return facts, nil
}
