package integration

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/services"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
	"github.com/ersonp/lore-core/internal/infrastructure/relationaldb/sqlite"
)

func TestEntityType_Integration_LoadDefaults(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := sqlite.NewRepository(config.SQLiteConfig{Path: dbPath})
	require.NoError(t, err)
	defer repo.Close()

	err = repo.EnsureSchema(context.Background())
	require.NoError(t, err)

	svc := services.NewEntityTypeService(repo)

	// Load defaults
	err = svc.LoadDefaults(context.Background())
	require.NoError(t, err)

	// Verify 6 types
	types, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, types, 6)

	// Verify each default exists
	for _, dt := range entities.DefaultEntityTypes {
		assert.True(t, svc.IsValid(context.Background(), dt.Name), "expected %s to be valid", dt.Name)
	}
}

func TestEntityType_Integration_CustomTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := sqlite.NewRepository(config.SQLiteConfig{Path: dbPath})
	require.NoError(t, err)
	defer repo.Close()

	err = repo.EnsureSchema(context.Background())
	require.NoError(t, err)

	svc := services.NewEntityTypeService(repo)

	// Load defaults first
	err = svc.LoadDefaults(context.Background())
	require.NoError(t, err)

	// Add custom type
	err = svc.Add(context.Background(), "weapon", "Weapons and artifacts")
	require.NoError(t, err)

	// Verify 7 types now
	types, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, types, 7)

	// Verify custom type is valid
	assert.True(t, svc.IsValid(context.Background(), "weapon"))

	// Remove custom type
	err = svc.Remove(context.Background(), "weapon")
	require.NoError(t, err)

	// Back to 6
	types, err = svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, types, 6)
}

func TestEntityType_Integration_CannotRemoveDefault(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := sqlite.NewRepository(config.SQLiteConfig{Path: dbPath})
	require.NoError(t, err)
	defer repo.Close()

	err = repo.EnsureSchema(context.Background())
	require.NoError(t, err)

	svc := services.NewEntityTypeService(repo)

	err = svc.LoadDefaults(context.Background())
	require.NoError(t, err)

	// Try to remove default type
	err = svc.Remove(context.Background(), "character")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot remove default")
}

func TestEntityType_Integration_Persistence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// First session: add custom type
	repo1, err := sqlite.NewRepository(config.SQLiteConfig{Path: dbPath})
	require.NoError(t, err)

	err = repo1.EnsureSchema(context.Background())
	require.NoError(t, err)

	svc1 := services.NewEntityTypeService(repo1)
	err = svc1.Add(context.Background(), "weapon", "Weapons")
	require.NoError(t, err)

	repo1.Close()

	// Second session: verify persistence
	repo2, err := sqlite.NewRepository(config.SQLiteConfig{Path: dbPath})
	require.NoError(t, err)
	defer repo2.Close()

	svc2 := services.NewEntityTypeService(repo2)
	assert.True(t, svc2.IsValid(context.Background(), "weapon"))
}

func TestEntityType_Integration_PromptList(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := sqlite.NewRepository(config.SQLiteConfig{Path: dbPath})
	require.NoError(t, err)
	defer repo.Close()

	err = repo.EnsureSchema(context.Background())
	require.NoError(t, err)

	svc := services.NewEntityTypeService(repo)
	err = svc.LoadDefaults(context.Background())
	require.NoError(t, err)

	err = svc.Add(context.Background(), "weapon", "Weapons")
	require.NoError(t, err)

	// Build prompt list
	list, err := svc.BuildPromptTypeList(context.Background())
	require.NoError(t, err)

	// Should contain all types
	assert.Contains(t, list, "character")
	assert.Contains(t, list, "location")
	assert.Contains(t, list, "weapon")
}

func TestEntityType_Integration_IdempotentDefaults(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := sqlite.NewRepository(config.SQLiteConfig{Path: dbPath})
	require.NoError(t, err)
	defer repo.Close()

	err = repo.EnsureSchema(context.Background())
	require.NoError(t, err)

	svc := services.NewEntityTypeService(repo)

	// Load defaults multiple times - should be idempotent
	for i := 0; i < 3; i++ {
		err = svc.LoadDefaults(context.Background())
		require.NoError(t, err)
	}

	// Still should have exactly 6 types
	types, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, types, 6)
}

func TestEntityType_Integration_ValidationRules(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := sqlite.NewRepository(config.SQLiteConfig{Path: dbPath})
	require.NoError(t, err)
	defer repo.Close()

	err = repo.EnsureSchema(context.Background())
	require.NoError(t, err)

	svc := services.NewEntityTypeService(repo)

	tests := []struct {
		name        string
		typeName    string
		shouldError bool
	}{
		{"valid lowercase", "weapon", false},
		{"valid with underscore", "magic_item", false},
		{"valid with number", "item2", false},
		{"invalid uppercase", "Weapon", true},
		{"invalid space", "magic item", true},
		{"invalid hyphen", "magic-item", true},
		{"invalid starts with number", "2item", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.Add(context.Background(), tc.typeName, "Test description")
			if tc.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Clean up for next test
				_ = svc.Remove(context.Background(), tc.typeName)
			}
		})
	}
}
