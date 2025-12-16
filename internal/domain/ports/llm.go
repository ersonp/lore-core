// Package ports defines interfaces for external service communication.
package ports

import (
	"context"

	"github.com/ersonp/lore-core/internal/domain/entities"
)

// LLMClient defines the interface for LLM operations.
type LLMClient interface {
	// ExtractFacts extracts facts from the given text.
	ExtractFacts(ctx context.Context, text string) ([]entities.Fact, error)

	// CheckConsistency checks if new facts are consistent with existing facts.
	CheckConsistency(ctx context.Context, newFacts []entities.Fact, existingFacts []entities.Fact) ([]ConsistencyIssue, error)
}

// ConsistencyIssue represents a detected inconsistency between facts.
type ConsistencyIssue struct {
	NewFact      entities.Fact `json:"new_fact"`
	ExistingFact entities.Fact `json:"existing_fact"`
	Description  string        `json:"description"`
	Severity     string        `json:"severity"`
}
