SHELL := /bin/sh

COMPOSE_FILE ?= $(CURDIR)/compose.yaml
TIMEOUT_SECONDS ?= 180
POLL_INTERVAL_SECONDS ?= 2
HOST ?= http://localhost:9200
CSV ?= testdata/used_cars_demo.csv
INDEX ?= used-cars-demo
DEV_INDEX ?= used-cars-demo
SEED_CSV ?= testdata/used_cars_demo.csv
SEED_INDEX ?= used-cars-demo
Q ?= Bucharest
QUERY_SIZE ?= 3
PYTHON ?= /usr/bin/python3
DATA_GEN_SCRIPT ?= scripts/gen_used_cars_dataset.py
DATA_ROWS ?= 1000
DATA_SEED ?= 42
DATA_OUT ?= testdata/used_cars_demo.csv
DATA_ARGS ?= --rows $(DATA_ROWS) --seed $(DATA_SEED) --out $(DATA_OUT)

COMPOSE := $(shell if command -v nerdctl >/dev/null 2>&1; then echo "nerdctl compose -f $(COMPOSE_FILE)"; elif command -v docker >/dev/null 2>&1; then echo "docker compose -f $(COMPOSE_FILE)"; fi)

.PHONY: help up up-nowait down delete-data pull restart status logs dev seed seed-dev indexes query data test

help:
	@echo "Targets:"
	@echo "  make up         - Start stack and wait for readiness"
	@echo "  make up-nowait  - Start stack in background without readiness checks"
	@echo "  make down       - Stop stack"
	@echo "  make delete-data - Stop stack and delete OpenSearch volume data"
	@echo "  make pull       - Pull images"
	@echo "  make restart    - Restart stack and wait for readiness"
	@echo "  make status     - Show compose status"
	@echo "  make logs       - Show recent logs from both services"
	@echo "  make dev        - Run local csv2opensearch against local OpenSearch"
	@echo "  make seed       - Ingest demo dataset with installed csv2opensearch CLI"
	@echo "  make seed-dev   - Ingest demo dataset with local go run"
	@echo "  make indexes    - List non-system indexes with docs and size"
	@echo "  make query      - Run a basic free-text search (Q=...) on INDEX"
	@echo "  make data       - Generate CSV data via script (override DATA_GEN_SCRIPT/DATA_ARGS)"
	@echo "  make test       - Run all Go tests"

up:
	@TIMEOUT_SECONDS=$(TIMEOUT_SECONDS) POLL_INTERVAL_SECONDS=$(POLL_INTERVAL_SECONDS) COMPOSE_FILE=$(COMPOSE_FILE) ./scripts/start-stack.sh

up-nowait:
	@test -n "$(COMPOSE)" || (echo "neither nerdctl nor docker is available in PATH" >&2; exit 1)
	@$(COMPOSE) up -d

down:
	@test -n "$(COMPOSE)" || (echo "neither nerdctl nor docker is available in PATH" >&2; exit 1)
	@$(COMPOSE) down

delete-data:
	@test -n "$(COMPOSE)" || (echo "neither nerdctl nor docker is available in PATH" >&2; exit 1)
	@$(COMPOSE) down -v

pull:
	@test -n "$(COMPOSE)" || (echo "neither nerdctl nor docker is available in PATH" >&2; exit 1)
	@$(COMPOSE) pull

restart:
	@$(MAKE) down
	@$(MAKE) up

status:
	@test -n "$(COMPOSE)" || (echo "neither nerdctl nor docker is available in PATH" >&2; exit 1)
	@$(COMPOSE) ps

logs:
	@test -n "$(COMPOSE)" || (echo "neither nerdctl nor docker is available in PATH" >&2; exit 1)
	@$(COMPOSE) logs --tail=100 opensearch-node opensearch-dashboards

dev:
	@go run ./cmd/csv2opensearch --host=$(HOST) --csv=$(CSV) --index=$(DEV_INDEX)

seed:
	@command -v csv2opensearch >/dev/null 2>&1 || (echo "csv2opensearch binary not found in PATH. Install it with: go install github.com/ovidiugiorgi/csv2opensearch/cmd/csv2opensearch@latest" >&2; exit 1)
	@csv2opensearch --host=$(HOST) --csv=$(SEED_CSV) --index=$(SEED_INDEX)

seed-dev:
	@go run ./cmd/csv2opensearch --host=$(HOST) --csv=$(SEED_CSV) --index=$(SEED_INDEX)

indexes:
	@curl -sS -o /dev/null "$(HOST)" >/dev/null 2>&1 || { \
		echo "error: unable to reach OpenSearch at $(HOST)" >&2; \
		echo "hint: cluster may be down; run 'make up' and then 'make status'." >&2; \
		exit 1; \
	}
	@echo "index docs.count store.size status"
	@curl -sS --fail-with-body "$(HOST)/_cat/indices?h=index,docs.count,store.size,status&s=index" | /usr/bin/grep -E -v '^\.' || true

query:
	@echo "OpenSearch Dev Tools query:"
	@echo "POST /$(INDEX)/_search"
	@echo "{"
	@echo "  \"size\": $(QUERY_SIZE),"
	@echo "  \"query\": {"
	@echo "    \"query_string\": {"
	@echo "      \"query\": \"$(Q)\""
	@echo "    }"
	@echo "  }"
	@echo "}"
	@echo
	@curl -sS -o /dev/null "$(HOST)" >/dev/null 2>&1 || { \
		echo "error: unable to reach OpenSearch at $(HOST)" >&2; \
		echo "hint: cluster may be down; run 'make up' and then 'make status'." >&2; \
		exit 1; \
	}
	@curl -sS --fail-with-body -X POST "$(HOST)/$(INDEX)/_search?pretty" \
		-H "Content-Type: application/json" \
		-d '{"query":{"query_string":{"query":"$(Q)"}},"size":$(QUERY_SIZE)}' || { \
		code=$$?; \
		echo; \
		echo "error: search request failed (curl exit $$code)." >&2; \
		echo "hint: current query index is '$(INDEX)'." >&2; \
		echo "hint: run 'make seed' to populate the demo index, or query another one with e.g. make query INDEX=$(DEV_INDEX)" >&2; \
		exit $$code; \
	}

data:
	@$(PYTHON) $(DATA_GEN_SCRIPT) $(DATA_ARGS)
	@echo "generated: $(DATA_OUT)"

test:
	@go test ./...
