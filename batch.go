package csv2opensearch

import (
	"context"
	"errors"
	"io"
	"log"
)

// Source returns individual records.
// It should return io.EOF when there are no more records to read.
type Source interface {
	Read(context.Context) (string, error)
}

// Sink handles a batch of records.
type Sink interface {
	Write(context.Context, []string) error
}

// BatchProcessor polls the source for individual records and sends the records to the sink into batches.
type BatchProcessor struct {
	size   int
	source Source
	sink   Sink
}

// NewBatchProcessor creates a processor that ships batch of records from the source to the sink.
func NewBatchProcessor(size int, source Source, sink Sink) *BatchProcessor {
	return &BatchProcessor{
		size:   size,
		source: source,
		sink:   sink,
	}
}

// Run the import. The method will automatically exit once the provided context is cancelled.
func (bw *BatchProcessor) Run(ctx context.Context) {
	records := make([]string, 0, bw.size)
	var i, batches int

	for {
		select {
		case <-ctx.Done():
			log.Printf("context is cancelled, flushing in progress batch of %d records\n", len(records))

			err := bw.sink.Write(ctx, records)
			if err != nil {
				log.Fatalf("failed to flush in progress batch: %v", err)
				return
			}
			return
		default:
			if i == bw.size {
				batches++
				log.Printf("flushing full batch: %d\n", batches)

				err := bw.sink.Write(ctx, records)
				if err != nil {
					log.Fatalf("failed to flush full batch %d: %v", batches, err)
					return
				}

				// reset state
				records = make([]string, 0, bw.size)
				i = 0
			}

			r, err := bw.source.Read(ctx)
			if err != nil {
				if errors.Is(err, io.EOF) {
					batches++
					log.Printf("flushing final batch of %d records\n", len(records))

					err = bw.sink.Write(ctx, records)
					if err != nil {
						log.Fatalf("failed to flush final batch: %v", err)
						return
					}
					return
				}

				log.Fatalf("failed to read records: %v", err)
				return
			}

			records = append(records, r)
			i++
		}
	}
}
