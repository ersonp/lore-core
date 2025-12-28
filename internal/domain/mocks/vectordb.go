package mocks

import (
	"context"
	"fmt"

	"github.com/ersonp/lore-core/internal/domain/entities"
)

// VectorDB is a mock implementation of ports.VectorDB.
type VectorDB struct {
	Facts []entities.Fact
	Err   error

	// Collection errors (separate from Err for fine-grained control)
	EnsureCollectionErr error
	DeleteCollectionErr error
	FindByIDErr         error // Error to return when fact is not found

	// Call tracking
	SaveBatchCallCount        int
	SaveBatchLastFacts        []entities.Fact
	EnsureCollectionCallCount int
	DeleteCollectionCallCount int
	FindByIDCallCount         int
}

// EnsureCollection creates the collection if it doesn't exist.
func (m *VectorDB) EnsureCollection(ctx context.Context, vectorSize uint64) error {
	m.EnsureCollectionCallCount++
	return m.EnsureCollectionErr
}

// DeleteCollection removes the collection and all its data.
func (m *VectorDB) DeleteCollection(ctx context.Context) error {
	m.DeleteCollectionCallCount++
	return m.DeleteCollectionErr
}

// Save stores a single fact.
func (m *VectorDB) Save(ctx context.Context, fact entities.Fact) error {
	return m.Err
}

// SaveBatch stores multiple facts.
func (m *VectorDB) SaveBatch(ctx context.Context, facts []entities.Fact) error {
	m.SaveBatchCallCount++
	m.SaveBatchLastFacts = facts
	return m.Err
}

// FindByID retrieves a fact by ID.
func (m *VectorDB) FindByID(ctx context.Context, id string) (entities.Fact, error) {
	m.FindByIDCallCount++
	if m.Err != nil {
		return entities.Fact{}, m.Err
	}
	for i := range m.Facts {
		if m.Facts[i].ID == id {
			return m.Facts[i], nil
		}
	}
	// Return FindByIDErr when fact not found (mimics real behavior)
	if m.FindByIDErr != nil {
		return entities.Fact{}, m.FindByIDErr
	}
	return entities.Fact{}, fmt.Errorf("fact not found: %s", id)
}

// ExistsByIDs checks which IDs exist in the mock database.
func (m *VectorDB) ExistsByIDs(ctx context.Context, ids []string) (map[string]bool, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	exists := make(map[string]bool, len(ids))
	for _, id := range ids {
		for i := range m.Facts {
			if m.Facts[i].ID == id {
				exists[id] = true
				break
			}
		}
	}
	return exists, nil
}

// FindByIDs retrieves multiple facts by their IDs.
func (m *VectorDB) FindByIDs(ctx context.Context, ids []string) ([]entities.Fact, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	var facts []entities.Fact
	for _, id := range ids {
		for i := range m.Facts {
			if m.Facts[i].ID == id {
				facts = append(facts, m.Facts[i])
				break
			}
		}
	}
	return facts, nil
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
	for i := range m.Facts {
		if m.Facts[i].Type == factType {
			filtered = append(filtered, m.Facts[i])
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
	for i := range m.Facts {
		if m.Facts[i].Type == factType {
			filtered = append(filtered, m.Facts[i])
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
	for i := range m.Facts {
		if m.Facts[i].SourceFile == sourceFile {
			filtered = append(filtered, m.Facts[i])
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
