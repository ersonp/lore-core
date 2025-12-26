package openai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ersonp/lore-core/internal/infrastructure/config"
)

func TestNewEmbedder(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.EmbedderConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			cfg: config.EmbedderConfig{
				APIKey: "test-key",
			},
			wantErr: false,
		},
		{
			name: "valid config with model",
			cfg: config.EmbedderConfig{
				APIKey: "test-key",
				Model:  "text-embedding-ada-002",
			},
			wantErr: false,
		},
		{
			name:    "missing API key",
			cfg:     config.EmbedderConfig{},
			wantErr: true,
			errMsg:  "API key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			embedder, err := NewEmbedder(tt.cfg)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, embedder)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, embedder)
			}
		})
	}
}

func TestVectorSize(t *testing.T) {
	// Verify the constant matches OpenAI's text-embedding-3-small dimension
	assert.Equal(t, 1536, VectorSize)
}
