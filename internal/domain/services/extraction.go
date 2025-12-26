// Package services contains domain business logic.
package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/ports"
)

// ExtractionOptions controls extraction behavior.
type ExtractionOptions struct {
	CheckConsistency bool // Check for contradictions with existing facts
	CheckOnly        bool // Only check, don't save facts
}

// ExtractionResult contains the result of extraction.
type ExtractionResult struct {
	Facts  []entities.Fact
	Issues []ports.ConsistencyIssue
}

const (
	// DefaultChunkSize is the default size for text chunks.
	DefaultChunkSize = 2000
	// DefaultChunkOverlap is the default overlap between chunks.
	DefaultChunkOverlap = 200
)

// ExtractionService handles fact extraction from text.
type ExtractionService struct {
	llm      ports.LLMClient
	embedder ports.Embedder
	vectorDB ports.VectorDB
}

// NewExtractionService creates a new extraction service.
func NewExtractionService(llm ports.LLMClient, embedder ports.Embedder, vectorDB ports.VectorDB) *ExtractionService {
	return &ExtractionService{
		llm:      llm,
		embedder: embedder,
		vectorDB: vectorDB,
	}
}

// ExtractAndStore extracts facts from text, generates embeddings, and stores them.
func (s *ExtractionService) ExtractAndStore(ctx context.Context, text string, sourceFile string) ([]entities.Fact, error) {
	result, err := s.ExtractAndStoreWithOptions(ctx, text, sourceFile, ExtractionOptions{})
	if err != nil {
		return nil, err
	}
	return result.Facts, nil
}

// ExtractAndStoreWithOptions extracts facts with consistency checking options.
func (s *ExtractionService) ExtractAndStoreWithOptions(ctx context.Context, text string, sourceFile string, opts ExtractionOptions) (*ExtractionResult, error) {
	chunks := ChunkText(text, DefaultChunkSize, DefaultChunkOverlap)

	var allFacts []entities.Fact
	for i, chunk := range chunks {
		facts, err := s.llm.ExtractFacts(ctx, chunk)
		if err != nil {
			return nil, fmt.Errorf("extracting facts from chunk %d: %w", i, err)
		}

		for j := range facts {
			facts[j].ID = uuid.New().String()
			facts[j].SourceFile = sourceFile
			facts[j].CreatedAt = time.Now()
			facts[j].UpdatedAt = time.Now()
		}

		allFacts = append(allFacts, facts...)
	}

	if len(allFacts) == 0 {
		return &ExtractionResult{}, nil
	}

	texts := make([]string, len(allFacts))
	for i, fact := range allFacts {
		texts[i] = factToText(fact)
	}

	embeddings, err := s.embedder.EmbedBatch(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("generating embeddings: %w", err)
	}

	for i := range allFacts {
		allFacts[i].Embedding = embeddings[i]
	}

	result := &ExtractionResult{
		Facts: allFacts,
	}

	// Check consistency if requested
	if opts.CheckConsistency {
		issues, err := s.checkConsistency(ctx, allFacts)
		if err != nil {
			return nil, fmt.Errorf("checking consistency: %w", err)
		}
		result.Issues = issues
	}

	// Save facts unless check-only mode
	if !opts.CheckOnly {
		if err := s.vectorDB.SaveBatch(ctx, allFacts); err != nil {
			return nil, fmt.Errorf("saving facts: %w", err)
		}
	}

	return result, nil
}

// checkConsistency checks new facts against existing facts for contradictions.
// Uses batched LLM call for efficiency - collects all similar facts first,
// then makes a single LLM call instead of one per fact.
func (s *ExtractionService) checkConsistency(ctx context.Context, newFacts []entities.Fact) ([]ports.ConsistencyIssue, error) {
	// Step 1: Collect all similar facts from DB (fast calls)
	var allSimilarFacts []entities.Fact
	seenIDs := make(map[string]bool)

	for _, fact := range newFacts {
		similarFacts, err := s.vectorDB.SearchByType(ctx, fact.Embedding, fact.Type, 5)
		if err != nil {
			return nil, fmt.Errorf("searching similar facts: %w", err)
		}

		// Deduplicate similar facts
		for _, sf := range similarFacts {
			if !seenIDs[sf.ID] {
				seenIDs[sf.ID] = true
				allSimilarFacts = append(allSimilarFacts, sf)
			}
		}
	}

	if len(allSimilarFacts) == 0 {
		return nil, nil
	}

	// Step 2: Single batched LLM call for all facts
	issues, err := s.llm.CheckConsistency(ctx, newFacts, allSimilarFacts)
	if err != nil {
		// Log warning but don't fail - consistency check is advisory
		return nil, nil
	}

	return issues, nil
}

// ChunkText splits text into chunks with overlap.
func ChunkText(text string, chunkSize int, overlap int) []string {
	if len(text) <= chunkSize {
		return []string{text}
	}

	var chunks []string
	paragraphs := strings.Split(text, "\n\n")

	var currentChunk strings.Builder
	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		if currentChunk.Len()+len(para)+2 > chunkSize && currentChunk.Len() > 0 {
			chunks = append(chunks, currentChunk.String())

			overlapText := getOverlapText(currentChunk.String(), overlap)
			currentChunk.Reset()
			currentChunk.WriteString(overlapText)
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(para)
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	if len(chunks) == 0 && len(text) > 0 {
		chunks = append(chunks, text)
	}

	return chunks
}

// getOverlapText returns the last n characters of text for overlap.
func getOverlapText(text string, n int) string {
	if len(text) <= n {
		return text
	}
	return text[len(text)-n:]
}

// factToText converts a fact to searchable text for embedding.
func factToText(fact entities.Fact) string {
	parts := []string{
		fact.Subject,
		fact.Predicate,
		fact.Object,
	}
	if fact.Context != "" {
		parts = append(parts, fact.Context)
	}
	return strings.Join(parts, " ")
}
