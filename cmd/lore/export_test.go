package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ersonp/lore-core/internal/domain/entities"
)

func TestFormatJSON(t *testing.T) {
	facts := []entities.Fact{
		{
			ID:         "test-id-1",
			Type:       entities.FactTypeCharacter,
			Subject:    "Frodo",
			Predicate:  "has_trait",
			Object:     "brave",
			Context:    "The Shire",
			SourceFile: "chapter1.txt",
			Confidence: 0.95,
		},
	}

	var buf bytes.Buffer
	err := formatJSON(&buf, facts)
	require.NoError(t, err)

	result := buf.String()

	// Verify it's valid JSON
	var parsed []map[string]interface{}
	err = json.Unmarshal([]byte(result), &parsed)
	require.NoError(t, err)

	assert.Len(t, parsed, 1)
	assert.Equal(t, "test-id-1", parsed[0]["id"])
	assert.Equal(t, "character", parsed[0]["type"])
	assert.Equal(t, "Frodo", parsed[0]["subject"])
	assert.Equal(t, "has_trait", parsed[0]["predicate"])
	assert.Equal(t, "brave", parsed[0]["object"])
	assert.Equal(t, "The Shire", parsed[0]["context"])
	assert.Equal(t, "chapter1.txt", parsed[0]["source_file"])
	assert.Equal(t, 0.95, parsed[0]["confidence"])
}

func TestFormatJSON_EmptyFacts(t *testing.T) {
	facts := []entities.Fact{}

	var buf bytes.Buffer
	err := formatJSON(&buf, facts)
	require.NoError(t, err)
	assert.Equal(t, "[]\n", buf.String())
}

func TestFormatCSV(t *testing.T) {
	facts := []entities.Fact{
		{
			ID:         "test-id-1",
			Type:       entities.FactTypeCharacter,
			Subject:    "Frodo",
			Predicate:  "has_trait",
			Object:     "brave",
			Context:    "",
			SourceFile: "chapter1.txt",
			Confidence: 0.95,
		},
	}

	var buf bytes.Buffer
	err := formatCSV(&buf, facts)
	require.NoError(t, err)

	result := buf.String()
	lines := strings.Split(strings.TrimSpace(result), "\n")
	require.Len(t, lines, 2)

	// Check header
	assert.Equal(t, "id,type,subject,predicate,object,context,source_file,confidence", lines[0])

	// Check data row
	assert.Contains(t, lines[1], "test-id-1")
	assert.Contains(t, lines[1], "character")
	assert.Contains(t, lines[1], "Frodo")
	assert.Contains(t, lines[1], "0.95")
}

func TestFormatCSV_SpecialCharacters(t *testing.T) {
	facts := []entities.Fact{
		{
			ID:        "test-id-1",
			Type:      entities.FactTypeCharacter,
			Subject:   "Name, with comma",
			Predicate: "has",
			Object:    "value \"quoted\"",
		},
	}

	var buf bytes.Buffer
	err := formatCSV(&buf, facts)
	require.NoError(t, err)

	result := buf.String()

	// CSV should properly escape commas and quotes
	assert.Contains(t, result, "\"Name, with comma\"")
}

func TestFormatMarkdown(t *testing.T) {
	facts := []entities.Fact{
		{
			ID:         "test-id-1",
			Type:       entities.FactTypeCharacter,
			Subject:    "Frodo",
			Predicate:  "has_trait",
			Object:     "brave",
			SourceFile: "chapter1.txt",
		},
	}

	var buf bytes.Buffer
	err := formatMarkdown(&buf, facts)
	require.NoError(t, err)

	result := buf.String()
	assert.Contains(t, result, "# Exported Facts")
	assert.Contains(t, result, "Total: 1 facts")
	assert.Contains(t, result, "| Type | Subject | Predicate | Object | Source |")
	assert.Contains(t, result, "| character | Frodo | has_trait | brave | chapter1.txt |")
}

func TestFormatMarkdown_LongSourceTruncated(t *testing.T) {
	longSource := "/very/long/path/to/some/deeply/nested/file/name.txt"
	facts := []entities.Fact{
		{
			ID:         "test-id-1",
			Type:       entities.FactTypeCharacter,
			Subject:    "Test",
			Predicate:  "is",
			Object:     "test",
			SourceFile: longSource,
		},
	}

	var buf bytes.Buffer
	err := formatMarkdown(&buf, facts)
	require.NoError(t, err)

	result := buf.String()
	// Source should be truncated with ...
	assert.Contains(t, result, "...")
	assert.NotContains(t, result, longSource) // Full path should not appear
}

func TestEscapeMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "pipe escaped",
			input:    "value|with|pipes",
			expected: "value\\|with\\|pipes",
		},
		{
			name:     "newline replaced",
			input:    "line1\nline2",
			expected: "line1 line2",
		},
		{
			name:     "no change needed",
			input:    "simple text",
			expected: "simple text",
		},
		{
			name:     "combined",
			input:    "pipe|and\nnewline",
			expected: "pipe\\|and newline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeMarkdown(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContains(t *testing.T) {
	slice := []string{"json", "csv", "markdown"}

	assert.True(t, contains(slice, "json"))
	assert.True(t, contains(slice, "csv"))
	assert.True(t, contains(slice, "markdown"))
	assert.False(t, contains(slice, "xml"))
	assert.False(t, contains(slice, ""))
	assert.False(t, contains(slice, "JSON")) // case sensitive
}
