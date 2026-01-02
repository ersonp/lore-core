// Package services contains domain business logic.
package services

import (
	"bufio"
	"context"
	"fmt"
	"io"
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
	llm               ports.LLMClient
	embedder          ports.Embedder
	vectorDB          ports.VectorDB
	entityTypeService *EntityTypeService
}

// NewExtractionService creates a new extraction service.
func NewExtractionService(llm ports.LLMClient, embedder ports.Embedder, vectorDB ports.VectorDB, entityTypeService *EntityTypeService) *ExtractionService {
	return &ExtractionService{
		llm:               llm,
		embedder:          embedder,
		vectorDB:          vectorDB,
		entityTypeService: entityTypeService,
	}
}

// extractFromChunks extracts facts from text chunks.
// Note: LLM calls in loop are intentional - LLMs have token limits, so text
// must be chunked and each chunk processed separately. Cannot be batched.
func (s *ExtractionService) extractFromChunks(ctx context.Context, text string, sourceFile string, validTypes []string) ([]entities.Fact, error) {
	chunks := ChunkText(text, DefaultChunkSize, DefaultChunkOverlap)

	var allFacts []entities.Fact
	for i, chunk := range chunks {
		//nolint:loopcall // LLM has token limits, must process chunks separately
		facts, err := s.llm.ExtractFacts(ctx, chunk, validTypes)
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

	return allFacts, nil
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
	// Get valid types for LLM prompt
	validTypes, err := s.entityTypeService.GetValidTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting valid types: %w", err)
	}

	allFacts, err := s.extractFromChunks(ctx, text, sourceFile, validTypes)
	if err != nil {
		return nil, err
	}

	if len(allFacts) == 0 {
		return &ExtractionResult{}, nil
	}

	texts := make([]string, len(allFacts))
	for i := range allFacts {
		texts[i] = factToText(&allFacts[i])
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

// streamChunker handles streaming chunking of text from an io.Reader.
type streamChunker struct {
	scanner       *bufio.Scanner
	currentChunk  strings.Builder
	lastParagraph strings.Builder
	inParagraph   bool
}

// newStreamChunker creates a chunker for the given reader.
func newStreamChunker(r io.Reader) *streamChunker {
	scanner := bufio.NewScanner(r)
	// Allow up to 1MB lines
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	return &streamChunker{scanner: scanner}
}

// addParagraphToChunk adds a completed paragraph to the current chunk.
// Returns true if chunk was processed (became full).
func (c *streamChunker) addParagraphToChunk(para string, processChunk func(string) error) (bool, error) {
	if len(para) == 0 {
		return false, nil
	}

	// Check if adding this paragraph would exceed chunk size
	if c.currentChunk.Len()+len(para)+2 > DefaultChunkSize && c.currentChunk.Len() > 0 {
		if err := processChunk(c.currentChunk.String()); err != nil {
			return false, err
		}

		// Start new chunk with overlap
		overlap := getOverlapText(c.currentChunk.String(), DefaultChunkOverlap)
		c.currentChunk.Reset()
		c.currentChunk.WriteString(overlap)

		// Add the paragraph that triggered the overflow to the new chunk
		if c.currentChunk.Len() > 0 {
			c.currentChunk.WriteString("\n\n")
		}
		c.currentChunk.WriteString(para)
		return true, nil
	}

	if c.currentChunk.Len() > 0 {
		c.currentChunk.WriteString("\n\n")
	}
	c.currentChunk.WriteString(para)
	return false, nil
}

// processLine handles a single line, accumulating paragraphs.
func (c *streamChunker) processLine(line string, processChunk func(string) error) error {
	if strings.TrimSpace(line) == "" {
		// Empty line marks paragraph boundary
		if c.inParagraph && c.lastParagraph.Len() > 0 {
			para := c.lastParagraph.String()
			if _, err := c.addParagraphToChunk(para, processChunk); err != nil {
				return err
			}
			c.lastParagraph.Reset()
			c.inParagraph = false
		}
		return nil
	}

	// Non-empty line: add to current paragraph
	if c.inParagraph {
		c.lastParagraph.WriteString("\n")
	}
	c.lastParagraph.WriteString(line)
	c.inParagraph = true
	return nil
}

// flush processes any remaining content.
func (c *streamChunker) flush(processChunk func(string) error) error {
	// Handle any remaining paragraph
	if c.lastParagraph.Len() > 0 {
		para := c.lastParagraph.String()
		if _, err := c.addParagraphToChunk(para, processChunk); err != nil {
			return err
		}
	}

	// Process the final chunk
	if c.currentChunk.Len() > 0 {
		return processChunk(c.currentChunk.String())
	}
	return nil
}

// ExtractFromReader extracts facts by streaming from an io.Reader.
// This reduces memory from O(file_size) to O(chunk_size) for large files.
func (s *ExtractionService) ExtractFromReader(ctx context.Context, r io.Reader, sourceFile string, opts ExtractionOptions) (*ExtractionResult, error) {
	// Get valid types for LLM prompt
	validTypes, err := s.entityTypeService.GetValidTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting valid types: %w", err)
	}

	chunker := newStreamChunker(r)
	var allFacts []entities.Fact

	// processChunk is called per chunk - LLM calls in loop are intentional
	// because LLMs have token limits and each chunk must be processed separately.
	processChunk := func(chunkText string) error {
		facts, err := s.llm.ExtractFacts(ctx, chunkText, validTypes)
		if err != nil {
			return fmt.Errorf("extracting facts: %w", err)
		}

		for i := range facts {
			facts[i].ID = uuid.New().String()
			facts[i].SourceFile = sourceFile
			facts[i].CreatedAt = time.Now()
			facts[i].UpdatedAt = time.Now()
		}

		allFacts = append(allFacts, facts...)
		return nil
	}

	for chunker.scanner.Scan() {
		if err := chunker.processLine(chunker.scanner.Text(), processChunk); err != nil {
			return nil, err
		}
	}

	if err := chunker.scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	if err := chunker.flush(processChunk); err != nil {
		return nil, err
	}

	if len(allFacts) == 0 {
		return &ExtractionResult{}, nil
	}

	return s.finalizeFacts(ctx, allFacts, opts)
}

// finalizeFacts generates embeddings, checks consistency, and saves facts.
func (s *ExtractionService) finalizeFacts(ctx context.Context, facts []entities.Fact, opts ExtractionOptions) (*ExtractionResult, error) {
	texts := make([]string, len(facts))
	for i := range facts {
		texts[i] = factToText(&facts[i])
	}

	embeddings, err := s.embedder.EmbedBatch(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("generating embeddings: %w", err)
	}

	for i := range facts {
		facts[i].Embedding = embeddings[i]
	}

	result := &ExtractionResult{
		Facts: facts,
	}

	if opts.CheckConsistency {
		issues, err := s.checkConsistency(ctx, facts)
		if err != nil {
			return nil, fmt.Errorf("checking consistency: %w", err)
		}
		result.Issues = issues
	}

	if !opts.CheckOnly {
		if err := s.vectorDB.SaveBatch(ctx, facts); err != nil {
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

	for i := range newFacts {
		similarFacts, err := s.vectorDB.SearchByType(ctx, newFacts[i].Embedding, newFacts[i].Type, 5)
		if err != nil {
			return nil, fmt.Errorf("searching similar facts: %w", err)
		}

		// Deduplicate similar facts
		for j := range similarFacts {
			if !seenIDs[similarFacts[j].ID] {
				seenIDs[similarFacts[j].ID] = true
				allSimilarFacts = append(allSimilarFacts, similarFacts[j])
			}
		}
	}

	if len(allSimilarFacts) == 0 {
		return nil, nil
	}

	// Step 2: Single batched LLM call for all facts
	issues, err := s.llm.CheckConsistency(ctx, newFacts, allSimilarFacts)
	if err != nil {
		return nil, fmt.Errorf("LLM consistency check: %w", err)
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
func factToText(fact *entities.Fact) string {
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
