# csv2opensearch

[![Go Reference](https://pkg.go.dev/badge/github.com/ovidiugiorgi/csv2opensearch.svg)](https://pkg.go.dev/github.com/ovidiugiorgi/csv2opensearch)
[![CI](https://github.com/ovidiugiorgi/csv2opensearch/actions/workflows/ci.yml/badge.svg)](https://github.com/ovidiugiorgi/csv2opensearch/actions/workflows/ci.yml)

Import CSV files into OpenSearch without needing to pre-configure the index mappings.

Use cases:
- data exploration
- load testing

Features & limitations:
- each CSV record is imported as a separate document
- document _properties_ are inferred from the CSV header (first record)
- document _values_ are imported as raw strings

## Install

```
go install github.com/ovidiugiorgi/csv2opensearch/cmd/csv2opensearch@latest
```

## Quickstart

```bash
make up
make seed
make query
```

By default, `make seed` ingests `testdata/used_cars_demo.csv` (1000 records) into the `used-cars-demo` index.

If you have not installed the `csv2opensearch` binary yet, use `make seed-dev` instead of `make seed`.

## Usage

### Auth (optional)

The provided `compose.yaml` disables security for local sandbox usage, so authentication is not required.

If you use a secured cluster, add `OS_USER` and `OS_PASSWORD` to the environment:

```
export OS_USER=your_user
export OS_PASSWORD=your_password
```

### Run the import

Read from file:
```
csv2opensearch --csv=test.csv --host=http://localhost:9200
```

Read from stdin:
```
cat test.csv | csv2opensearch --index=test
```

### Quick Demo Dataset

This repository includes a synthetic used-cars marketplace dataset:

- `testdata/used_cars_demo.csv` (1000 records)

This dataset was synthetically generated using `scripts/gen_used_cars_dataset.py` and represents used-car marketplace listings for demo and testing workflows.

Example import:

```bash
csv2opensearch --csv=testdata/used_cars_demo.csv --index=used-cars-demo --host=http://localhost:9200
```

Local dev import (uses local source code and HTTP host):

```bash
make dev
```

Seed the larger demo index:

```bash
make seed
make seed-dev
```

- `make seed` uses the installed `csv2opensearch` binary from your PATH
- `make seed-dev` uses local source (`go run`) for the same ingest
- `make seed` defaults to index `used-cars-demo`, which is also the default for `make query`

Basic free-text query on the demo index:

```bash
make query
make query N=10
```

List ingested indexes:

```bash
make indexes
```

Override Make variables (for remote or secured clusters):

```bash
make indexes HOST=http://127.0.0.1:9200
make query HOST=https://my-secure-os:9200 INDEX=used-cars-demo Q="toyota" N=5
make seed HOST=https://my-secure-os:9200 SEED_CSV=testdata/used_cars_demo.csv SEED_INDEX=used-cars-demo
```

Equivalent raw request:

```bash
curl -s -X POST "http://localhost:9200/used-cars-demo/_search?pretty" \
  -H "Content-Type: application/json" \
  -d '{"query":{"query_string":{"query":"Bucharest"}},"size":3}'
```

Regenerate datasets (deterministic):

```bash
/usr/bin/python3 scripts/gen_used_cars_dataset.py --rows 1000 --seed 42 --out testdata/used_cars_demo.csv
```

Or use Make:

```bash
make data
make data DATA_ROWS=200 DATA_SEED=7 DATA_OUT=testdata/custom.csv
```

Use a different generator script:

```bash
make data DATA_GEN_SCRIPT=scripts/my_generator.py DATA_ARGS="--out testdata/custom.csv --rows 200"
```

### Options

Run `csv2opensearch --help` for the full list of options:

```
➜  ~ csv2opensearch --help
Usage of ./csv2opensearch:
  -batch int
        Number of records that will be indexed in a single Bulk API request (default 100)
  -csv string
        Optional. Path to the CSV file. If missing, the data is read from stdin.
  -host string
        URL of the OpenSearch cluster (default "https://localhost:9200")
  -index string
        Optional. Name of the OpenSearch index where the data will end up. If not provided, data will be written to an index based on the CSV file name and current timestamp.  OpenSearch will automatically create the field mappings.
  -rate int
        Throttle the number of records that are indexed per second. By default, the ingestion is unthrottled. (default -1)
```

## Library

The `csv2opensearch` package exposes types that can be embedded into other applications. See the [Go documentation](https://pkg.go.dev/github.com/ovidiugiorgi/csv2opensearch) for more details.

```
go get github.com/ovidiugiorgi/csv2opensearch@latest
```

## Docker

You can use the `compose.yaml` file to quickly spin up a local (single node) OpenSearch cluster and Dashboards application (FKA Kibana).

URLs:
- OpenSearch: localhost:9200
- Dashboards: localhost:5601

### Start

Preferred:
```bash
make up
```

Alternatives:

```bash
./scripts/start-stack.sh
```

```bash
docker compose up -d
```

> Note: The default `compose.yaml` setup disables security and runs OpenSearch over HTTP on localhost.

### Stop

```bash
docker compose down
```

or:

```bash
make down
```

> Avoid `docker compose down -v` for normal workflows. The `-v` flag deletes the OpenSearch data volume and all indexed data.
> If you explicitly want a full reset, use `make delete-data`.

## Developer Hooks (pre-commit)

Install `pre-commit` and register hooks:

```bash
pre-commit install --hook-type pre-commit --hook-type pre-push
```

Install `goimports` if missing:

```bash
go install golang.org/x/tools/cmd/goimports@latest
```

Run all hooks once:

```bash
pre-commit run --all-files
```

Configured hooks:
- pre-commit: `check-yaml`, `gofmt`, `goimports`, `shellcheck`
- pre-push: `go test ./...`, `go vet ./...`
