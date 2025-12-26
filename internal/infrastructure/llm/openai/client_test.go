package openai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.LLMConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			cfg: config.LLMConfig{
				APIKey: "test-key",
			},
			wantErr: false,
		},
		{
			name: "valid config with model",
			cfg: config.LLMConfig{
				APIKey: "test-key",
				Model:  "gpt-4",
			},
			wantErr: false,
		},
		{
			name:    "missing API key",
			cfg:     config.LLMConfig{},
			wantErr: true,
			errMsg:  "API key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.cfg)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestCleanJSONResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain JSON",
			input:    `[{"type": "character"}]`,
			expected: `[{"type": "character"}]`,
		},
		{
			name:     "JSON with json code block",
			input:    "```json\n[{\"type\": \"character\"}]\n```",
			expected: `[{"type": "character"}]`,
		},
		{
			name:     "JSON with plain code block",
			input:    "```\n[{\"type\": \"character\"}]\n```",
			expected: `[{"type": "character"}]`,
		},
		{
			name:     "JSON with whitespace",
			input:    "  \n[{\"type\": \"character\"}]\n  ",
			expected: `[{"type": "character"}]`,
		},
		{
			name:     "empty array",
			input:    "[]",
			expected: "[]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanJSONResponse(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestObjectToString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "string value",
			input:    "hobbit",
			expected: "hobbit",
		},
		{
			name:     "integer as float64",
			input:    float64(42),
			expected: "42",
		},
		{
			name:     "float value",
			input:    float64(3.14),
			expected: "3.14",
		},
		{
			name:     "int value",
			input:    100,
			expected: "100",
		},
		{
			name:     "bool true",
			input:    true,
			expected: "true",
		},
		{
			name:     "bool false",
			input:    false,
			expected: "false",
		},
		{
			name:     "nil value",
			input:    nil,
			expected: "<nil>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := objectToString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFactsToRaw(t *testing.T) {
	facts := []entities.Fact{
		{
			Type:       entities.FactTypeCharacter,
			Subject:    "Frodo",
			Predicate:  "is_a",
			Object:     "hobbit",
			Context:    "from the Shire",
			Confidence: 0.95,
		},
		{
			Type:       entities.FactTypeLocation,
			Subject:    "Mordor",
			Predicate:  "is",
			Object:     "dangerous",
			Confidence: 0.9,
		},
	}

	raw := factsToRaw(facts)

	require.Len(t, raw, 2)

	assert.Equal(t, "character", raw[0].Type)
	assert.Equal(t, "Frodo", raw[0].Subject)
	assert.Equal(t, "is_a", raw[0].Predicate)
	assert.Equal(t, "hobbit", raw[0].Object)
	assert.Equal(t, "from the Shire", raw[0].Context)
	assert.Equal(t, 0.95, raw[0].Confidence)

	assert.Equal(t, "location", raw[1].Type)
	assert.Equal(t, "Mordor", raw[1].Subject)
}

func TestFactsToRaw_Empty(t *testing.T) {
	raw := factsToRaw([]entities.Fact{})
	assert.Empty(t, raw)
}
