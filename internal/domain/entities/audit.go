package entities

import "time"

// AuditEntry represents a logged action in the system.
type AuditEntry struct {
	ID        int64          `json:"id"`
	Action    string         `json:"action"`
	FactID    string         `json:"fact_id,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}
