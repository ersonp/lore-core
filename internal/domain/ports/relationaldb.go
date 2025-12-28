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

	// SaveRelationship saves or updates a relationship.
	SaveRelationship(ctx context.Context, rel *entities.Relationship) error

	// FindRelationshipsByFact finds all relationships involving a fact.
	// Returns relationships where the fact is source, or target if bidirectional.
	FindRelationshipsByFact(ctx context.Context, factID string) ([]entities.Relationship, error)

	// FindRelationshipsByType finds all relationships of a given type.
	FindRelationshipsByType(ctx context.Context, relType string) ([]entities.Relationship, error)

	// DeleteRelationship deletes a relationship by ID.
	DeleteRelationship(ctx context.Context, id string) error

	// DeleteRelationshipsByFact deletes all relationships involving a fact.
	DeleteRelationshipsByFact(ctx context.Context, factID string) error

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
