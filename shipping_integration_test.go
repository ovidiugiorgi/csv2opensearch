package csv2opensearch_test

import (
	"context"
	"encoding/csv"
	"os"
	"strings"
	"testing"

	csv2opensearch "github.com/ovidiugiorgi/csv2opensearch"
	"golang.org/x/time/rate"
)

func TestShipping_ExactMultipleOfBatchSize_DoesNotSendEmptyFinalBulk(t *testing.T) {
	const (
		index     = "cars"
		batchSize = 5
	)

	var actionCounts []int
	ts := newMockOpenSearchServer(t, 200, func(body string) {
		actionCounts = append(actionCounts, countActions(body, index))
	})
	defer ts.Close()

	runShippingTest(t, "testdata/used_cars_integration_exact_20.csv", ts.URL, index, batchSize)

	if len(actionCounts) != 4 {
		t.Fatalf("expected 4 bulk calls, got %d", len(actionCounts))
	}

	for i, actions := range actionCounts {
		if actions != batchSize {
			t.Fatalf("bulk call #%d has %d actions, expected %d", i+1, actions, batchSize)
		}
	}
}

func TestShipping_NonMultipleOfBatchSize_SendsPartialFinalBulk(t *testing.T) {
	const (
		index     = "cars"
		batchSize = 5
	)

	var actionCounts []int
	ts := newMockOpenSearchServer(t, 200, func(body string) {
		actionCounts = append(actionCounts, countActions(body, index))
	})
	defer ts.Close()

	runShippingTest(t, "testdata/used_cars_integration_partial_23.csv", ts.URL, index, batchSize)

	if len(actionCounts) != 5 {
		t.Fatalf("expected 5 bulk calls, got %d", len(actionCounts))
	}

	expectedActions := []int{5, 5, 5, 5, 3}
	for i, actions := range actionCounts {
		if actions != expectedActions[i] {
			t.Fatalf("bulk call #%d has %d actions, expected %d", i+1, actions, expectedActions[i])
		}
	}
}

func runShippingTest(t *testing.T, csvPath, host, index string, batchSize int) {
	t.Helper()

	f, err := os.Open(csvPath)
	if err != nil {
		t.Fatalf("failed to open CSV fixture %q: %v", csvPath, err)
	}
	defer f.Close()

	reader := csv2opensearch.NewReader(csv.NewReader(f))
	writer, err := csv2opensearch.NewWriter(host, index)
	if err != nil {
		t.Fatalf("failed to create writer: %v", err)
	}

	proc := csv2opensearch.NewBatchProcessor(
		batchSize,
		reader,
		csv2opensearch.NewRateLimiter(rate.Inf, batchSize, writer),
	)
	proc.Run(context.Background())
}

func countActions(bulkBody, index string) int {
	return strings.Count(bulkBody, `{"index": {"_index": "`+index+`"}}`)
}
