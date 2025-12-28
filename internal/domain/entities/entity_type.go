package entities

import "time"

// EntityType represents a custom entity type defined by the user.
// These extend the built-in FactType values.
type EntityType struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
