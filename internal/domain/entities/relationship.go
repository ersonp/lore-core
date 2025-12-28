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

// Relationship represents a directed connection between two facts.
type Relationship struct {
	ID            string       `json:"id"`
	SourceFactID  string       `json:"source_fact_id"`
	TargetFactID  string       `json:"target_fact_id"`
	Type          RelationType `json:"type"`
	Bidirectional bool         `json:"bidirectional"`
	CreatedAt     time.Time    `json:"created_at"`
}
