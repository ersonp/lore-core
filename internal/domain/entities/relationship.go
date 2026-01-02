package entities

import "time"

// RelationType defines the kind of relationship between entities.
type RelationType string

const (
	RelationParent    RelationType = "parent"
	RelationChild     RelationType = "child"
	RelationSibling   RelationType = "sibling"
	RelationSpouse    RelationType = "spouse"
	RelationAlly      RelationType = "ally"
	RelationEnemy     RelationType = "enemy"
	RelationLocatedIn RelationType = "located_in"
	RelationOwns      RelationType = "owns"
	RelationMemberOf  RelationType = "member_of"
	RelationCreated   RelationType = "created"
)

// Relationship represents a directed connection between two entities.
type Relationship struct {
	ID             string       `json:"id"`
	SourceEntityID string       `json:"source_entity_id"`
	TargetEntityID string       `json:"target_entity_id"`
	Type           RelationType `json:"type"`
	Bidirectional  bool         `json:"bidirectional"`
	CreatedAt      time.Time    `json:"created_at"`
}
