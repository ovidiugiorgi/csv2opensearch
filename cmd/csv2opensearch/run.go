package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ovidiugiorgi/csv2opensearch"
	"golang.org/x/time/rate"
)

func run(ctx context.Context, cfg config) error {
	// Setup reader
	f, err := os.Open(cfg.csv)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %v", err)
	}
	cr := csv.NewReader(f)
	cr.LazyQuotes = true
	reader := csv2opensearch.NewReader(cr)

	// Setup writer
	writer, err := csv2opensearch.NewWriter(
		cfg.host,
		indexName(cfg),
		csv2opensearch.WithBasicAuth(cfg.user, cfg.password))
	if err != nil {
		return fmt.Errorf("failed to create OpenSearch writer: %v", err)
	}

	// Setup batch processor
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

func indexName(cfg config) string {
	if cfg.index != "" {
		return cfg.index
	}

	s := filepath.Base(cfg.csv)
	s, _ = strings.CutSuffix(s, ".csv")
	return fmt.Sprintf("%s_%v", s, time.Now().Unix())
}
