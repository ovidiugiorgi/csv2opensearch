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

// NewBatchProcessor creates a processor that ships batches of records from source to sink.
func NewBatchProcessor(size int, source Source, sink Sink) *BatchProcessor {
	return &BatchProcessor{
		size:   size,
		source: source,
		sink:   sink,
	}
}

// Run the import. The method will automatically exit once the context is cancelled.
func (bp *BatchProcessor) Run(ctx context.Context) {
	records := make([]string, 0, bp.size)
	var off, batchNum int

	for {
		select {
		case <-ctx.Done():
			log.Printf("Context is cancelled, flushing partial batch of %d records\n", len(records))

			err := bp.sink.Write(ctx, records)
			if err != nil {
				log.Fatalf("Failed to flush partial batch: %v", err)
				return
			}
			return
		default:
			if off == bp.size {
				batchNum++
				log.Printf("Flushing batch #%d with %d records\n", batchNum, len(records))

				err := bp.sink.Write(ctx, records)
				if err != nil {
					log.Fatalf("Failed to flush batch #%d: %v", batchNum, err)
					return
				}

				// Reset batch
				records = make([]string, 0, bp.size)
				off = 0
			}

			r, err := bp.source.Read(ctx)
			if err != nil {
				if errors.Is(err, io.EOF) {
					batchNum++
					log.Printf("Flushing final batch with %d records\n", len(records))

					err = bp.sink.Write(ctx, records)
					if err != nil {
						log.Fatalf("Failed to flush final batch: %v", err)
						return
					}
					return
				}

				log.Fatalf("Failed to read records: %v", err)
				return
			}

			records = append(records, r)
			off++
		}
	}
}
