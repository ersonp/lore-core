package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRepo creates an in-memory SQLite repository for testing.
func setupTestRepo(t *testing.T) *Repository {
	t.Helper()
	repo, err := NewRepository(config.SQLiteConfig{Path: ":memory:"})
	require.NoError(t, err)
	t.Cleanup(func() { repo.Close() })

	err = repo.EnsureSchema(context.Background())
	require.NoError(t, err)

	return repo
}

func TestNewRepository(t *testing.T) {
	t.Run("success with memory database", func(t *testing.T) {
		repo, err := NewRepository(config.SQLiteConfig{Path: ":memory:"})
		require.NoError(t, err)
		defer repo.Close()
		assert.NotNil(t, repo)
	})

	t.Run("error with empty path", func(t *testing.T) {
		_, err := NewRepository(config.SQLiteConfig{Path: ""})
		require.Error(t, err)
	})
}

func TestRepository_EnsureSchema(t *testing.T) {
	repo := setupTestRepo(t)

	// Verify tables exist
	tables := []string{"entities", "relationships", "fact_versions", "entity_types", "audit_log"}
	for _, table := range tables {
		var count int
		err := repo.db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "table %s should exist", table)
	}
}

func TestRepository_EnsureSchema_Idempotent(t *testing.T) {
	repo := setupTestRepo(t)

	// Should not error when called again
	err := repo.EnsureSchema(context.Background())
	require.NoError(t, err)
}

func TestRepository_Relationships(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	t.Run("save and find by entity", func(t *testing.T) {
		rel := &entities.Relationship{
			ID:             "rel-1",
			SourceEntityID: "entity-1",
			TargetEntityID: "entity-2",
			Type:           entities.RelationAlly,
			Bidirectional:  false,
			CreatedAt:      time.Now(),
		}

		err := repo.SaveRelationship(ctx, rel)
		require.NoError(t, err)

		found, err := repo.FindRelationshipsByEntity(ctx, "entity-1")
		require.NoError(t, err)
		require.Len(t, found, 1)
		assert.Equal(t, "rel-1", found[0].ID)
	})

	t.Run("bidirectional relationship", func(t *testing.T) {
		rel := &entities.Relationship{
			ID:             "rel-2",
			SourceEntityID: "entity-3",
			TargetEntityID: "entity-4",
			Type:           entities.RelationSibling,
			Bidirectional:  true,
			CreatedAt:      time.Now(),
		}

		err := repo.SaveRelationship(ctx, rel)
		require.NoError(t, err)

		// Should find from target side too
		found, err := repo.FindRelationshipsByEntity(ctx, "entity-4")
		require.NoError(t, err)
		require.Len(t, found, 1)
	})

	t.Run("find by type", func(t *testing.T) {
		found, err := repo.FindRelationshipsByType(ctx, "ally")
		require.NoError(t, err)
		assert.Len(t, found, 1)
	})

	t.Run("delete relationship", func(t *testing.T) {
		err := repo.DeleteRelationship(ctx, "rel-1")
		require.NoError(t, err)

		found, err := repo.FindRelationshipsByEntity(ctx, "entity-1")
		require.NoError(t, err)
		assert.Len(t, found, 0)
	})

	t.Run("delete nonexistent relationship", func(t *testing.T) {
		err := repo.DeleteRelationship(ctx, "nonexistent")
		require.Error(t, err)
	})

	t.Run("delete by entity", func(t *testing.T) {
		err := repo.DeleteRelationshipsByEntity(ctx, "entity-3")
		require.NoError(t, err)

		found, err := repo.FindRelationshipsByEntity(ctx, "entity-3")
		require.NoError(t, err)
		assert.Len(t, found, 0)
	})
}

func TestRepository_FindRelationshipBetween(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	// Setup test relationships
	rel1 := &entities.Relationship{
		ID:             "rel-between-1",
		SourceEntityID: "alice",
		TargetEntityID: "bob",
		Type:           entities.RelationAlly,
		Bidirectional:  false,
		CreatedAt:      time.Now(),
	}
	rel2 := &entities.Relationship{
		ID:             "rel-between-2",
		SourceEntityID: "eve",
		TargetEntityID: "charlie",
		Type:           entities.RelationSibling,
		Bidirectional:  true,
		CreatedAt:      time.Now(),
	}

	require.NoError(t, repo.SaveRelationship(ctx, rel1))
	require.NoError(t, repo.SaveRelationship(ctx, rel2))

	t.Run("direct relationship", func(t *testing.T) {
		found, err := repo.FindRelationshipBetween(ctx, "alice", "bob")
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "rel-between-1", found.ID)
		assert.Equal(t, entities.RelationAlly, found.Type)
	})

	t.Run("reverse direction non-bidirectional returns nil", func(t *testing.T) {
		found, err := repo.FindRelationshipBetween(ctx, "bob", "alice")
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("bidirectional from source", func(t *testing.T) {
		found, err := repo.FindRelationshipBetween(ctx, "eve", "charlie")
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "rel-between-2", found.ID)
	})

	t.Run("bidirectional from target", func(t *testing.T) {
		found, err := repo.FindRelationshipBetween(ctx, "charlie", "eve")
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "rel-between-2", found.ID)
	})

	t.Run("no relationship", func(t *testing.T) {
		found, err := repo.FindRelationshipBetween(ctx, "alice", "eve")
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("nonexistent entities", func(t *testing.T) {
		found, err := repo.FindRelationshipBetween(ctx, "unknown1", "unknown2")
		require.NoError(t, err)
		assert.Nil(t, found)
	})
}

// setupGraphTestData creates a test graph: A --ally--> B --sibling--> C --enemy--> D, A --located_in--> City
func setupGraphTestData(t *testing.T, repo *Repository) {
	t.Helper()
	ctx := context.Background()
	relationships := []*entities.Relationship{
		{ID: "rel-ab", SourceEntityID: "A", TargetEntityID: "B", Type: entities.RelationAlly, Bidirectional: true, CreatedAt: time.Now()},
		{ID: "rel-bc", SourceEntityID: "B", TargetEntityID: "C", Type: entities.RelationSibling, Bidirectional: true, CreatedAt: time.Now()},
		{ID: "rel-cd", SourceEntityID: "C", TargetEntityID: "D", Type: entities.RelationEnemy, Bidirectional: false, CreatedAt: time.Now()},
		{ID: "rel-a-city", SourceEntityID: "A", TargetEntityID: "City", Type: entities.RelationLocatedIn, Bidirectional: false, CreatedAt: time.Now()},
	}
	for _, rel := range relationships {
		require.NoError(t, repo.SaveRelationship(ctx, rel))
	}
}

func TestRepository_FindRelatedEntities(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()
	setupGraphTestData(t, repo)

	t.Run("depth 0 returns empty", func(t *testing.T) {
		entityIDs, err := repo.FindRelatedEntities(ctx, "A", 0)
		require.NoError(t, err)
		assert.Empty(t, entityIDs)
	})

	t.Run("depth 1 returns direct connections", func(t *testing.T) {
		entityIDs, err := repo.FindRelatedEntities(ctx, "A", 1)
		require.NoError(t, err)
		assert.Len(t, entityIDs, 2)
		assert.Contains(t, entityIDs, "B")
		assert.Contains(t, entityIDs, "City")
	})

	t.Run("depth 2 includes 2-hop connections", func(t *testing.T) {
		entityIDs, err := repo.FindRelatedEntities(ctx, "A", 2)
		require.NoError(t, err)
		assert.Len(t, entityIDs, 3)
		assert.Contains(t, entityIDs, "C")
	})

	t.Run("depth 3 reaches D", func(t *testing.T) {
		entityIDs, err := repo.FindRelatedEntities(ctx, "A", 3)
		require.NoError(t, err)
		assert.Len(t, entityIDs, 4)
		assert.Contains(t, entityIDs, "D")
	})

	t.Run("respects non-bidirectional direction", func(t *testing.T) {
		entityIDs, err := repo.FindRelatedEntities(ctx, "D", 1)
		require.NoError(t, err)
		assert.Empty(t, entityIDs)
	})

	t.Run("bidirectional traversal from target", func(t *testing.T) {
		entityIDs, err := repo.FindRelatedEntities(ctx, "C", 1)
		require.NoError(t, err)
		assert.Len(t, entityIDs, 2)
		assert.Contains(t, entityIDs, "B")
		assert.Contains(t, entityIDs, "D")
	})

	t.Run("no relationships returns empty", func(t *testing.T) {
		entityIDs, err := repo.FindRelatedEntities(ctx, "isolated", 5)
		require.NoError(t, err)
		assert.Empty(t, entityIDs)
	})

	t.Run("excludes self from results", func(t *testing.T) {
		entityIDs, err := repo.FindRelatedEntities(ctx, "A", 10)
		require.NoError(t, err)
		assert.NotContains(t, entityIDs, "A")
	})
}

func TestRepository_CountRelationships(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	t.Run("empty database", func(t *testing.T) {
		count, err := repo.CountRelationships(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("with relationships", func(t *testing.T) {
		relationships := []*entities.Relationship{
			{
				ID:             "count-1",
				SourceEntityID: "x",
				TargetEntityID: "y",
				Type:           entities.RelationAlly,
				Bidirectional:  false,
				CreatedAt:      time.Now(),
			},
			{
				ID:             "count-2",
				SourceEntityID: "y",
				TargetEntityID: "z",
				Type:           entities.RelationEnemy,
				Bidirectional:  true,
				CreatedAt:      time.Now(),
			},
		}

		for _, rel := range relationships {
			require.NoError(t, repo.SaveRelationship(ctx, rel))
		}

		count, err := repo.CountRelationships(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("after delete", func(t *testing.T) {
		err := repo.DeleteRelationship(ctx, "count-1")
		require.NoError(t, err)

		count, err := repo.CountRelationships(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

func TestRepository_FactVersions(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	fact := entities.Fact{
		ID:        "fact-1",
		Type:      entities.FactTypeCharacter,
		Subject:   "Gandalf",
		Predicate: "has_trait",
		Object:    "wise",
	}

	t.Run("save and find versions", func(t *testing.T) {
		v1 := &entities.FactVersion{
			ID:         "v1",
			FactID:     "fact-1",
			Version:    1,
			ChangeType: entities.ChangeCreation,
			Data:       fact,
			Reason:     "Initial",
			CreatedAt:  time.Now(),
		}

		err := repo.SaveVersion(ctx, v1)
		require.NoError(t, err)

		versions, err := repo.FindVersionsByFact(ctx, "fact-1")
		require.NoError(t, err)
		require.Len(t, versions, 1)
		assert.Equal(t, 1, versions[0].Version)
	})

	t.Run("find latest version", func(t *testing.T) {
		fact.Object = "very wise"
		v2 := &entities.FactVersion{
			ID:         "v2",
			FactID:     "fact-1",
			Version:    2,
			ChangeType: entities.ChangeUpdate,
			Data:       fact,
			Reason:     "Updated",
			CreatedAt:  time.Now(),
		}

		err := repo.SaveVersion(ctx, v2)
		require.NoError(t, err)

		latest, err := repo.FindLatestVersion(ctx, "fact-1")
		require.NoError(t, err)
		require.NotNil(t, latest)
		assert.Equal(t, 2, latest.Version)
		assert.Equal(t, "very wise", latest.Data.Object)
	})

	t.Run("count versions", func(t *testing.T) {
		count, err := repo.CountVersions(ctx, "fact-1")
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("find latest for nonexistent fact", func(t *testing.T) {
		latest, err := repo.FindLatestVersion(ctx, "nonexistent")
		require.NoError(t, err)
		assert.Nil(t, latest)
	})
}

func TestRepository_EntityTypes(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	t.Run("save and find", func(t *testing.T) {
		et := &entities.EntityType{
			Name:        "weapon",
			Description: "Weapons and artifacts",
			CreatedAt:   time.Now(),
		}

		err := repo.SaveEntityType(ctx, et)
		require.NoError(t, err)

		found, err := repo.FindEntityType(ctx, "weapon")
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "weapon", found.Name)
	})

	t.Run("list all", func(t *testing.T) {
		et2 := &entities.EntityType{
			Name:        "organization",
			Description: "Groups and factions",
			CreatedAt:   time.Now(),
		}
		err := repo.SaveEntityType(ctx, et2)
		require.NoError(t, err)

		types, err := repo.ListEntityTypes(ctx)
		require.NoError(t, err)
		assert.Len(t, types, 2)
	})

	t.Run("update existing", func(t *testing.T) {
		et := &entities.EntityType{
			Name:        "weapon",
			Description: "Updated description",
			CreatedAt:   time.Now(),
		}

		err := repo.SaveEntityType(ctx, et)
		require.NoError(t, err)

		found, err := repo.FindEntityType(ctx, "weapon")
		require.NoError(t, err)
		assert.Equal(t, "Updated description", found.Description)
	})

	t.Run("delete", func(t *testing.T) {
		err := repo.DeleteEntityType(ctx, "weapon")
		require.NoError(t, err)

		found, err := repo.FindEntityType(ctx, "weapon")
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("delete nonexistent", func(t *testing.T) {
		err := repo.DeleteEntityType(ctx, "nonexistent")
		require.Error(t, err)
	})

	t.Run("find nonexistent", func(t *testing.T) {
		found, err := repo.FindEntityType(ctx, "nonexistent")
		require.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestRepository_AuditLog(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	t.Run("log action with details", func(t *testing.T) {
		err := repo.LogAction(ctx, "ingest", "fact-1", map[string]any{
			"source": "chapter1.txt",
			"count":  5,
		})
		require.NoError(t, err)
	})

	t.Run("log action without fact ID", func(t *testing.T) {
		err := repo.LogAction(ctx, "export", "", map[string]any{
			"format": "json",
		})
		require.NoError(t, err)
	})

	t.Run("log action without details", func(t *testing.T) {
		err := repo.LogAction(ctx, "query", "fact-2", nil)
		require.NoError(t, err)
	})

	t.Run("find by fact", func(t *testing.T) {
		entries, err := repo.FindAuditLog(ctx, "fact-1")
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, "ingest", entries[0].Action)
		assert.Equal(t, "chapter1.txt", entries[0].Details["source"])
	})

	t.Run("find by action", func(t *testing.T) {
		entries, err := repo.FindAuditLogByAction(ctx, "export", 10)
		require.NoError(t, err)
		assert.Len(t, entries, 1)
	})

	t.Run("find by action with limit", func(t *testing.T) {
		// Log more actions
		for i := 0; i < 5; i++ {
			err := repo.LogAction(ctx, "bulk", "", nil)
			require.NoError(t, err)
		}

		entries, err := repo.FindAuditLogByAction(ctx, "bulk", 3)
		require.NoError(t, err)
		assert.Len(t, entries, 3)
	})
}

func TestRepository_Path(t *testing.T) {
	repo, err := NewRepository(config.SQLiteConfig{Path: ":memory:"})
	require.NoError(t, err)
	defer repo.Close()

	assert.Equal(t, ":memory:", repo.Path())
}
