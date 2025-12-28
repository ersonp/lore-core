package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
	"github.com/ersonp/lore-core/internal/infrastructure/relationaldb/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteIntegration_FileDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create temp directory
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create repository
	repo, err := sqlite.NewRepository(config.SQLiteConfig{Path: dbPath})
	require.NoError(t, err)
	defer repo.Close()

	// Ensure schema
	err = repo.EnsureSchema(context.Background())
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(dbPath)
	require.NoError(t, err, "database file should exist")

	// Perform operations
	ctx := context.Background()

	// Create relationships
	rel := &entities.Relationship{
		ID:            "rel-1",
		SourceFactID:  "fact-1",
		TargetFactID:  "fact-2",
		Type:          entities.RelationAlly,
		Bidirectional: true,
		CreatedAt:     time.Now(),
	}
	err = repo.SaveRelationship(ctx, rel)
	require.NoError(t, err)

	// Create version
	version := &entities.FactVersion{
		ID:         "v1",
		FactID:     "fact-1",
		Version:    1,
		ChangeType: entities.ChangeCreation,
		Data: entities.Fact{
			ID:        "fact-1",
			Type:      entities.FactTypeCharacter,
			Subject:   "Test",
			Predicate: "is",
			Object:    "testing",
		},
		CreatedAt: time.Now(),
	}
	err = repo.SaveVersion(ctx, version)
	require.NoError(t, err)

	// Log action
	err = repo.LogAction(ctx, "test", "fact-1", map[string]any{"key": "value"})
	require.NoError(t, err)

	// Close and reopen
	repo.Close()

	repo2, err := sqlite.NewRepository(config.SQLiteConfig{Path: dbPath})
	require.NoError(t, err)
	defer repo2.Close()

	// Data should persist
	rels, err := repo2.FindRelationshipsByFact(ctx, "fact-1")
	require.NoError(t, err)
	assert.Len(t, rels, 1)

	versions, err := repo2.FindVersionsByFact(ctx, "fact-1")
	require.NoError(t, err)
	assert.Len(t, versions, 1)
}

func TestSQLiteIntegration_WALMode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "wal-test.db")

	repo, err := sqlite.NewRepository(config.SQLiteConfig{Path: dbPath})
	require.NoError(t, err)
	defer repo.Close()

	err = repo.EnsureSchema(context.Background())
	require.NoError(t, err)

	// Perform some writes to trigger WAL file creation
	for i := 0; i < 10; i++ {
		err := repo.LogAction(context.Background(), "test", "", nil)
		require.NoError(t, err)
	}

	// WAL file might be created (depends on SQLite behavior)
	// Just verify the database works correctly
	entries, err := repo.FindAuditLogByAction(context.Background(), "test", 100)
	require.NoError(t, err)
	assert.Len(t, entries, 10)
}

func TestSQLiteIntegration_ConcurrentReads(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "concurrent-test.db")

	repo, err := sqlite.NewRepository(config.SQLiteConfig{Path: dbPath})
	require.NoError(t, err)
	defer repo.Close()

	err = repo.EnsureSchema(context.Background())
	require.NoError(t, err)

	// Insert some data
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		et := &entities.EntityType{
			Name:        fmt.Sprintf("type-%d", i),
			Description: fmt.Sprintf("Type number %d", i),
			CreatedAt:   time.Now(),
		}
		err := repo.SaveEntityType(ctx, et)
		require.NoError(t, err)
	}

	// Concurrent reads
	errCh := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func() {
			types, err := repo.ListEntityTypes(context.Background())
			if err != nil {
				errCh <- err
				return
			}
			if len(types) != 100 {
				errCh <- fmt.Errorf("expected 100 types, got %d", len(types))
				return
			}
			errCh <- nil
		}()
	}

	for i := 0; i < 10; i++ {
		err := <-errCh
		require.NoError(t, err)
	}
}

func TestSQLiteIntegration_WorldLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	worldDir := filepath.Join(tmpDir, ".lore", "worlds", "test_world")

	// Simulate world creation
	err := os.MkdirAll(worldDir, 0755)
	require.NoError(t, err)

	dbPath := filepath.Join(worldDir, "lore.db")

	// Create and initialize
	repo, err := sqlite.NewRepository(config.SQLiteConfig{Path: dbPath})
	require.NoError(t, err)

	err = repo.EnsureSchema(context.Background())
	require.NoError(t, err)

	// Add some data
	err = repo.SaveEntityType(context.Background(), &entities.EntityType{
		Name:        "custom",
		Description: "Custom type",
		CreatedAt:   time.Now(),
	})
	require.NoError(t, err)

	repo.Close()

	// Verify file exists
	_, err = os.Stat(dbPath)
	require.NoError(t, err)

	// Simulate world deletion
	err = os.Remove(dbPath)
	require.NoError(t, err)

	// Clean up WAL files if they exist
	os.Remove(dbPath + "-wal")
	os.Remove(dbPath + "-shm")

	// Verify deleted
	_, err = os.Stat(dbPath)
	require.True(t, os.IsNotExist(err))
}

func TestSQLiteIntegration_VersionHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "version-test.db")

	repo, err := sqlite.NewRepository(config.SQLiteConfig{Path: dbPath})
	require.NoError(t, err)
	defer repo.Close()

	err = repo.EnsureSchema(context.Background())
	require.NoError(t, err)

	ctx := context.Background()
	factID := "fact-version-test"

	// Create multiple versions
	for i := 1; i <= 5; i++ {
		v := &entities.FactVersion{
			ID:         fmt.Sprintf("v%d", i),
			FactID:     factID,
			Version:    i,
			ChangeType: entities.ChangeUpdate,
			Data: entities.Fact{
				ID:        factID,
				Type:      entities.FactTypeCharacter,
				Subject:   "Character",
				Predicate: "has_level",
				Object:    strconv.Itoa(i * 10),
			},
			Reason:    fmt.Sprintf("Level up to %d", i*10),
			CreatedAt: time.Now(),
		}
		err := repo.SaveVersion(ctx, v)
		require.NoError(t, err)
	}

	// Verify version count
	count, err := repo.CountVersions(ctx, factID)
	require.NoError(t, err)
	assert.Equal(t, 5, count)

	// Verify latest version
	latest, err := repo.FindLatestVersion(ctx, factID)
	require.NoError(t, err)
	require.NotNil(t, latest)
	assert.Equal(t, 5, latest.Version)
	assert.Equal(t, "50", latest.Data.Object)

	// Verify all versions ordered correctly
	versions, err := repo.FindVersionsByFact(ctx, factID)
	require.NoError(t, err)
	assert.Len(t, versions, 5)
	// Should be ordered DESC
	assert.Equal(t, 5, versions[0].Version)
	assert.Equal(t, 1, versions[4].Version)
}
