package entities

// DefaultEntityTypes are the built-in entity types seeded on world creation.
// These cannot be deleted by users.
var DefaultEntityTypes = []EntityType{
	{
		Name:        "character",
		Description: "People, beings, named entities in the world",
	},
	{
		Name:        "location",
		Description: "Places, regions, buildings, geographical features",
	},
	{
		Name:        "event",
		Description: "Historical events, battles, ceremonies, occurrences",
	},
	{
		Name:        "relationship",
		Description: "Connections between entities (ally, enemy, family)",
	},
	{
		Name:        "rule",
		Description: "Laws, customs, magic rules, world mechanics",
	},
	{
		Name:        "timeline",
		Description: "Temporal facts, dates, sequences, eras",
	},
}

// defaultTypeNames is pre-computed at package init for O(1) access.
// Treat as read-only.
var defaultTypeNames = func() []string {
	names := make([]string, len(DefaultEntityTypes))
	for i, t := range DefaultEntityTypes {
		names[i] = t.Name
	}
	return names
}()

// defaultTypeSet is pre-computed at package init for O(1) lookup.
var defaultTypeSet = func() map[string]bool {
	m := make(map[string]bool, len(DefaultEntityTypes))
	for _, t := range DefaultEntityTypes {
		m[t.Name] = true
	}
	return m
}()

// DefaultTypeNames returns the names of default types.
// The returned slice is shared and must not be modified by callers.
func DefaultTypeNames() []string {
	return defaultTypeNames
}

// IsDefaultType checks if a type name is a built-in default. O(1) lookup.
func IsDefaultType(name string) bool {
	return defaultTypeSet[name]
}
