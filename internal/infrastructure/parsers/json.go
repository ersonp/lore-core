package parsers

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONParser parses facts from JSON format.
type JSONParser struct{}

// Parse reads JSON from the reader and returns parsed facts.
func (p *JSONParser) Parse(r io.Reader) ([]RawFact, error) {
	var facts []RawFact

	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&facts); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	// Set line numbers (array index + 1, 1-indexed)
	for i := range facts {
		facts[i].LineNum = i + 1
	}

	return facts, nil
}
