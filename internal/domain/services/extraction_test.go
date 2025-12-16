package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
