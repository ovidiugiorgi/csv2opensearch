# csv2opensearch

Import CSV files into your OpenSearch cluster without needing to define a schema for the data. 

The tool automatically creates all document fields based on the file headers. All values are treated as strings.

## Install

```
go install github.com/ovidiugiorgi/csv2opensearch
```

## Usage

Add `OS_USER` and `OS_PASSWORD` to the environment if the cluster is using basic authentication. 

```
csv2opensearch --file=test.csv --host=https://localhost:9200 --index=test --batch=100 --rate=500
```

## Docker

You can use the `docker-compose.yaml` file to quickly spin up a local (single node) OpenSearch cluster and Dashboards application (FKA Kibana).

Start:
```
docker-compose up -d
```

Note: By default, the cluster will require basic authentication using the `admin/admin` credentials.

Stop:
```
docker-compose down
```