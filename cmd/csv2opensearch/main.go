package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/ovidiugiorgi/csv2opensearch"
	"golang.org/x/time/rate"
)

type config struct {
	csv, host, index string
	user, password   string
	batch, rate      int
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var (
		csv   = flag.String("csv", "", "Path to the CSV file")
		host  = flag.String("host", "https://localhost:9200", "URL of the OpenSearch cluster")
		index = flag.String("index", "", "Name for OpenSearch index where the data will end up. OpenSearch will automatically create the field mappings.")
		batch = flag.Int("batch", 100, "Number of records that will be indexed in a single Bulk API request")
		rate  = flag.Int("rate", -1, "Rate limit for number of records that are indexed per second")
	)
	flag.Parse()

	if *csv == "" {
		log.Fatalf("missing `csv` flag")
	}

	if *index == "" {
		log.Fatalf("missing `index` flag")
	}

	if err := run(ctx, config{
		csv:      *csv,
		host:     *host,
		user:     os.Getenv("OS_USER"),
		password: os.Getenv("OS_PASSWORD"),
		index:    *index,
		batch:    *batch,
		rate:     *rate,
	}); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, cfg config) error {
	reader, err := csv2opensearch.NewReader(cfg.csv)
	if err != nil {
		return fmt.Errorf("failed to create CSV reader: %v", err)
	}

	writer, err := csv2opensearch.NewWriter(
		cfg.host,
		cfg.index,
		csv2opensearch.WithBasicAuth(cfg.user, cfg.password))
	if err != nil {
		return fmt.Errorf("failed to create OpenSearch writer: %v", err)
	}

	var limit rate.Limit
	if cfg.rate > 0 {
		limit = rate.Limit(cfg.rate)
	} else {
		limit = rate.Inf
	}
	proc := csv2opensearch.NewBatchProcessor(
		cfg.batch,
		reader,
		csv2opensearch.NewRateLimiter(limit, cfg.batch, writer),
	)

	proc.Run(ctx)

	return nil
}
