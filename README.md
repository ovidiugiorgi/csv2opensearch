# csv2opensearch

[![Go Reference](https://pkg.go.dev/badge/github.com/ovidiugiorgi/csv2opensearch.svg)](https://pkg.go.dev/github.com/ovidiugiorgi/csv2opensearch)

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

## Usage

### Auth (optional)

Add `OS_USER` and `OS_PASSWORD` to the environment when the cluster is using basic authentication.

```
export OS_USER=admin
export OS_PASSWORD=admin
```

### Run the import

Read from file:
```
csv2opensearch --csv=test.csv 
```

Read from stdin:
```
cat test.csv | csv2opensearch --index=test
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

You can use the `docker-compose.yaml` file to quickly spin up a local (single node) OpenSearch cluster and Dashboards application (FKA Kibana).

URLs:
- OpenSearch: localhost:9200
- Dashboards: localhost:5601

### Start

```
docker-compose up -d
```

> Note: By default, the cluster will require basic authentication using the `admin/admin` credentials.

### Stop

```
docker-compose down
```