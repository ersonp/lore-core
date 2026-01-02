// Package sqlite provides a SQLite implementation of the RelationalDB interface.
package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
	"github.com/google/uuid"
	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// generateUUID returns a new UUID string.
func generateUUID() string {
	return uuid.New().String()
}

// timeNow returns the current time (can be mocked in tests).
var timeNow = time.Now

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
	-- Entities (named subjects that can have relationships)
	CREATE TABLE IF NOT EXISTS entities (
		id TEXT PRIMARY KEY,
		world_id TEXT NOT NULL,
		name TEXT NOT NULL,
		normalized_name TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(world_id, normalized_name)
	);
	CREATE INDEX IF NOT EXISTS idx_entities_world ON entities(world_id);
	CREATE INDEX IF NOT EXISTS idx_entities_normalized ON entities(world_id, normalized_name);

	-- Entity relationships (connects two entities)
	CREATE TABLE IF NOT EXISTS relationships (
		id TEXT PRIMARY KEY,
		source_entity_id TEXT NOT NULL,
		target_entity_id TEXT NOT NULL,
		type TEXT NOT NULL,
		bidirectional INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_relationships_source ON relationships(source_entity_id);
	CREATE INDEX IF NOT EXISTS idx_relationships_target ON relationships(target_entity_id);
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

// SaveEntity saves or updates an entity.
func (r *Repository) SaveEntity(ctx context.Context, entity *entities.Entity) error {
	query := `
		INSERT INTO entities (id, world_id, name, normalized_name, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(world_id, normalized_name) DO UPDATE SET
			name = excluded.name
	`
	_, err := r.db.ExecContext(ctx, query,
		entity.ID,
		entity.WorldID,
		entity.Name,
		entity.NormalizedName,
		entity.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("saving entity: %w", err)
	}
	return nil
}

// FindEntityByName finds an entity by its normalized name (case-insensitive).
func (r *Repository) FindEntityByName(ctx context.Context, worldID, name string) (*entities.Entity, error) {
	normalizedName := entities.NormalizeName(name)
	query := `
		SELECT id, world_id, name, normalized_name, created_at
		FROM entities
		WHERE world_id = ? AND normalized_name = ?
	`
	row := r.db.QueryRowContext(ctx, query, worldID, normalizedName)

	var entity entities.Entity
	err := row.Scan(
		&entity.ID,
		&entity.WorldID,
		&entity.Name,
		&entity.NormalizedName,
		&entity.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning entity: %w", err)
	}
	return &entity, nil
}

// FindOrCreateEntity finds an entity by name or creates it if not found.
// This method is atomic - it uses INSERT OR IGNORE followed by SELECT to avoid race conditions.
func (r *Repository) FindOrCreateEntity(ctx context.Context, worldID, name string) (*entities.Entity, error) {
	normalizedName := entities.NormalizeName(name)

	// Atomically insert if not exists (ON CONFLICT DO NOTHING)
	insertQuery := `
		INSERT OR IGNORE INTO entities (id, world_id, name, normalized_name, created_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, insertQuery,
		generateUUID(),
		worldID,
		name,
		normalizedName,
		timeNow(),
	)
	if err != nil {
		return nil, fmt.Errorf("inserting entity: %w", err)
	}

	// Always fetch the entity (either newly inserted or pre-existing)
	return r.FindEntityByName(ctx, worldID, name)
}

// FindEntityByID finds an entity by its ID.
func (r *Repository) FindEntityByID(ctx context.Context, entityID string) (*entities.Entity, error) {
	query := `
		SELECT id, world_id, name, normalized_name, created_at
		FROM entities
		WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, entityID)

	var entity entities.Entity
	err := row.Scan(
		&entity.ID,
		&entity.WorldID,
		&entity.Name,
		&entity.NormalizedName,
		&entity.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning entity: %w", err)
	}
	return &entity, nil
}

// FindEntitiesByIDs finds multiple entities by their IDs in a single query.
func (r *Repository) FindEntitiesByIDs(ctx context.Context, ids []string) ([]*entities.Entity, error) {
	if len(ids) == 0 {
		return []*entities.Entity{}, nil
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, world_id, name, normalized_name, created_at
		FROM entities
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying entities: %w", err)
	}
	defer rows.Close()

	result := make([]*entities.Entity, 0, len(ids))
	for rows.Next() {
		var entity entities.Entity
		if err := rows.Scan(
			&entity.ID,
			&entity.WorldID,
			&entity.Name,
			&entity.NormalizedName,
			&entity.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning entity: %w", err)
		}
		result = append(result, &entity)
	}
	return result, rows.Err()
}

// ListEntities lists all entities for a world with pagination.
func (r *Repository) ListEntities(ctx context.Context, worldID string, limit, offset int) ([]*entities.Entity, error) {
	query := `
		SELECT id, world_id, name, normalized_name, created_at
		FROM entities
		WHERE world_id = ?
		ORDER BY name ASC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, worldID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("querying entities: %w", err)
	}
	defer rows.Close()

	result := make([]*entities.Entity, 0, limit)
	for rows.Next() {
		var entity entities.Entity
		if err := rows.Scan(
			&entity.ID,
			&entity.WorldID,
			&entity.Name,
			&entity.NormalizedName,
			&entity.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning entity: %w", err)
		}
		result = append(result, &entity)
	}
	return result, rows.Err()
}

// SearchEntities searches entities by name pattern.
func (r *Repository) SearchEntities(ctx context.Context, worldID, query string, limit int) ([]*entities.Entity, error) {
	normalizedQuery := "%" + entities.NormalizeName(query) + "%"
	sqlQuery := `
		SELECT id, world_id, name, normalized_name, created_at
		FROM entities
		WHERE world_id = ? AND normalized_name LIKE ?
		ORDER BY name ASC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, sqlQuery, worldID, normalizedQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("searching entities: %w", err)
	}
	defer rows.Close()

	result := make([]*entities.Entity, 0, limit)
	for rows.Next() {
		var entity entities.Entity
		if err := rows.Scan(
			&entity.ID,
			&entity.WorldID,
			&entity.Name,
			&entity.NormalizedName,
			&entity.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning entity: %w", err)
		}
		result = append(result, &entity)
	}
	return result, rows.Err()
}

// DeleteEntity deletes an entity by ID.
func (r *Repository) DeleteEntity(ctx context.Context, entityID string) error {
	query := `DELETE FROM entities WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, entityID)
	if err != nil {
		return fmt.Errorf("deleting entity: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("entity not found: %s", entityID)
	}
	return nil
}

// CountEntities returns the total number of entities for a world.
func (r *Repository) CountEntities(ctx context.Context, worldID string) (int, error) {
	query := `SELECT COUNT(*) FROM entities WHERE world_id = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, worldID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting entities: %w", err)
	}
	return count, nil
}

// SaveRelationship saves or updates a relationship.
func (r *Repository) SaveRelationship(ctx context.Context, rel *entities.Relationship) error {
	query := `
		INSERT INTO relationships (id, source_entity_id, target_entity_id, type, bidirectional, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			source_entity_id = excluded.source_entity_id,
			target_entity_id = excluded.target_entity_id,
			type = excluded.type,
			bidirectional = excluded.bidirectional
	`
	_, err := r.db.ExecContext(ctx, query,
		rel.ID,
		rel.SourceEntityID,
		rel.TargetEntityID,
		string(rel.Type),
		rel.Bidirectional,
		rel.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("saving relationship: %w", err)
	}
	return nil
}

// FindRelationshipsByEntity finds all relationships involving an entity.
// Returns relationships where the entity is source, or target if bidirectional.
func (r *Repository) FindRelationshipsByEntity(ctx context.Context, entityID string) ([]entities.Relationship, error) {
	query := `
		SELECT id, source_entity_id, target_entity_id, type, bidirectional, created_at
		FROM relationships
		WHERE source_entity_id = ? OR (target_entity_id = ? AND bidirectional = 1)
		ORDER BY created_at DESC
	`
	return r.queryRelationships(ctx, query, entityID, entityID)
}

// FindRelationshipsByType finds all relationships of a given type.
func (r *Repository) FindRelationshipsByType(ctx context.Context, relType string) ([]entities.Relationship, error) {
	query := `
		SELECT id, source_entity_id, target_entity_id, type, bidirectional, created_at
		FROM relationships
		WHERE type = ?
		ORDER BY created_at DESC
	`
	return r.queryRelationships(ctx, query, relType)
}

// DeleteRelationship deletes a relationship by ID.
func (r *Repository) DeleteRelationship(ctx context.Context, id string) error {
	query := `DELETE FROM relationships WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting relationship: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("relationship not found: %s", id)
	}
	return nil
}

// DeleteRelationshipsByEntity deletes all relationships involving an entity.
func (r *Repository) DeleteRelationshipsByEntity(ctx context.Context, entityID string) error {
	query := `DELETE FROM relationships WHERE source_entity_id = ? OR target_entity_id = ?`
	_, err := r.db.ExecContext(ctx, query, entityID, entityID)
	if err != nil {
		return fmt.Errorf("deleting relationships by entity: %w", err)
	}
	return nil
}

// FindRelationshipBetween finds a direct relationship between two entities.
// Returns nil if no relationship exists. Checks both directions for bidirectional relationships.
func (r *Repository) FindRelationshipBetween(ctx context.Context, sourceEntityID, targetEntityID string) (*entities.Relationship, error) {
	query := `
		SELECT id, source_entity_id, target_entity_id, type, bidirectional, created_at
		FROM relationships
		WHERE (source_entity_id = ? AND target_entity_id = ?)
		   OR (bidirectional = 1 AND source_entity_id = ? AND target_entity_id = ?)
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, sourceEntityID, targetEntityID, targetEntityID, sourceEntityID)

	var rel entities.Relationship
	var relType string

	err := row.Scan(
		&rel.ID,
		&rel.SourceEntityID,
		&rel.TargetEntityID,
		&relType,
		&rel.Bidirectional,
		&rel.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning relationship: %w", err)
	}

	rel.Type = entities.RelationType(relType)
	return &rel, nil
}

// FindRelatedEntities finds all entity IDs connected to the given entity up to the specified depth.
// Depth 1 returns directly connected entities, depth 2 includes their connections, etc.
// Uses a recursive CTE for efficient graph traversal.
func (r *Repository) FindRelatedEntities(ctx context.Context, entityID string, depth int) ([]string, error) {
	if depth < 1 {
		return []string{}, nil
	}

	query := `
		WITH RECURSIVE related(entity_id, level) AS (
			-- Base case: direct connections from source
			SELECT target_entity_id, 1
			FROM relationships
			WHERE source_entity_id = ?
			UNION
			SELECT source_entity_id, 1
			FROM relationships
			WHERE target_entity_id = ? AND bidirectional = 1

			UNION

			-- Recursive case: connections from already found entities
			SELECT r.target_entity_id, related.level + 1
			FROM relationships r
			JOIN related ON r.source_entity_id = related.entity_id
			WHERE related.level < ?
			UNION
			SELECT r.source_entity_id, related.level + 1
			FROM relationships r
			JOIN related ON r.target_entity_id = related.entity_id AND r.bidirectional = 1
			WHERE related.level < ?
		)
		SELECT DISTINCT entity_id
		FROM related
		WHERE entity_id != ?
		ORDER BY entity_id
	`

	rows, err := r.db.QueryContext(ctx, query, entityID, entityID, depth, depth, entityID)
	if err != nil {
		return nil, fmt.Errorf("querying related entities: %w", err)
	}
	defer rows.Close()

	entityIDs := make([]string, 0, 16)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning entity id: %w", err)
		}
		entityIDs = append(entityIDs, id)
	}
	return entityIDs, rows.Err()
}

// CountRelationships returns the total number of relationships in the database.
func (r *Repository) CountRelationships(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM relationships`
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting relationships: %w", err)
	}
	return count, nil
}

// queryRelationships is a helper to execute relationship queries.
func (r *Repository) queryRelationships(ctx context.Context, query string, args ...any) ([]entities.Relationship, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying relationships: %w", err)
	}
	defer rows.Close()

	relationships := make([]entities.Relationship, 0, 16)
	for rows.Next() {
		var rel entities.Relationship
		var relType string
		if err := rows.Scan(
			&rel.ID,
			&rel.SourceEntityID,
			&rel.TargetEntityID,
			&relType,
			&rel.Bidirectional,
			&rel.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning relationship: %w", err)
		}
		rel.Type = entities.RelationType(relType)
		relationships = append(relationships, rel)
	}
	return relationships, rows.Err()
}

// SaveVersion saves a new fact version.
func (r *Repository) SaveVersion(ctx context.Context, version *entities.FactVersion) error {
	data, err := json.Marshal(version.Data)
	if err != nil {
		return fmt.Errorf("marshaling fact data: %w", err)
	}

	query := `
		INSERT INTO fact_versions (id, fact_id, version, change_type, data, reason, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err = r.db.ExecContext(ctx, query,
		version.ID,
		version.FactID,
		version.Version,
		string(version.ChangeType),
		string(data),
		version.Reason,
		version.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("saving fact version: %w", err)
	}
	return nil
}

// FindVersionsByFact finds all versions of a fact, ordered by version descending.
func (r *Repository) FindVersionsByFact(ctx context.Context, factID string) ([]entities.FactVersion, error) {
	query := `
		SELECT id, fact_id, version, change_type, data, reason, created_at
		FROM fact_versions
		WHERE fact_id = ?
		ORDER BY version DESC
	`
	rows, err := r.db.QueryContext(ctx, query, factID)
	if err != nil {
		return nil, fmt.Errorf("querying fact versions: %w", err)
	}
	defer rows.Close()

	versions := make([]entities.FactVersion, 0, 16)
	for rows.Next() {
		v, err := r.scanFactVersion(rows)
		if err != nil {
			return nil, err
		}
		versions = append(versions, *v)
	}
	return versions, rows.Err()
}

// FindLatestVersion finds the most recent version of a fact.
func (r *Repository) FindLatestVersion(ctx context.Context, factID string) (*entities.FactVersion, error) {
	query := `
		SELECT id, fact_id, version, change_type, data, reason, created_at
		FROM fact_versions
		WHERE fact_id = ?
		ORDER BY version DESC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, factID)

	var v entities.FactVersion
	var changeType, data string
	var reason sql.NullString

	err := row.Scan(
		&v.ID,
		&v.FactID,
		&v.Version,
		&changeType,
		&data,
		&reason,
		&v.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning fact version: %w", err)
	}

	v.ChangeType = entities.ChangeType(changeType)
	v.Reason = reason.String

	if err := json.Unmarshal([]byte(data), &v.Data); err != nil {
		return nil, fmt.Errorf("unmarshaling fact data: %w", err)
	}

	return &v, nil
}

// CountVersions counts how many versions a fact has.
func (r *Repository) CountVersions(ctx context.Context, factID string) (int, error) {
	query := `SELECT COUNT(*) FROM fact_versions WHERE fact_id = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, factID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting versions: %w", err)
	}
	return count, nil
}

// scanFactVersion is a helper to scan a fact version row.
func (r *Repository) scanFactVersion(rows *sql.Rows) (*entities.FactVersion, error) {
	var v entities.FactVersion
	var changeType, data string
	var reason sql.NullString

	err := rows.Scan(
		&v.ID,
		&v.FactID,
		&v.Version,
		&changeType,
		&data,
		&reason,
		&v.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning fact version: %w", err)
	}

	v.ChangeType = entities.ChangeType(changeType)
	v.Reason = reason.String

	if err := json.Unmarshal([]byte(data), &v.Data); err != nil {
		return nil, fmt.Errorf("unmarshaling fact data: %w", err)
	}

	return &v, nil
}

// SaveEntityType saves or updates a custom entity type.
func (r *Repository) SaveEntityType(ctx context.Context, entityType *entities.EntityType) error {
	query := `
		INSERT INTO entity_types (name, description, created_at)
		VALUES (?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			description = excluded.description
	`
	_, err := r.db.ExecContext(ctx, query,
		entityType.Name,
		entityType.Description,
		entityType.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("saving entity type: %w", err)
	}
	return nil
}

// FindEntityType finds a custom entity type by name.
func (r *Repository) FindEntityType(ctx context.Context, name string) (*entities.EntityType, error) {
	query := `
		SELECT name, description, created_at
		FROM entity_types
		WHERE name = ?
	`
	row := r.db.QueryRowContext(ctx, query, name)

	var et entities.EntityType
	var description sql.NullString

	err := row.Scan(&et.Name, &description, &et.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning entity type: %w", err)
	}

	et.Description = description.String
	return &et, nil
}

// ListEntityTypes lists all custom entity types.
func (r *Repository) ListEntityTypes(ctx context.Context) ([]entities.EntityType, error) {
	query := `
		SELECT name, description, created_at
		FROM entity_types
		ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying entity types: %w", err)
	}
	defer rows.Close()

	entityTypes := make([]entities.EntityType, 0, 16)
	for rows.Next() {
		var et entities.EntityType
		var description sql.NullString

		if err := rows.Scan(&et.Name, &description, &et.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning entity type: %w", err)
		}
		et.Description = description.String
		entityTypes = append(entityTypes, et)
	}
	return entityTypes, rows.Err()
}

// DeleteEntityType deletes a custom entity type by name.
func (r *Repository) DeleteEntityType(ctx context.Context, name string) error {
	query := `DELETE FROM entity_types WHERE name = ?`
	result, err := r.db.ExecContext(ctx, query, name)
	if err != nil {
		return fmt.Errorf("deleting entity type: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("entity type not found: %s", name)
	}
	return nil
}

// LogAction logs an action to the audit log.
func (r *Repository) LogAction(ctx context.Context, action string, factID string, details map[string]any) error {
	var detailsJSON sql.NullString
	if details != nil {
		data, err := json.Marshal(details)
		if err != nil {
			return fmt.Errorf("marshaling details: %w", err)
		}
		detailsJSON = sql.NullString{String: string(data), Valid: true}
	}

	var factIDPtr sql.NullString
	if factID != "" {
		factIDPtr = sql.NullString{String: factID, Valid: true}
	}

	query := `INSERT INTO audit_log (action, fact_id, details) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, action, factIDPtr, detailsJSON)
	if err != nil {
		return fmt.Errorf("logging action: %w", err)
	}
	return nil
}

// FindAuditLog finds audit log entries for a specific fact.
func (r *Repository) FindAuditLog(ctx context.Context, factID string) ([]entities.AuditEntry, error) {
	query := `
		SELECT id, action, fact_id, details, created_at
		FROM audit_log
		WHERE fact_id = ?
		ORDER BY created_at DESC
	`
	return r.queryAuditLog(ctx, query, factID)
}

// FindAuditLogByAction finds audit log entries by action type.
func (r *Repository) FindAuditLogByAction(ctx context.Context, action string, limit int) ([]entities.AuditEntry, error) {
	query := `
		SELECT id, action, fact_id, details, created_at
		FROM audit_log
		WHERE action = ?
		ORDER BY created_at DESC
		LIMIT ?
	`
	return r.queryAuditLog(ctx, query, action, limit)
}

// queryAuditLog is a helper to execute audit log queries.
func (r *Repository) queryAuditLog(ctx context.Context, query string, args ...any) ([]entities.AuditEntry, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying audit log: %w", err)
	}
	defer rows.Close()

	// Use limit parameter as capacity hint if available
	var entries []entities.AuditEntry
	if len(args) > 0 {
		if limit, ok := args[len(args)-1].(int); ok && limit > 0 {
			entries = make([]entities.AuditEntry, 0, limit)
		}
	}

	for rows.Next() {
		var entry entities.AuditEntry
		var factID, details sql.NullString

		if err := rows.Scan(
			&entry.ID,
			&entry.Action,
			&factID,
			&details,
			&entry.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning audit entry: %w", err)
		}

		entry.FactID = factID.String

		if details.Valid && details.String != "" {
			if err := json.Unmarshal([]byte(details.String), &entry.Details); err != nil {
				return nil, fmt.Errorf("unmarshaling details: %w", err)
			}
		}

		entries = append(entries, entry)
	}
	return entries, rows.Err()
}
