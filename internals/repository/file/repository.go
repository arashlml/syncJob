package file

import (
	"encoding/json"
	"fmt"
	"os"
)

type JSONFileIterator struct {
	filePath string
	records  []map[string]interface{}
}

func NewJSONFileIterator(filePath string) (*JSONFileIterator, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %w", filePath, err)
	}
	defer file.Close()

	var records []map[string]interface{}
	if err := json.NewDecoder(file).Decode(&records); err != nil {
		return nil, fmt.Errorf("decoding json from %s: %w", filePath, err)
	}

	return &JSONFileIterator{
		filePath: filePath,
		records:  records,
	}, nil
}

func (it *JSONFileIterator) HasNext(cursor int) bool {
	return cursor < len(it.records)
}

func (it *JSONFileIterator) Next(cursor int) (map[string]interface{}, error) {
	if !it.HasNext(cursor) {
		return nil, fmt.Errorf("cursor %d out of range (total: %d)", cursor, len(it.records))
	}
	return it.records[cursor], nil
}
