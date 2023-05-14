package csv2opensearch

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Reader reads the CSV and returns JSON stringified versions for each record.
type Reader struct {
	reader  *csv.Reader
	headers []string
}

// NewReader returns a new CSV reader pointing to the file path.
func NewReader(path string) (*Reader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV: %v", err)
	}

	r := csv.NewReader(f)
	r.LazyQuotes = true

	return &Reader{reader: r}, nil
}

// Read returns a JSON serialized record from the CSV file and then advances the offset.
func (ci *Reader) Read(_ context.Context) (string, error) {
	if len(ci.headers) == 0 { // lazy load headers
		h, err := ci.reader.Read()
		if err != nil {
			return "", fmt.Errorf("failed to read CSV headers: %v", err)
		}
		if len(h) == 0 {
			return "", errors.New("missing headers")
		}
		ci.headers = h
	}

	row, err := ci.reader.Read()
	if err != nil {
		return "", fmt.Errorf("failed to read row: %w", err)
	}

	if len(row) == 0 {
		return "", nil
	}

	rs := ci.jsonify(row)

	return rs, nil
}

func (ci *Reader) jsonify(row []string) string {
	b := strings.Builder{}
	b.WriteString("{")

	for i := range ci.headers {
		b.WriteString(fmt.Sprintf("\"%s\":", ci.headers[i])) // Key

		// Sanitize value
		v := row[i]
		v = strings.ReplaceAll(v, "\n", "")
		v = strings.ReplaceAll(v, "\"", "")
		v = strings.ReplaceAll(v, "\\", "")

		b.WriteString(fmt.Sprintf("\"%s\"", v)) // Property

		if i < len(ci.headers)-1 {
			b.WriteString(",")
		}
	}

	b.WriteString("}")

	return b.String()
}
