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
	schema := `
	-- Entity relationships (connects two facts)
	CREATE TABLE IF NOT EXISTS relationships (
		id TEXT PRIMARY KEY,
		source_fact_id TEXT NOT NULL,
		target_fact_id TEXT NOT NULL,
		type TEXT NOT NULL,
		bidirectional INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_relationships_source ON relationships(source_fact_id);
	CREATE INDEX IF NOT EXISTS idx_relationships_target ON relationships(target_fact_id);
	CREATE INDEX IF NOT EXISTS idx_relationships_type ON relationships(type);

	-- Fact version history (tracks changes over time)
	CREATE TABLE IF NOT EXISTS fact_versions (
		id TEXT PRIMARY KEY,
		fact_id TEXT NOT NULL,
		version INTEGER NOT NULL,
		change_type TEXT NOT NULL,
		data TEXT NOT NULL,
		reason TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(fact_id, version)
	);
	CREATE INDEX IF NOT EXISTS idx_fact_versions_fact ON fact_versions(fact_id);
	CREATE INDEX IF NOT EXISTS idx_fact_versions_type ON fact_versions(change_type);

	-- Custom entity types (user-defined extensions to FactType)
	CREATE TABLE IF NOT EXISTS entity_types (
		name TEXT PRIMARY KEY,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Audit log (tracks all actions)
	CREATE TABLE IF NOT EXISTS audit_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		action TEXT NOT NULL,
		fact_id TEXT,
		details TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_audit_log_fact ON audit_log(fact_id);
	CREATE INDEX IF NOT EXISTS idx_audit_log_action ON audit_log(action);
	CREATE INDEX IF NOT EXISTS idx_audit_log_created ON audit_log(created_at);
	`

	_, err := r.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("creating schema: %w", err)
	}
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
