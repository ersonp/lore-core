package entities

import "time"

// ChangeType indicates why a fact was changed.
type ChangeType string

const (
	ChangeCreation   ChangeType = "creation"
	ChangeCorrection ChangeType = "correction"
	ChangeRetcon     ChangeType = "retcon"
	ChangeUpdate     ChangeType = "update"
	ChangeDeletion   ChangeType = "deletion"
)

// FactVersion represents a historical snapshot of a fact.
type FactVersion struct {
	ID         string     `json:"id"`
	FactID     string     `json:"fact_id"`
	Version    int        `json:"version"`
	ChangeType ChangeType `json:"change_type"`
	Data       Fact       `json:"data"`
	Reason     string     `json:"reason"`
	CreatedAt  time.Time  `json:"created_at"`
}
