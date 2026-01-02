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
	service  *services.RelationshipService
	vectorDB ports.VectorDB
}

// NewRelationshipHandler creates a new RelationshipHandler.
func NewRelationshipHandler(service *services.RelationshipService, vectorDB ports.VectorDB) *RelationshipHandler {
	return &RelationshipHandler{
		service:  service,
		vectorDB: vectorDB,
	}
}

// ListOptions configures relationship listing behavior.
type ListOptions struct {
	Type  string // Filter by relationship type (empty = all)
	Depth int    // Graph traversal depth (default 1)
}

// RelationshipInfo contains a relationship with optional fact details.
type RelationshipInfo struct {
	Relationship entities.Relationship `json:"relationship"`
	SourceFact   *entities.Fact        `json:"source_fact,omitempty"`
	TargetFact   *entities.Fact        `json:"target_fact,omitempty"`
}

// ListResult contains the result of listing relationships.
type ListResult struct {
	Relationships []RelationshipInfo `json:"relationships"`
}

// HandleCreate creates a new relationship between two facts.
func (h *RelationshipHandler) HandleCreate(
	ctx context.Context,
	sourceID string,
	relType string,
	targetID string,
	bidirectional bool,
) (*entities.Relationship, error) {
	// Validate relationship type
	rt, err := parseRelationType(relType)
	if err != nil {
		return nil, err
	}

	return h.service.Create(ctx, sourceID, rt, targetID, bidirectional)
}

// HandleDelete removes a relationship by ID.
func (h *RelationshipHandler) HandleDelete(ctx context.Context, id string) error {
	return h.service.Delete(ctx, id)
}

// HandleList returns relationships for a fact with optional filtering.
func (h *RelationshipHandler) HandleList(ctx context.Context, factID string, opts ListOptions) (*ListResult, error) {
	// Get relationships
	relationships, err := h.service.List(ctx, factID)
	if err != nil {
		return nil, fmt.Errorf("listing relationships: %w", err)
	}

	// Filter by type if specified
	if opts.Type != "" {
		filtered := make([]entities.Relationship, 0, len(relationships))
		for i := range relationships {
			if string(relationships[i].Type) == opts.Type {
				filtered = append(filtered, relationships[i])
			}
		}
		relationships = filtered
	}

	// Build result with fact details
	result := &ListResult{
		Relationships: make([]RelationshipInfo, 0, len(relationships)),
	}

	// Collect unique fact IDs to fetch
	factIDs := make(map[string]bool)
	for i := range relationships {
		factIDs[relationships[i].SourceFactID] = true
		factIDs[relationships[i].TargetFactID] = true
	}

	// Fetch facts
	ids := make([]string, 0, len(factIDs))
	for id := range factIDs {
		ids = append(ids, id)
	}

	facts, err := h.vectorDB.FindByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("fetching facts: %w", err)
	}

	// Build fact lookup map
	factMap := make(map[string]*entities.Fact, len(facts))
	for i := range facts {
		factMap[facts[i].ID] = &facts[i]
	}

	// Build relationship info
	for i := range relationships {
		info := RelationshipInfo{
			Relationship: relationships[i],
			SourceFact:   factMap[relationships[i].SourceFactID],
			TargetFact:   factMap[relationships[i].TargetFactID],
		}
		result.Relationships = append(result.Relationships, info)
	}

	return result, nil
}

// HandleFindBetween finds a direct relationship between two facts.
func (h *RelationshipHandler) HandleFindBetween(ctx context.Context, sourceID, targetID string) (*entities.Relationship, error) {
	return h.service.FindBetween(ctx, sourceID, targetID)
}

// HandleCount returns the total number of relationships.
func (h *RelationshipHandler) HandleCount(ctx context.Context) (int, error) {
	return h.service.Count(ctx)
}

// parseRelationType validates and converts a string to RelationType.
func parseRelationType(s string) (entities.RelationType, error) {
	switch s {
	case "parent":
		return entities.RelationParent, nil
	case "child":
		return entities.RelationChild, nil
	case "sibling":
		return entities.RelationSibling, nil
	case "spouse":
		return entities.RelationSpouse, nil
	case "ally":
		return entities.RelationAlly, nil
	case "enemy":
		return entities.RelationEnemy, nil
	case "located_in":
		return entities.RelationLocatedIn, nil
	case "owns":
		return entities.RelationOwns, nil
	case "member_of":
		return entities.RelationMemberOf, nil
	case "created":
		return entities.RelationCreated, nil
	default:
		return "", fmt.Errorf("invalid relationship type: %s (valid: parent, child, sibling, spouse, ally, enemy, located_in, owns, member_of, created)", s)
	}
}
