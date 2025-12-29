package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFactType_IsValid tests the deprecated IsValid() method.
// Note: IsValid() only checks built-in types. For custom type validation,
// use EntityTypeService.IsValid() which supports dynamic custom types.
func TestFactType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		factType FactType
		expected bool
	}{
		{
			name:     "character is valid",
			factType: FactTypeCharacter,
			expected: true,
		},
		{
			name:     "location is valid",
			factType: FactTypeLocation,
			expected: true,
		},
		{
			name:     "event is valid",
			factType: FactTypeEvent,
			expected: true,
		},
		{
			name:     "relationship is valid",
			factType: FactTypeRelationship,
			expected: true,
		},
		{
			name:     "rule is valid",
			factType: FactTypeRule,
			expected: true,
		},
		{
			name:     "timeline is valid",
			factType: FactTypeTimeline,
			expected: true,
		},
		{
			name:     "empty string is invalid",
			factType: FactType(""),
			expected: false,
		},
		{
			name:     "unknown type is invalid",
			factType: FactType("unknown"),
			expected: false,
		},
		{
			name:     "misspelled type is invalid",
			factType: FactType("charactr"),
			expected: false,
		},
		{
			name:     "uppercase type is invalid",
			factType: FactType("CHARACTER"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.factType.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFactTypeConstants(t *testing.T) {
	// Verify constant values match expected strings
	assert.Equal(t, FactType("character"), FactTypeCharacter)
	assert.Equal(t, FactType("location"), FactTypeLocation)
	assert.Equal(t, FactType("event"), FactTypeEvent)
	assert.Equal(t, FactType("relationship"), FactTypeRelationship)
	assert.Equal(t, FactType("rule"), FactTypeRule)
	assert.Equal(t, FactType("timeline"), FactTypeTimeline)
}

// TestFactType_CustomTypesNotValidatedByIsValid documents that IsValid()
// returns false for custom types. This is expected behavior - use
// EntityTypeService.IsValid() for validating custom types.
func TestFactType_CustomTypesNotValidatedByIsValid(t *testing.T) {
	customTypes := []string{"weapon", "organization", "magic_system", "creature"}

	for _, typeName := range customTypes {
		customType := FactType(typeName)
		assert.False(t, customType.IsValid(),
			"IsValid() should return false for custom type %q; use EntityTypeService.IsValid()", typeName)
	}
}
