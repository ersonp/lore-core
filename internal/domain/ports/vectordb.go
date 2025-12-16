package ports

import (
	"context"

	"github.com/ersonp/lore-core/internal/domain/entities"
)

// VectorDB defines the interface for vector database operations.
type VectorDB interface {
	// Save stores a fact with its embedding.
	Save(ctx context.Context, fact entities.Fact) error

	// SaveBatch stores multiple facts.
	SaveBatch(ctx context.Context, facts []entities.Fact) error

	// FindByID retrieves a fact by its ID.
	FindByID(ctx context.Context, id string) (entities.Fact, error)

	// Search performs a semantic search and returns similar facts.
	Search(ctx context.Context, embedding []float32, limit int) ([]entities.Fact, error)

	// SearchByType performs a semantic search filtered by fact type.
	SearchByType(ctx context.Context, embedding []float32, factType entities.FactType, limit int) ([]entities.Fact, error)

	// Delete removes a fact by its ID.
	Delete(ctx context.Context, id string) error
}
