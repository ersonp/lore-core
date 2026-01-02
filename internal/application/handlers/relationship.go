package handlers

import (
	"context"
	"fmt"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/ports"
	"github.com/ersonp/lore-core/internal/domain/services"
)

// ValidRelationTypes lists all valid relationship type strings.
var ValidRelationTypes = []string{
	"parent", "child", "sibling", "spouse",
	"ally", "enemy",
	"located_in", "owns", "member_of", "created",
}

// RelationshipHandler handles relationship operations.
type RelationshipHandler struct {
	service      *services.RelationshipService
	relationalDB ports.RelationalDB
}

// NewRelationshipHandler creates a new RelationshipHandler.
func NewRelationshipHandler(service *services.RelationshipService, relationalDB ports.RelationalDB) *RelationshipHandler {
	return &RelationshipHandler{
		service:      service,
		relationalDB: relationalDB,
	}
}

// ListOptions configures relationship listing behavior.
type ListOptions struct {
	Type  string // Filter by relationship type (empty = all)
	Depth int    // Graph traversal depth (default 1)
}

// RelationshipInfo contains a relationship with entity details.
type RelationshipInfo struct {
	Relationship entities.Relationship `json:"relationship"`
	SourceEntity *entities.Entity      `json:"source_entity,omitempty"`
	TargetEntity *entities.Entity      `json:"target_entity,omitempty"`
}

// ListResult contains the result of listing relationships.
type ListResult struct {
	Relationships   []RelationshipInfo `json:"relationships"`
	RelatedEntities []string           `json:"related_entities,omitempty"`
}

// HandleCreate creates a new relationship between two entities.
func (h *RelationshipHandler) HandleCreate(
	ctx context.Context,
	worldID string,
	sourceEntityName string,
	relType string,
	targetEntityName string,
	bidirectional bool,
) (*entities.Relationship, error) {
	// Validate relationship type
	rt, err := parseRelationType(relType)
	if err != nil {
		return nil, err
	}

	return h.service.Create(ctx, worldID, sourceEntityName, rt, targetEntityName, bidirectional)
}

// HandleDelete removes a relationship by ID.
func (h *RelationshipHandler) HandleDelete(ctx context.Context, id string) error {
	return h.service.Delete(ctx, id)
}

// HandleList returns relationships for an entity with optional filtering.
func (h *RelationshipHandler) HandleList(ctx context.Context, worldID, entityName string, opts ListOptions) (*ListResult, error) {
	// Find the entity first to get its ID
	entity, err := h.relationalDB.FindEntityByName(ctx, worldID, entityName)
	if err != nil {
		return nil, fmt.Errorf("finding entity: %w", err)
	}
	if entity == nil {
		return &ListResult{Relationships: []RelationshipInfo{}}, nil
	}

	// Get relationships
	relationships, err := h.service.List(ctx, entity.ID)
	if err != nil {
		return nil, fmt.Errorf("listing relationships: %w", err)
	}

	// Get related entities if depth > 1
	relatedEntityNames, err := h.fetchRelatedEntityNames(ctx, entity.ID, opts.Depth)
	if err != nil {
		return nil, err
	}

	// Filter by type if specified
	relationships = filterByType(relationships, opts.Type)

	// Build entity lookup map for relationship info
	entityMap, err := h.buildEntityMap(ctx, relationships)
	if err != nil {
		return nil, err
	}

	// Build result with entity details
	return h.buildListResult(relationships, entityMap, relatedEntityNames), nil
}

// fetchRelatedEntityNames fetches names of entities connected at the given depth.
func (h *RelationshipHandler) fetchRelatedEntityNames(ctx context.Context, entityID string, depth int) ([]string, error) {
	if depth <= 1 {
		return nil, nil
	}

	relatedEntities, err := h.service.ListWithDepth(ctx, entityID, depth)
	if err != nil {
		return nil, fmt.Errorf("listing related entities: %w", err)
	}

	names := make([]string, 0, len(relatedEntities))
	for _, re := range relatedEntities {
		e, err := h.relationalDB.FindEntityByID(ctx, re.EntityID)
		if err != nil {
			return nil, fmt.Errorf("fetching entity %s: %w", re.EntityID, err)
		}
		if e != nil {
			names = append(names, e.Name)
		}
	}
	return names, nil
}

// filterByType filters relationships by type, returning all if typeFilter is empty.
func filterByType(relationships []entities.Relationship, typeFilter string) []entities.Relationship {
	if typeFilter == "" {
		return relationships
	}

	filtered := make([]entities.Relationship, 0, len(relationships))
	for i := range relationships {
		if string(relationships[i].Type) == typeFilter {
			filtered = append(filtered, relationships[i])
		}
	}
	return filtered
}

// buildEntityMap fetches all entities referenced in relationships and builds a lookup map.
func (h *RelationshipHandler) buildEntityMap(ctx context.Context, relationships []entities.Relationship) (map[string]*entities.Entity, error) {
	// Collect unique entity IDs
	entityIDSet := make(map[string]bool)
	for i := range relationships {
		entityIDSet[relationships[i].SourceEntityID] = true
		entityIDSet[relationships[i].TargetEntityID] = true
	}

	// Convert to slice for batch fetch
	entityIDs := make([]string, 0, len(entityIDSet))
	for id := range entityIDSet {
		entityIDs = append(entityIDs, id)
	}

	// Fetch entities in single query
	fetchedEntities, err := h.relationalDB.FindEntitiesByIDs(ctx, entityIDs)
	if err != nil {
		return nil, fmt.Errorf("fetching entities: %w", err)
	}

	// Build map for quick lookup
	entityMap := make(map[string]*entities.Entity, len(fetchedEntities))
	for _, entity := range fetchedEntities {
		entityMap[entity.ID] = entity
	}
	return entityMap, nil
}

// buildListResult constructs the final ListResult from relationships and entity data.
func (h *RelationshipHandler) buildListResult(relationships []entities.Relationship, entityMap map[string]*entities.Entity, relatedEntityNames []string) *ListResult {
	result := &ListResult{
		Relationships:   make([]RelationshipInfo, 0, len(relationships)),
		RelatedEntities: relatedEntityNames,
	}

	for i := range relationships {
		info := RelationshipInfo{
			Relationship: relationships[i],
			SourceEntity: entityMap[relationships[i].SourceEntityID],
			TargetEntity: entityMap[relationships[i].TargetEntityID],
		}
		result.Relationships = append(result.Relationships, info)
	}
	return result
}

// HandleFindBetween finds a direct relationship between two entities.
func (h *RelationshipHandler) HandleFindBetween(ctx context.Context, sourceEntityID, targetEntityID string) (*entities.Relationship, error) {
	return h.service.FindBetween(ctx, sourceEntityID, targetEntityID)
}

// HandleCount returns the total number of relationships.
func (h *RelationshipHandler) HandleCount(ctx context.Context) (int, error) {
	return h.service.Count(ctx)
}

// relationTypeMap provides O(1) lookup for relationship type validation.
var relationTypeMap = map[string]entities.RelationType{
	"parent":     entities.RelationParent,
	"child":      entities.RelationChild,
	"sibling":    entities.RelationSibling,
	"spouse":     entities.RelationSpouse,
	"ally":       entities.RelationAlly,
	"enemy":      entities.RelationEnemy,
	"located_in": entities.RelationLocatedIn,
	"owns":       entities.RelationOwns,
	"member_of":  entities.RelationMemberOf,
	"created":    entities.RelationCreated,
}

// parseRelationType validates and converts a string to RelationType.
func parseRelationType(s string) (entities.RelationType, error) {
	if rt, ok := relationTypeMap[s]; ok {
		return rt, nil
	}
	return "", fmt.Errorf("invalid relationship type: %s (valid: parent, child, sibling, spouse, ally, enemy, located_in, owns, member_of, created)", s)
}
