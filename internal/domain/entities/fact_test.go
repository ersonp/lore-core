package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFactTypeConstants(t *testing.T) {
	// Verify constant values match expected strings
	assert.Equal(t, FactType("character"), FactTypeCharacter)
	assert.Equal(t, FactType("location"), FactTypeLocation)
	assert.Equal(t, FactType("event"), FactTypeEvent)
	assert.Equal(t, FactType("relationship"), FactTypeRelationship)
	assert.Equal(t, FactType("rule"), FactTypeRule)
	assert.Equal(t, FactType("timeline"), FactTypeTimeline)
}
