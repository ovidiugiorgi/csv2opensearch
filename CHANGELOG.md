# Changelog

All notable changes to this project will be documented in this file.

## v0.1.0 (March 1, 2026)

First tagged release of `csv2opensearch`, focused on correctness, fast onboarding, and a tighter development loop.

```bash
go install github.com/ovidiugiorgi/csv2opensearch/cmd/csv2opensearch@v0.1.0
```

### Fixes

- Fixed #1: incorrect handling of an empty final bulk flush when CSV row count is an exact multiple of batch size.

### Local OpenSeach Stack

- Compose workflow supports both `nerdctl compose` and `docker compose`.
- OpenSearch and OpenSearch Dashboards images are pinned to `3.5.0` for reproducible local runs.
- Local sandbox stack runs without auth by default (HTTP, security disabled) for easier onboarding.
- CLI default host remains `https://localhost:9200` for compatibility with secured clusters, while local Make targets use `http://localhost:9200`.

### Demo Data / Quick Onboarding

- Added committed demo seed `testdata/used_cars_demo.csv` (1000 rows) so new users can ingest and query immediately.
- Added deterministic data generation via `scripts/gen_used_cars_dataset.py` and `make data`.
- Added integration fixtures for batch-boundary behavior: `testdata/used_cars_integration_exact_20.csv` and `testdata/used_cars_integration_partial_23.csv`.
- Quick onboarding path: `make up`, `make seed`, `make query`.

### Make-Powered Dev Workflow

- Stack lifecycle: `make up`, `make up-nowait`, `make status`, `make logs`, `make down`, `make restart`, `make pull`, `make delete-data`.
- Ingestion and querying: `make seed`, `make seed-dev`, `make indexes`, `make query`.
- Data generation and tests: `make data`, `make test`.

Examples:

```bash
make up
make seed
make query N=5 Q="Bucharest"
```

```bash
make data DATA_ROWS=200 DATA_SEED=7 DATA_OUT=testdata/custom.csv
make seed-dev SEED_CSV=testdata/custom.csv SEED_INDEX=custom-demo
make indexes
```

### Dev Loop / CI

- Added `make test` to run all Go tests.
- Added `N` support in `make query` to control result size.
- Added CI validation for `go test`, `go vet`, `gofmt`, and `goimports` on Go `1.20.x`.
- Added pre-commit checks including YAML validation, shellcheck, and markdown formatting with prettier.
