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

// DefaultTypeNames returns just the names of default types for quick lookup.
func DefaultTypeNames() []string {
	names := make([]string, len(DefaultEntityTypes))
	for i, t := range DefaultEntityTypes {
		names[i] = t.Name
	}
	return names
}

// IsDefaultType checks if a type name is a built-in default.
func IsDefaultType(name string) bool {
	for _, t := range DefaultEntityTypes {
		if t.Name == name {
			return true
		}
	}
	return false
}
