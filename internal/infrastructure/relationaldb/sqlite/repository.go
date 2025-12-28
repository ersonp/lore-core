// Package sqlite provides a SQLite implementation of the RelationalDB interface.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// Repository implements ports.RelationalDB using SQLite.
type Repository struct {
	db   *sql.DB
	path string
}

// NewRepository creates a new SQLite repository.
func NewRepository(cfg config.SQLiteConfig) (*Repository, error) {
	if cfg.Path == "" {
		return nil, errors.New("sqlite path is required")
	}

	db, err := sql.Open("sqlite", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite database: %w", err)
	}

	// Enable foreign keys for referential integrity
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	// Enable WAL mode for better concurrent read/write performance
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enabling WAL mode: %w", err)
	}

	// Set busy timeout to avoid "database is locked" errors
	if _, err := db.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting busy timeout: %w", err)
	}

	return &Repository{
		db:   db,
		path: cfg.Path,
	}, nil
}

// Close closes the database connection.
func (r *Repository) Close() error {
	return r.db.Close()
}

// Path returns the database file path.
func (r *Repository) Path() string {
	return r.path
}

// EnsureSchema creates the database schema if it doesn't exist.
func (r *Repository) EnsureSchema(ctx context.Context) error {
	// TODO: Implement in Task 05
	return nil
}

// SaveRelationship saves or updates a relationship.
func (r *Repository) SaveRelationship(ctx context.Context, rel *entities.Relationship) error {
	// TODO: Implement in Task 06
	return nil
}

// FindRelationshipsByFact finds all relationships involving a fact.
func (r *Repository) FindRelationshipsByFact(ctx context.Context, factID string) ([]entities.Relationship, error) {
	// TODO: Implement in Task 06
	return nil, nil
}

// FindRelationshipsByType finds all relationships of a given type.
func (r *Repository) FindRelationshipsByType(ctx context.Context, relType string) ([]entities.Relationship, error) {
	// TODO: Implement in Task 06
	return nil, nil
}

// DeleteRelationship deletes a relationship by ID.
func (r *Repository) DeleteRelationship(ctx context.Context, id string) error {
	// TODO: Implement in Task 06
	return nil
}

// DeleteRelationshipsByFact deletes all relationships involving a fact.
func (r *Repository) DeleteRelationshipsByFact(ctx context.Context, factID string) error {
	// TODO: Implement in Task 06
	return nil
}

// SaveVersion saves a new fact version.
func (r *Repository) SaveVersion(ctx context.Context, version *entities.FactVersion) error {
	// TODO: Implement in Task 07
	return nil
}

// FindVersionsByFact finds all versions of a fact.
func (r *Repository) FindVersionsByFact(ctx context.Context, factID string) ([]entities.FactVersion, error) {
	// TODO: Implement in Task 07
	return nil, nil
}

// FindLatestVersion finds the most recent version of a fact.
func (r *Repository) FindLatestVersion(ctx context.Context, factID string) (*entities.FactVersion, error) {
	// TODO: Implement in Task 07
	return nil, nil
}

// CountVersions counts how many versions a fact has.
func (r *Repository) CountVersions(ctx context.Context, factID string) (int, error) {
	// TODO: Implement in Task 07
	return 0, nil
}

// SaveEntityType saves or updates a custom entity type.
func (r *Repository) SaveEntityType(ctx context.Context, entityType *entities.EntityType) error {
	// TODO: Implement in Task 08
	return nil
}

// FindEntityType finds a custom entity type by name.
func (r *Repository) FindEntityType(ctx context.Context, name string) (*entities.EntityType, error) {
	// TODO: Implement in Task 08
	return nil, nil
}

// ListEntityTypes lists all custom entity types.
func (r *Repository) ListEntityTypes(ctx context.Context) ([]entities.EntityType, error) {
	// TODO: Implement in Task 08
	return nil, nil
}

// DeleteEntityType deletes a custom entity type by name.
func (r *Repository) DeleteEntityType(ctx context.Context, name string) error {
	// TODO: Implement in Task 08
	return nil
}

// LogAction logs an action to the audit log.
func (r *Repository) LogAction(ctx context.Context, action string, factID string, details map[string]any) error {
	// TODO: Implement in Task 09
	return nil
}

// FindAuditLog finds audit log entries for a specific fact.
func (r *Repository) FindAuditLog(ctx context.Context, factID string) ([]entities.AuditEntry, error) {
	// TODO: Implement in Task 09
	return nil, nil
}

// FindAuditLogByAction finds audit log entries by action type.
func (r *Repository) FindAuditLogByAction(ctx context.Context, action string, limit int) ([]entities.AuditEntry, error) {
	// TODO: Implement in Task 09
	return nil, nil
}
