package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestConfigIndexName(t *testing.T) {
	t.Run("uses the provided name", func(t *testing.T) {
		var cfg config
		cfg.csv = "my_records.csv"
		cfg.index = "test"
		got := indexName(cfg)
		if got != cfg.index {
			t.Errorf(fmt.Sprintf("used index name %q, want %q", got, cfg.index))
		}
	})

	t.Run("creates a name based on the CSV file name and current timestamp", func(t *testing.T) {
		tests := []struct {
			fileName string
			prefix   string
		}{
			{
				fileName: "/Users/user/data/my_records.csv",
				prefix:   "my_records_",
			},
			{
				fileName: "my_records.csv",
				prefix:   "my_records_",
			},
			{
				fileName: "records",
				prefix:   "records_",
			},
		}

		for _, tt := range tests {
			t.Run(tt.fileName, func(t *testing.T) {
				var cfg config
				cfg.csv = tt.fileName
				got := indexName(cfg)
				if !strings.HasPrefix(got, tt.prefix) {
					t.Errorf(fmt.Sprintf("used index name %q does not start with %q", got, tt.prefix))
				}
			})
		}
	})
}

func TestRun_FailsWhenCSVPathDoesNotExist(t *testing.T) {
	err := run(context.Background(), config{
		csv:   "/definitely/missing/file.csv",
		batch: 100,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to open CSV file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRun_FailsWhenBatchIsNotPositive(t *testing.T) {
	err := run(context.Background(), config{
		csv:   "/definitely/missing/file.csv",
		batch: 0,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid batch size") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateConfig_RejectsNegativeBatch(t *testing.T) {
	err := validateConfig(config{batch: -1})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid batch size") {
		t.Fatalf("unexpected error: %v", err)
	}
}
