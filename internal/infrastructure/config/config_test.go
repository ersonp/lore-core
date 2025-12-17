package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeWorldName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple lowercase",
			input:    "myworld",
			expected: "myworld",
		},
		{
			name:     "uppercase converted",
			input:    "MyWorld",
			expected: "myworld",
		},
		{
			name:     "spaces to underscores",
			input:    "my world",
			expected: "my_world",
		},
		{
			name:     "hyphens to underscores",
			input:    "my-world",
			expected: "my_world",
		},
		{
			name:     "special characters removed",
			input:    "my@world!",
			expected: "myworld",
		},
		{
			name:     "consecutive underscores collapsed",
			input:    "my--world",
			expected: "my_world",
		},
		{
			name:     "leading trailing underscores trimmed",
			input:    "-my-world-",
			expected: "my_world",
		},
		{
			name:     "empty string returns default",
			input:    "",
			expected: "default",
		},
		{
			name:     "only special chars returns default",
			input:    "!!!",
			expected: "default",
		},
		{
			name:     "numbers preserved",
			input:    "world123",
			expected: "world123",
		},
		{
			name:     "complex mixed input",
			input:    "Iron-Throne (Book 1)",
			expected: "iron_throne_book_1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeWorldName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateCollectionName(t *testing.T) {
	tests := []struct {
		name      string
		worldName string
		expected  string
	}{
		{
			name:      "simple world",
			worldName: "myworld",
			expected:  "lore_myworld",
		},
		{
			name:      "world with spaces",
			worldName: "my world",
			expected:  "lore_my_world",
		},
		{
			name:      "world with special chars",
			worldName: "Iron-Throne!",
			expected:  "lore_iron_throne",
		},
		{
			name:      "empty world uses default",
			worldName: "",
			expected:  "lore_default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateCollectionName(tt.worldName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefault(t *testing.T) {
	cfg := Default()

	assert.Equal(t, "openai", cfg.LLM.Provider)
	assert.Equal(t, "gpt-4o-mini", cfg.LLM.Model)
	assert.Equal(t, "openai", cfg.Embedder.Provider)
	assert.Equal(t, "text-embedding-3-small", cfg.Embedder.Model)
	assert.Equal(t, "localhost", cfg.Qdrant.Host)
	assert.Equal(t, 6334, cfg.Qdrant.Port)
	assert.NotNil(t, cfg.Worlds)
	assert.Empty(t, cfg.Worlds)
}

func TestConfigDir(t *testing.T) {
	result := ConfigDir("/home/user/project")
	assert.Equal(t, "/home/user/project/.lore", result)
}

func TestConfigFilePath(t *testing.T) {
	result := ConfigFilePath("/home/user/project")
	assert.Equal(t, "/home/user/project/.lore/config.yaml", result)
}
