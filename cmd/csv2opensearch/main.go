package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
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
		csv = flag.String(
			"csv",
			"",
			"Path to the CSV file",
		)
		host = flag.String("host",
			"https://localhost:9200",
			"URL of the OpenSearch cluster",
		)
		index = flag.String("index",
			"",
			`Name of the OpenSearch index where the data will end up. If not provided, data will be
			written to an index based on the CSV file name and current timestamp.
			OpenSearch will automatically create the field mappings.`,
		)
		batch = flag.Int("batch",
			100,
			"Number of records that will be indexed in a single Bulk API request",
		)
		rate = flag.Int("rate",
			-1,
			"Rate limit for number of records that are indexed per second",
		)
	)
	flag.Parse()

	if *csv == "" {
		log.Fatalf("Missing `csv` flag")
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
