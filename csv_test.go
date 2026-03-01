package csv2opensearch_test

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"strings"
	"testing"

	csv2opensearch "github.com/ovidiugiorgi/csv2opensearch"
)

func TestReaderRead_ParsesAndSerializesRecord(t *testing.T) {
	cr := csv.NewReader(strings.NewReader("make,model\nFord,Fiesta\n"))
	r := csv2opensearch.NewReader(cr)

	got, err := r.Read(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `{"make":"Ford","model":"Fiesta"}`
	if got != want {
		t.Fatalf("unexpected JSON record: got %q want %q", got, want)
	}
}

func TestReaderRead_HeaderReadError(t *testing.T) {
	cr := csv.NewReader(strings.NewReader(""))
	r := csv2opensearch.NewReader(cr)

	_, err := r.Read(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read CSV headers") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReaderRead_RowReadErrorWrapsEOF(t *testing.T) {
	cr := csv.NewReader(strings.NewReader("make,model\n"))
	r := csv2opensearch.NewReader(cr)

	_, err := r.Read(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read row") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected wrapped io.EOF, got %v", err)
	}
}

func TestReaderRead_SanitizesQuotesBackslashesAndNewlines(t *testing.T) {
	cr := csv.NewReader(strings.NewReader("name,note\nAlice,\"hello\n\"\"quoted\"\"\\\\path\"\n"))
	r := csv2opensearch.NewReader(cr)

	got, err := r.Read(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `{"name":"Alice","note":"helloquotedpath"}`
	if got != want {
		t.Fatalf("unexpected JSON record: got %q want %q", got, want)
	}
}
