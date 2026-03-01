package csv2opensearch_test

import (
	"context"
	"testing"

	csv2opensearch "github.com/ovidiugiorgi/csv2opensearch"
	"golang.org/x/time/rate"
)

type countingSink struct {
	calls int
	last  []string
}

func (s *countingSink) Write(_ context.Context, v []string) error {
	s.calls++
	s.last = append([]string(nil), v...)
	return nil
}

func TestRateLimiterWrite_EmptyBatchNoop(t *testing.T) {
	sink := &countingSink{}
	rl := csv2opensearch.NewRateLimiter(rate.Inf, 1, sink)

	if err := rl.Write(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sink.calls != 0 {
		t.Fatalf("expected 0 sink calls, got %d", sink.calls)
	}
}

func TestRateLimiterWrite_ContextCancelledDoesNotCallSink(t *testing.T) {
	sink := &countingSink{}
	rl := csv2opensearch.NewRateLimiter(rate.Limit(1), 1, sink)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := rl.Write(ctx, []string{"a"}); err == nil {
		t.Fatal("expected an error for cancelled context")
	}
	if sink.calls != 0 {
		t.Fatalf("expected 0 sink calls, got %d", sink.calls)
	}
}

func TestRateLimiterWrite_NonEmptyBatchPassesThrough(t *testing.T) {
	sink := &countingSink{}
	rl := csv2opensearch.NewRateLimiter(rate.Inf, 10, sink)

	in := []string{"a", "b"}
	if err := rl.Write(context.Background(), in); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sink.calls != 1 {
		t.Fatalf("expected 1 sink call, got %d", sink.calls)
	}
	if len(sink.last) != len(in) {
		t.Fatalf("unexpected batch size: got %d want %d", len(sink.last), len(in))
	}
}
