package csv2opensearch_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	csv2opensearch "github.com/ovidiugiorgi/csv2opensearch"
)

func newMockOpenSearchServer(t *testing.T, bulkStatus int, onBulk func(string)) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
  "name":"node-1",
  "cluster_name":"mock-cluster",
  "cluster_uuid":"abc123",
  "version":{
    "distribution":"opensearch",
    "number":"3.5.0",
    "build_type":"tar",
    "build_hash":"hash",
    "build_date":"2026-01-01T00:00:00Z",
    "build_snapshot":false,
    "lucene_version":"10.1.0",
    "minimum_wire_compatibility_version":"7.10.0",
    "minimum_index_compatibility_version":"7.0.0"
  },
  "tagline":"The OpenSearch Project: https://opensearch.org/"
}`))
			return
		case "/_bulk":
			defer r.Body.Close()
			b, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read bulk request body: %v", err)
			}
			if onBulk != nil {
				onBulk(string(b))
			}

			w.WriteHeader(bulkStatus)
			if bulkStatus == http.StatusOK {
				_, _ = w.Write([]byte(`{"took":1,"errors":false,"items":[]}`))
			} else {
				_, _ = w.Write([]byte(`{"error":"boom","status":400}`))
			}
			return
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestWriterWrite_AllEmptyDocsIsNoop(t *testing.T) {
	requests := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	w, err := csv2opensearch.NewWriter(ts.URL, "cars")
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	if err := w.Write(context.Background(), []string{"", "   "}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requests != 0 {
		t.Fatalf("expected no bulk request, got %d", requests)
	}
}

func TestWriterWrite_SkipsEmptyDocsInBulkPayload(t *testing.T) {
	bulkCalls := 0
	var gotBody string
	ts := newMockOpenSearchServer(t, http.StatusOK, func(body string) {
		bulkCalls++
		gotBody = body
	})
	defer ts.Close()

	w, err := csv2opensearch.NewWriter(ts.URL, "cars")
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	if err := w.Write(context.Background(), []string{
		"",
		"   ",
		`{"id":"1"}`,
		`{"id":"2"}`,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bulkCalls != 1 {
		t.Fatalf("expected 1 bulk request, got %d", bulkCalls)
	}
	if strings.Count(gotBody, `{"index": {"_index": "cars"}}`) != 2 {
		t.Fatalf("expected 2 index actions in body, got %q", gotBody)
	}
	if !strings.Contains(gotBody, `{"id":"1"}`) || !strings.Contains(gotBody, `{"id":"2"}`) {
		t.Fatalf("missing documents in payload: %q", gotBody)
	}
}

func TestWriterWrite_Non200ResponseReturnsError(t *testing.T) {
	ts := newMockOpenSearchServer(t, http.StatusBadRequest, nil)
	defer ts.Close()

	w, err := csv2opensearch.NewWriter(ts.URL, "cars")
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	err = w.Write(context.Background(), []string{`{"id":"1"}`})
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
	if !strings.Contains(err.Error(), "received non-200 status code") {
		t.Fatalf("unexpected error: %v", err)
	}
}
