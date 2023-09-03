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
			"Optional. Path to the CSV file. If missing, the data is read from stdin.",
		)
		host = flag.String("host",
			"https://localhost:9200",
			"URL of the OpenSearch cluster",
		)
		index = flag.String("index",
			"",
			"Optional. Name of the OpenSearch index where the data will end up. "+
				"If not provided, data will be written to an index based on the CSV file name and current timestamp. "+
				" OpenSearch will automatically create the field mappings.",
		)
		batch = flag.Int("batch",
			100,
			"Number of records that will be indexed in a single Bulk API request",
		)
		rate = flag.Int("rate",
			-1,
			"Throttle the number of records that are indexed per second. By default, the ingestion is unthrottled.",
		)
	)
	flag.Parse()

	if *csv == "" && *index == "" {
		log.Printf("Index name is required when data is read from stdin.")
		flag.Usage()
		os.Exit(1)
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
