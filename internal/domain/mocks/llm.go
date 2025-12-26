// Package mocks provides mock implementations for testing.
package mocks

import (
	"context"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/ports"
)

// LLMClient is a mock implementation of ports.LLMClient.
type LLMClient struct {
	// ExtractFacts return values
	Facts      []entities.Fact
	ExtractErr error

	// CheckConsistency return values
	Issues         []ports.ConsistencyIssue
	ConsistencyErr error
}

// ExtractFacts returns the configured facts or error.
func (m *LLMClient) ExtractFacts(ctx context.Context, text string) ([]entities.Fact, error) {
	if m.ExtractErr != nil {
		return nil, m.ExtractErr
	}
	return m.Facts, nil
}

// CheckConsistency returns the configured issues or error.
func (m *LLMClient) CheckConsistency(ctx context.Context, newFacts []entities.Fact, existingFacts []entities.Fact) ([]ports.ConsistencyIssue, error) {
	if m.ConsistencyErr != nil {
		return nil, m.ConsistencyErr
	}
	return m.Issues, nil
}
