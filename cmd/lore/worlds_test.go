package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ersonp/lore-core/internal/infrastructure/config"
)

func TestAddWorldToConfig(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Initialize config first
	err := config.WriteDefaultWithWorld(tmpDir, "initial", "Initial world")
	require.NoError(t, err)

	// Add a new world
	newWorld := config.WorldConfig{
		Collection:  "lore_test_world",
		Description: "Test world description",
	}
	err = addWorldToConfig(tmpDir, "test-world", newWorld)
	require.NoError(t, err)

	// Verify the world was added
	cfg, err := config.Load(tmpDir)
	require.NoError(t, err)

	assert.Contains(t, cfg.Worlds, "test-world")
	assert.Equal(t, "lore_test_world", cfg.Worlds["test-world"].Collection)
	assert.Equal(t, "Test world description", cfg.Worlds["test-world"].Description)
}

func TestAddWorldToConfig_CreatesWorldsMapIfNil(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal config without worlds section
	configDir := filepath.Join(tmpDir, ".lore")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	configContent := `llm:
  provider: openai
  model: gpt-4o-mini
qdrant:
  host: localhost
  port: 6334
`
	err = os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configContent), 0600)
	require.NoError(t, err)

	// Add a world
	newWorld := config.WorldConfig{
		Collection:  "lore_new_world",
		Description: "New world",
	}
	err = addWorldToConfig(tmpDir, "new-world", newWorld)
	require.NoError(t, err)

	// Verify world was added
	cfg, err := config.Load(tmpDir)
	require.NoError(t, err)
	assert.Contains(t, cfg.Worlds, "new-world")
}

func TestRemoveWorldFromConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize config with a world
	err := config.WriteDefaultWithWorld(tmpDir, "to-delete", "World to delete")
	require.NoError(t, err)

	// Verify world exists
	cfg, err := config.Load(tmpDir)
	require.NoError(t, err)
	assert.Contains(t, cfg.Worlds, "to-delete")

	// Remove the world
	err = removeWorldFromConfig(tmpDir, "to-delete")
	require.NoError(t, err)

	// Verify world was removed
	cfg, err = config.Load(tmpDir)
	require.NoError(t, err)
	assert.NotContains(t, cfg.Worlds, "to-delete")
}

func TestRemoveWorldFromConfig_NonExistentWorld(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize config
	err := config.WriteDefaultWithWorld(tmpDir, "existing", "Existing world")
	require.NoError(t, err)

	// Remove non-existent world (should not error)
	err = removeWorldFromConfig(tmpDir, "non-existent")
	require.NoError(t, err)

	// Verify existing world still exists
	cfg, err := config.Load(tmpDir)
	require.NoError(t, err)
	assert.Contains(t, cfg.Worlds, "existing")
}

func TestAddWorldToConfig_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to add world without config file
	newWorld := config.WorldConfig{
		Collection:  "lore_test",
		Description: "Test",
	}
	err := addWorldToConfig(tmpDir, "test", newWorld)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reading config file")
}

func TestRemoveWorldFromConfig_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to remove world without config file
	err := removeWorldFromConfig(tmpDir, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reading config file")
}
