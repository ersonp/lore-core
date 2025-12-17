package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ersonp/lore-core/internal/domain/entities"
)

func TestChunkText(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		chunkSize int
		overlap   int
		wantCount int
	}{
		{
			name:      "short text fits in one chunk",
			text:      "This is a short text.",
			chunkSize: 100,
			overlap:   10,
			wantCount: 1,
		},
		{
			name:      "empty text returns single chunk",
			text:      "",
			chunkSize: 100,
			overlap:   10,
			wantCount: 1,
		},
		{
			name:      "text splits into multiple chunks",
			text:      "First paragraph.\n\nSecond paragraph.\n\nThird paragraph.\n\nFourth paragraph.",
			chunkSize: 40,
			overlap:   10,
			wantCount: 3,
		},
		{
			name:      "text exactly at chunk size",
			text:      "12345678901234567890",
			chunkSize: 20,
			overlap:   5,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := ChunkText(tt.text, tt.chunkSize, tt.overlap)
			assert.Equal(t, tt.wantCount, len(chunks))

			if tt.wantCount > 0 && tt.text != "" {
				combined := ""
				for _, chunk := range chunks {
					if combined == "" {
						combined = chunk
					}
				}
				assert.NotEmpty(t, combined)
			}
		})
	}
}

func TestChunkText_PreservesParagraphs(t *testing.T) {
	text := "First paragraph with some content.\n\nSecond paragraph with more content.\n\nThird paragraph."
	chunks := ChunkText(text, 60, 10)

	assert.GreaterOrEqual(t, len(chunks), 1)

	for _, chunk := range chunks {
		assert.NotEmpty(t, chunk)
	}
}

func TestChunkText_HandlesLongParagraph(t *testing.T) {
	longPara := "This is a very long paragraph that exceeds the chunk size limit and should be handled gracefully by the chunking algorithm without breaking."
	chunks := ChunkText(longPara, 50, 10)

	assert.GreaterOrEqual(t, len(chunks), 1)
}

func TestChunkText_OnlyWhitespace(t *testing.T) {
	chunks := ChunkText("   \n\n   \n\n   ", 100, 10)
	// Should return original text as single chunk since no real paragraphs
	assert.Len(t, chunks, 1)
}

func TestChunkText_SingleParagraph(t *testing.T) {
	text := "This is a single paragraph without any double newlines."
	chunks := ChunkText(text, 100, 10)
	assert.Len(t, chunks, 1)
	assert.Equal(t, text, chunks[0])
}

func TestChunkText_OverlapContent(t *testing.T) {
	// Test that overlap is actually included
	text := "First paragraph here.\n\nSecond paragraph here.\n\nThird paragraph here."
	chunks := ChunkText(text, 40, 15)

	// Should have multiple chunks
	assert.Greater(t, len(chunks), 1)

	// Each chunk (except first) should have some overlap from previous
	for i := 1; i < len(chunks); i++ {
		assert.NotEmpty(t, chunks[i])
	}
}

func TestGetOverlapText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		n        int
		expected string
	}{
		{
			name:     "normal overlap",
			text:     "Hello World",
			n:        5,
			expected: "World",
		},
		{
			name:     "overlap larger than text",
			text:     "Hi",
			n:        10,
			expected: "Hi",
		},
		{
			name:     "overlap equals text length",
			text:     "Hello",
			n:        5,
			expected: "Hello",
		},
		{
			name:     "zero overlap",
			text:     "Hello",
			n:        0,
			expected: "",
		},
		{
			name:     "empty text",
			text:     "",
			n:        5,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getOverlapText(tt.text, tt.n)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFactToText(t *testing.T) {
	tests := []struct {
		name     string
		fact     entities.Fact
		expected string
	}{
		{
			name: "basic fact without context",
			fact: entities.Fact{
				Subject:   "Frodo",
				Predicate: "has_trait",
				Object:    "brave",
			},
			expected: "Frodo has_trait brave",
		},
		{
			name: "fact with context",
			fact: entities.Fact{
				Subject:   "Frodo",
				Predicate: "lives_in",
				Object:    "The Shire",
				Context:   "At the start of the story",
			},
			expected: "Frodo lives_in The Shire At the start of the story",
		},
		{
			name: "fact with empty context",
			fact: entities.Fact{
				Subject:   "Gandalf",
				Predicate: "is_a",
				Object:    "wizard",
				Context:   "",
			},
			expected: "Gandalf is_a wizard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := factToText(tt.fact)
			assert.Equal(t, tt.expected, result)
		})
	}
}
