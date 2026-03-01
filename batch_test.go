package csv2opensearch_test

import (
	"context"
	"io"
	"testing"

	csv2opensearch "github.com/ovidiugiorgi/csv2opensearch"
)

type sliceSource struct {
	records []string
	i       int
}

func (s *sliceSource) Read(context.Context) (string, error) {
	if s.i >= len(s.records) {
		return "", io.EOF
	}

	r := s.records[s.i]
	s.i++
	return r, nil
}

type recordingSink struct {
	writes [][]string
}

func (s *recordingSink) Write(_ context.Context, batch []string) error {
	cp := make([]string, len(batch))
	copy(cp, batch)
	s.writes = append(s.writes, cp)
	return nil
}

type cancelingSource struct {
	records []string
	i       int
	cancel  context.CancelFunc
}

func (s *cancelingSource) Read(context.Context) (string, error) {
	if s.i >= len(s.records) {
		return "", io.EOF
	}

	r := s.records[s.i]
	s.i++
	if s.i == 1 && s.cancel != nil {
		s.cancel()
	}
	return r, nil
}

func TestBatchProcessorRun_DoesNotFlushEmptyFinalBatch(t *testing.T) {
	src := &sliceSource{
		records: []string{"a", "b", "c", "d"},
	}
	sink := &recordingSink{}

	p := csv2opensearch.NewBatchProcessor(2, src, sink)
	p.Run(context.Background())

	if len(sink.writes) != 2 {
		t.Fatalf("expected 2 writes, got %d", len(sink.writes))
	}

	for i, batch := range sink.writes {
		if len(batch) == 0 {
			t.Fatalf("write #%d is empty", i+1)
		}
	}
}

func TestBatchProcessorRun_EmptySourceDoesNotWrite(t *testing.T) {
	src := &sliceSource{}
	sink := &recordingSink{}

	p := csv2opensearch.NewBatchProcessor(100, src, sink)
	p.Run(context.Background())

	if len(sink.writes) != 0 {
		t.Fatalf("expected 0 writes, got %d", len(sink.writes))
	}
}

func TestBatchProcessorRun_FlushesPartialFinalBatch(t *testing.T) {
	src := &sliceSource{
		records: []string{"a", "b", "c"},
	}
	sink := &recordingSink{}

	p := csv2opensearch.NewBatchProcessor(2, src, sink)
	p.Run(context.Background())

	if len(sink.writes) != 2 {
		t.Fatalf("expected 2 writes, got %d", len(sink.writes))
	}
	if len(sink.writes[0]) != 2 {
		t.Fatalf("expected first write size 2, got %d", len(sink.writes[0]))
	}
	if len(sink.writes[1]) != 1 {
		t.Fatalf("expected second write size 1, got %d", len(sink.writes[1]))
	}
}

func TestBatchProcessorRun_CancelledContextWithNoPendingRecordsDoesNotWrite(t *testing.T) {
	src := &sliceSource{records: []string{"a"}}
	sink := &recordingSink{}

	p := csv2opensearch.NewBatchProcessor(100, src, sink)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p.Run(ctx)

	if len(sink.writes) != 0 {
		t.Fatalf("expected 0 writes, got %d", len(sink.writes))
	}
}

func TestBatchProcessorRun_CancelledContextFlushesPendingRecords(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := &cancelingSource{
		records: []string{"a", "b"},
		cancel:  cancel,
	}
	sink := &recordingSink{}

	p := csv2opensearch.NewBatchProcessor(100, src, sink)
	p.Run(ctx)

	if len(sink.writes) != 1 {
		t.Fatalf("expected 1 write, got %d", len(sink.writes))
	}
	if len(sink.writes[0]) != 1 {
		t.Fatalf("expected 1 pending record flushed, got %d", len(sink.writes[0]))
	}
}
