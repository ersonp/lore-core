package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ersonp/lore-core/internal/infrastructure/config"
)

func TestWorldsConfig_Add(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize config first
	err := config.WriteDefaultWithWorld(tmpDir, "initial", "Initial world")
	require.NoError(t, err)

	// Load worlds and add a new one
	worlds, err := config.LoadWorlds(tmpDir)
	require.NoError(t, err)

	worlds.Add("test-world", config.WorldEntry{
		Collection:  "lore_test_world",
		Description: "Test world description",
	})

	err = worlds.Save(tmpDir)
	require.NoError(t, err)

	// Reload and verify
	worlds, err = config.LoadWorlds(tmpDir)
	require.NoError(t, err)

	assert.True(t, worlds.Exists("test-world"))
	entry, err := worlds.Get("test-world")
	require.NoError(t, err)
	assert.Equal(t, "lore_test_world", entry.Collection)
	assert.Equal(t, "Test world description", entry.Description)
}

func TestWorldsConfig_AddToEmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config directory
	configDir := filepath.Join(tmpDir, ".lore")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	// Load worlds (should return empty)
	worlds, err := config.LoadWorlds(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, worlds.Worlds)

	// Add a world
	worlds.Add("new-world", config.WorldEntry{
		Collection:  "lore_new_world",
		Description: "New world",
	})

	err = worlds.Save(tmpDir)
	require.NoError(t, err)

	// Reload and verify
	worlds, err = config.LoadWorlds(tmpDir)
	require.NoError(t, err)
	assert.True(t, worlds.Exists("new-world"))
}

func TestWorldsConfig_Remove(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize config with a world
	err := config.WriteDefaultWithWorld(tmpDir, "to-delete", "World to delete")
	require.NoError(t, err)

	// Verify world exists
	worlds, err := config.LoadWorlds(tmpDir)
	require.NoError(t, err)
	assert.True(t, worlds.Exists("to-delete"))

	// Remove the world
	worlds.Remove("to-delete")
	err = worlds.Save(tmpDir)
	require.NoError(t, err)

	// Reload and verify removal
	worlds, err = config.LoadWorlds(tmpDir)
	require.NoError(t, err)
	assert.False(t, worlds.Exists("to-delete"))
}

func TestWorldsConfig_RemoveNonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize config
	err := config.WriteDefaultWithWorld(tmpDir, "existing", "Existing world")
	require.NoError(t, err)

	// Load worlds
	worlds, err := config.LoadWorlds(tmpDir)
	require.NoError(t, err)

	// Remove non-existent world (should not error)
	worlds.Remove("non-existent")
	err = worlds.Save(tmpDir)
	require.NoError(t, err)

	// Verify existing world still exists
	worlds, err = config.LoadWorlds(tmpDir)
	require.NoError(t, err)
	assert.True(t, worlds.Exists("existing"))
}

func TestWorldsConfig_Get(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize config with a world
	err := config.WriteDefaultWithWorld(tmpDir, "test-world", "Test description")
	require.NoError(t, err)

	// Load and get world
	worlds, err := config.LoadWorlds(tmpDir)
	require.NoError(t, err)

	entry, err := worlds.Get("test-world")
	require.NoError(t, err)
	assert.Equal(t, "lore_test_world", entry.Collection)
	assert.Equal(t, "Test description", entry.Description)
}

func TestWorldsConfig_GetNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize config with a world
	err := config.WriteDefaultWithWorld(tmpDir, "existing", "Existing world")
	require.NoError(t, err)

	// Try to get non-existent world
	worlds, err := config.LoadWorlds(tmpDir)
	require.NoError(t, err)

	_, err = worlds.Get("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestWorldsConfig_GetCollection(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize config with a world
	err := config.WriteDefaultWithWorld(tmpDir, "my-world", "My world")
	require.NoError(t, err)

	// Load and get collection
	worlds, err := config.LoadWorlds(tmpDir)
	require.NoError(t, err)

	collection, err := worlds.GetCollection("my-world")
	require.NoError(t, err)
	assert.Equal(t, "lore_my_world", collection)
}

func TestLoadWorlds_NoFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config directory only
	configDir := filepath.Join(tmpDir, ".lore")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	// Load worlds without file (should return empty)
	worlds, err := config.LoadWorlds(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, worlds.Worlds)
}
