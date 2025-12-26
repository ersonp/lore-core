package parsers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONParser_Parse_ValidInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []RawFact
	}{
		{
			name:  "single fact",
			input: `[{"type": "character", "subject": "Gandalf", "predicate": "is a", "object": "wizard"}]`,
			expected: []RawFact{
				{Type: "character", Subject: "Gandalf", Predicate: "is a", Object: "wizard"},
			},
		},
		{
			name:     "empty array",
			input:    "[]",
			expected: []RawFact{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &JSONParser{}
			result, err := parser.Parse(strings.NewReader(tt.input))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJSONParser_Parse_AllFields(t *testing.T) {
	input := `[{
		"id": "fact-1",
		"type": "location",
		"subject": "Mordor",
		"predicate": "is located in",
		"object": "Middle Earth",
		"context": "The land of shadow",
		"source_file": "lotr.txt",
		"confidence": 0.95
	}]`

	parser := &JSONParser{}
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, result, 1)

	fact := result[0]
	assert.Equal(t, "fact-1", fact.ID)
	assert.Equal(t, "location", fact.Type)
	assert.Equal(t, "Mordor", fact.Subject)
	assert.Equal(t, "Middle Earth", fact.Object)
	assert.Equal(t, "The land of shadow", fact.Context)
	assert.Equal(t, "lotr.txt", fact.SourceFile)
	assert.Equal(t, 0.95, fact.Confidence)
}

func TestJSONParser_Parse_InvalidInput(t *testing.T) {
	parser := &JSONParser{}
	_, err := parser.Parse(strings.NewReader("not json"))
	require.Error(t, err)
}

func TestCSVParser_Parse_ValidInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []RawFact
	}{
		{
			name:  "required columns only",
			input: "type,subject,predicate,object\ncharacter,Gandalf,is a,wizard\n",
			expected: []RawFact{
				{Type: "character", Subject: "Gandalf", Predicate: "is a", Object: "wizard"},
			},
		},
		{
			name:     "empty CSV (header only)",
			input:    "type,subject,predicate,object\n",
			expected: nil,
		},
		{
			name:  "columns in different order",
			input: "object,predicate,subject,type\nwizard,is a,Gandalf,character\n",
			expected: []RawFact{
				{Type: "character", Subject: "Gandalf", Predicate: "is a", Object: "wizard"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &CSVParser{}
			result, err := parser.Parse(strings.NewReader(tt.input))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCSVParser_Parse_AllColumns(t *testing.T) {
	input := "id,type,subject,predicate,object,context,source_file,confidence\n" +
		"fact-1,location,Mordor,is in,Middle Earth,Dark land,lotr.txt,0.95\n"

	parser := &CSVParser{}
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, result, 1)

	fact := result[0]
	assert.Equal(t, "fact-1", fact.ID)
	assert.Equal(t, "location", fact.Type)
	assert.Equal(t, "Mordor", fact.Subject)
	assert.Equal(t, "Middle Earth", fact.Object)
	assert.Equal(t, "Dark land", fact.Context)
	assert.Equal(t, "lotr.txt", fact.SourceFile)
	assert.Equal(t, 0.95, fact.Confidence)
}

func TestCSVParser_Parse_Errors(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		errMsg string
	}{
		{
			name:   "missing required column",
			input:  "type,subject,predicate\ncharacter,Gandalf,is a\n",
			errMsg: "missing required column: object",
		},
		{
			name:   "invalid confidence value",
			input:  "type,subject,predicate,object,confidence\ncharacter,Gandalf,is a,wizard,invalid\n",
			errMsg: "invalid confidence value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &CSVParser{}
			_, err := parser.Parse(strings.NewReader(tt.input))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestForFormat(t *testing.T) {
	assert.IsType(t, &JSONParser{}, ForFormat("json"))
	assert.IsType(t, &CSVParser{}, ForFormat("csv"))
	assert.Nil(t, ForFormat("unknown"))
}

func TestForFile(t *testing.T) {
	assert.IsType(t, &JSONParser{}, ForFile("facts.json"))
	assert.IsType(t, &CSVParser{}, ForFile("data.csv"))
	assert.Nil(t, ForFile("file.txt"))
	assert.Nil(t, ForFile("noextension"))
}
