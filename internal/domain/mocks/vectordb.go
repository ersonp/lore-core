package mocks

import (
	"context"

	"github.com/ersonp/lore-core/internal/domain/entities"
)

// VectorDB is a mock implementation of ports.VectorDB.
type VectorDB struct {
	Facts []entities.Fact
	Err   error
}

// Save stores a single fact.
func (m *VectorDB) Save(ctx context.Context, fact entities.Fact) error {
	return m.Err
}

// SaveBatch stores multiple facts.
func (m *VectorDB) SaveBatch(ctx context.Context, facts []entities.Fact) error {
	return m.Err
}

// FindByID retrieves a fact by ID.
func (m *VectorDB) FindByID(ctx context.Context, id string) (entities.Fact, error) {
	if m.Err != nil {
		return entities.Fact{}, m.Err
	}
	for _, f := range m.Facts {
		if f.ID == id {
			return f, nil
		}
	}
	return entities.Fact{}, nil
}

// Search finds facts by embedding similarity.
func (m *VectorDB) Search(ctx context.Context, embedding []float32, limit int) ([]entities.Fact, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	if limit > len(m.Facts) {
		return m.Facts, nil
	}
	return m.Facts[:limit], nil
}

// SearchByType finds facts by embedding and type.
func (m *VectorDB) SearchByType(ctx context.Context, embedding []float32, factType entities.FactType, limit int) ([]entities.Fact, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	var filtered []entities.Fact
	for _, f := range m.Facts {
		if f.Type == factType {
			filtered = append(filtered, f)
		}
	}
	if limit > len(filtered) {
		return filtered, nil
	}
	return filtered[:limit], nil
}

// Delete removes a fact by ID.
func (m *VectorDB) Delete(ctx context.Context, id string) error {
	return m.Err
}

// List returns facts with pagination.
func (m *VectorDB) List(ctx context.Context, limit int, offset uint64) ([]entities.Fact, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Facts, nil
}

// ListByType returns facts filtered by type.
func (m *VectorDB) ListByType(ctx context.Context, factType entities.FactType, limit int) ([]entities.Fact, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	var filtered []entities.Fact
	for _, f := range m.Facts {
		if f.Type == factType {
			filtered = append(filtered, f)
		}
	}
	return filtered, nil
}

// ListBySource returns facts filtered by source file.
func (m *VectorDB) ListBySource(ctx context.Context, sourceFile string, limit int) ([]entities.Fact, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	var filtered []entities.Fact
	for _, f := range m.Facts {
		if f.SourceFile == sourceFile {
			filtered = append(filtered, f)
		}
	}
	return filtered, nil
}

// DeleteBySource removes all facts from a source file.
func (m *VectorDB) DeleteBySource(ctx context.Context, sourceFile string) error {
	return m.Err
}

// DeleteAll removes all facts.
func (m *VectorDB) DeleteAll(ctx context.Context) error {
	return m.Err
}

// Count returns the total number of facts.
func (m *VectorDB) Count(ctx context.Context) (uint64, error) {
	if m.Err != nil {
		return 0, m.Err
	}
	return uint64(len(m.Facts)), nil
}

// Close closes the connection.
func (m *VectorDB) Close() error {
	return nil
}
