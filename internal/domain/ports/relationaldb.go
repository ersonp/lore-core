package ports

import (
	"context"

	"github.com/ersonp/lore-core/internal/domain/entities"
)

// RelationalDB defines the interface for relational database operations.
// This interface handles data that requires transactions, relationships,
// and complex queries - complementing VectorDB for semantic search.
type RelationalDB interface {
	// EnsureSchema creates the database schema if it doesn't exist.
	EnsureSchema(ctx context.Context) error

	// Close closes the database connection.
	Close() error

	// Entity operations

	// SaveEntity saves or updates an entity.
	SaveEntity(ctx context.Context, entity *entities.Entity) error

	// FindEntityByName finds an entity by its normalized name (case-insensitive).
	FindEntityByName(ctx context.Context, worldID, name string) (*entities.Entity, error)

	// FindOrCreateEntity finds an entity by name or creates it if not found.
	FindOrCreateEntity(ctx context.Context, worldID, name string) (*entities.Entity, error)

	// FindEntityByID finds an entity by its ID.
	FindEntityByID(ctx context.Context, entityID string) (*entities.Entity, error)

	// ListEntities lists all entities for a world with pagination.
	ListEntities(ctx context.Context, worldID string, limit, offset int) ([]*entities.Entity, error)

	// SearchEntities searches entities by name pattern.
	SearchEntities(ctx context.Context, worldID, query string, limit int) ([]*entities.Entity, error)

	// DeleteEntity deletes an entity by ID.
	DeleteEntity(ctx context.Context, entityID string) error

	// CountEntities returns the total number of entities for a world.
	CountEntities(ctx context.Context, worldID string) (int, error)

	// Relationship operations

	// SaveRelationship saves or updates a relationship.
	SaveRelationship(ctx context.Context, rel *entities.Relationship) error

	// FindRelationshipsByEntity finds all relationships involving an entity.
	// Returns relationships where the entity is source, or target if bidirectional.
	FindRelationshipsByEntity(ctx context.Context, entityID string) ([]entities.Relationship, error)

	// FindRelationshipsByType finds all relationships of a given type.
	FindRelationshipsByType(ctx context.Context, relType string) ([]entities.Relationship, error)

	// DeleteRelationship deletes a relationship by ID.
	DeleteRelationship(ctx context.Context, id string) error

	// DeleteRelationshipsByEntity deletes all relationships involving an entity.
	DeleteRelationshipsByEntity(ctx context.Context, entityID string) error

	// FindRelationshipBetween finds a direct relationship between two entities.
	// Returns nil if no relationship exists.
	FindRelationshipBetween(ctx context.Context, sourceEntityID, targetEntityID string) (*entities.Relationship, error)

	// FindRelatedEntities finds all entity IDs connected to the given entity up to the specified depth.
	// Depth 1 returns directly connected entities, depth 2 includes their connections, etc.
	FindRelatedEntities(ctx context.Context, entityID string, depth int) ([]string, error)

	// CountRelationships returns the total number of relationships in the database.
	CountRelationships(ctx context.Context) (int, error)

	// SaveVersion saves a new fact version.
	SaveVersion(ctx context.Context, version *entities.FactVersion) error

	// FindVersionsByFact finds all versions of a fact, ordered by version descending.
	FindVersionsByFact(ctx context.Context, factID string) ([]entities.FactVersion, error)

	// FindLatestVersion finds the most recent version of a fact.
	FindLatestVersion(ctx context.Context, factID string) (*entities.FactVersion, error)

	// CountVersions counts how many versions a fact has.
	CountVersions(ctx context.Context, factID string) (int, error)

	// SaveEntityType saves or updates a custom entity type.
	SaveEntityType(ctx context.Context, entityType *entities.EntityType) error

	// FindEntityType finds a custom entity type by name.
	FindEntityType(ctx context.Context, name string) (*entities.EntityType, error)

	// ListEntityTypes lists all custom entity types.
	ListEntityTypes(ctx context.Context) ([]entities.EntityType, error)

	// DeleteEntityType deletes a custom entity type by name.
	DeleteEntityType(ctx context.Context, name string) error

	// LogAction logs an action to the audit log.
	LogAction(ctx context.Context, action string, factID string, details map[string]any) error

	// FindAuditLog finds audit log entries for a specific fact.
	FindAuditLog(ctx context.Context, factID string) ([]entities.AuditEntry, error)

	// FindAuditLogByAction finds audit log entries by action type.
	FindAuditLogByAction(ctx context.Context, action string, limit int) ([]entities.AuditEntry, error)
}
