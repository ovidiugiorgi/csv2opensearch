package csv2opensearch

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"
)

// Reader reads the CSV and returns JSON stringified versions for each record.
type Reader struct {
	reader  *csv.Reader
	headers []string
}

// NewReader returns a new Reader that maps CSV records to JSONs.
func NewReader(reader *csv.Reader) *Reader {
	return &Reader{reader: reader}
}

// Read returns a JSON serialized record from the CSV file and then advances the offset.
func (r *Reader) Read(_ context.Context) (string, error) {
	if len(r.headers) == 0 { // lazy load headers
		h, err := r.reader.Read()
		if err != nil {
			return "", fmt.Errorf("failed to read CSV headers: %v", err)
		}
		if len(h) == 0 {
			return "", errors.New("missing headers")
		}
		r.headers = h
	}

	row, err := r.reader.Read()
	if err != nil {
		return "", fmt.Errorf("failed to read row: %w", err)
	}

	if len(row) == 0 {
		return "", nil
	}

	rs := r.jsonify(row)

	return rs, nil
}

func (r *Reader) jsonify(row []string) string {
	b := strings.Builder{}
	b.WriteString("{")

	for i := range r.headers {
		// Set the key
		b.WriteString(fmt.Sprintf("\"%s\":", r.headers[i]))

		// Sanitize the value
		v := row[i]
		v = strings.ReplaceAll(v, "\n", "")
		v = strings.ReplaceAll(v, "\"", "")
		v = strings.ReplaceAll(v, "\\", "")

		// Set the value
		b.WriteString(fmt.Sprintf("\"%s\"", v))

		if i < len(r.headers)-1 {
			b.WriteString(",")
		}
	}

	b.WriteString("}")

	return b.String()
}
