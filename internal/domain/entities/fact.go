// Package entities contains core domain data structures.
package entities

import "time"

// FactType represents the category of a fact.
// Validation of fact types is now handled by EntityTypeService, which supports
// both built-in and custom user-defined types.
type FactType string

// Default fact types. Custom types can be added via EntityTypeService.
const (
	FactTypeCharacter    FactType = "character"
	FactTypeLocation     FactType = "location"
	FactTypeEvent        FactType = "event"
	FactTypeRelationship FactType = "relationship"
	FactTypeRule         FactType = "rule"
	FactTypeTimeline     FactType = "timeline"
)

// Fact represents a single piece of factual information about a fictional world.
type Fact struct {
	ID         string    `json:"id"`
	Type       FactType  `json:"type"`
	Subject    string    `json:"subject"`
	Predicate  string    `json:"predicate"`
	Object     string    `json:"object"`
	Context    string    `json:"context"`
	SourceFile string    `json:"source_file"`
	SourceLine int       `json:"source_line"`
	Confidence float64   `json:"confidence"`
	Embedding  []float32 `json:"embedding,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

