package entities

import (
	"strings"
	"time"
)

// Entity represents a named subject (character, location, etc.) that can
// participate in relationships. Unlike EntityType which categorizes facts,
// Entity represents the actual subjects like "Alice" or "Northern Kingdom".
type Entity struct {
	ID             string    `json:"id"`
	WorldID        string    `json:"world_id"`
	Name           string    `json:"name"`            // Original name (e.g., "Alice")
	NormalizedName string    `json:"normalized_name"` // Lowercase for matching (e.g., "alice")
	CreatedAt      time.Time `json:"created_at"`
}

// NormalizeName converts a name to lowercase for case-insensitive matching.
func NormalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
