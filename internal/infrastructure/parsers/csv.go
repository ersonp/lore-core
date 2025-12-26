package parsers

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

// CSVParser parses facts from CSV format.
type CSVParser struct{}

// Parse reads CSV from the reader and returns parsed facts.
// Expected columns: type, subject, predicate, object, context, source_file, confidence
func (p *CSVParser) Parse(r io.Reader) ([]RawFact, error) {
	reader := csv.NewReader(r)

	colIndex, err := p.readHeader(reader)
	if err != nil {
		return nil, err
	}

	return p.readRecords(reader, colIndex)
}

// readHeader reads and validates the CSV header row.
func (p *CSVParser) readHeader(reader *csv.Reader) (map[string]int, error) {
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("reading CSV header: %w", err)
	}

	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[col] = i
	}

	requiredCols := []string{"type", "subject", "predicate", "object"}
	for _, col := range requiredCols {
		if _, ok := colIndex[col]; !ok {
			return nil, fmt.Errorf("missing required column: %s", col)
		}
	}

	return colIndex, nil
}

// readRecords reads all data rows and converts them to RawFacts.
func (p *CSVParser) readRecords(reader *csv.Reader, colIndex map[string]int) ([]RawFact, error) {
	var facts []RawFact
	lineNum := 1 // Header is line 1

	for {
		lineNum++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		fact, err := p.parseRecord(record, colIndex, lineNum)
		if err != nil {
			return nil, err
		}
		facts = append(facts, fact)
	}

	return facts, nil
}

// parseRecord converts a CSV record to a RawFact.
func (p *CSVParser) parseRecord(record []string, colIndex map[string]int, lineNum int) (RawFact, error) {
	fact := RawFact{
		Type:       getColumn(record, colIndex, "type"),
		Subject:    getColumn(record, colIndex, "subject"),
		Predicate:  getColumn(record, colIndex, "predicate"),
		Object:     getColumn(record, colIndex, "object"),
		Context:    getColumn(record, colIndex, "context"),
		ID:         getColumn(record, colIndex, "id"),
		SourceFile: getColumn(record, colIndex, "source_file"),
	}

	confStr := getColumn(record, colIndex, "confidence")
	if confStr != "" {
		conf, err := strconv.ParseFloat(confStr, 64)
		if err != nil {
			return RawFact{}, fmt.Errorf("line %d: invalid confidence value %q: %w", lineNum, confStr, err)
		}
		fact.Confidence = conf
	}

	return fact, nil
}

// getColumn safely retrieves a column value from a record.
func getColumn(record []string, colIndex map[string]int, col string) string {
	if idx, ok := colIndex[col]; ok && idx < len(record) {
		return record[idx]
	}
	return ""
}
