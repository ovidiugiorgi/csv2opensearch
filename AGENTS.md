# AGENTS.md

Guidance for coding agents working in this repository.

## Project Intent

- `csv2opensearch` imports CSV records into OpenSearch using inferred fields.
- Primary local workflow is a sandbox/dev stack (OpenSearch + Dashboards) via Compose and Make.

## Current Local Stack Defaults

- Compose file: `compose.yaml`
- Backward compatibility alias: `docker-compose.yaml` (symlink to `compose.yaml`)
- OpenSearch image: `opensearchproject/opensearch:3.5.0`
- Dashboards image: `opensearchproject/opensearch-dashboards:3.5.0`
- Security is intentionally disabled for local sandbox usage.
- Ports are bound to localhost:
  - OpenSearch: `http://localhost:9200`
  - Dashboards: `http://localhost:5601`

## CLI/Auth Expectations

- CLI default host is intentionally `https://localhost:9200` for backward compatibility.
- Local sandbox runs over HTTP, so local commands must pass `--host=http://localhost:9200` (or use Make targets that already do this).
- `OS_USER` and `OS_PASSWORD` support remains in code for secured clusters.

## Makefile Conventions

Prefer these commands when guiding users:

1. `make up` (starts stack and waits until both services are ready)
2. `make seed` (ingests demo dataset with installed CLI)
3. `make query` (queries demo index; default returns 3 results)

Useful targets:

- `make up-nowait`
- `make status`
- `make logs`
- `make down` (safe stop)
- `make delete-data` (destructive reset; deletes volume data)
- `make seed-dev` (same as seed but uses local `go run`)
- `make indexes` (lists non-system indexes)
- `make data` (dataset generation wrapper)

Important defaults:

- `HOST=http://localhost:9200`
- `INDEX=used-cars-demo` (used by `make query`)
- `DEV_INDEX=used-cars-demo` (used by `make dev`)
- `SEED_INDEX=used-cars-demo`
- `QUERY_SIZE=3`

## Demo Data

Committed datasets:

- `testdata/used_cars_demo.csv` (1000 rows)
- `testdata/used_cars_integration_exact_20.csv` (20 rows)
- `testdata/used_cars_integration_partial_23.csv` (23 rows)

Generator:

- `scripts/gen_used_cars_dataset.py` (deterministic via seed)
- Exposed through `make data`
- If new datasets are needed, generate deterministic CSVs and commit them under `testdata/`.
- Prefer dedicated integration fixtures over ad-hoc inline data when behavior depends on row counts or batch boundaries.

## Documentation/UX Rules

- Keep README quickstart tight and in this order:
  - `make up`
  - `make seed`
  - `make query`
- If Make defaults change, update README examples in the same change.
- If Make defaults or datasets change, update README and this file in the same change.
- Keep error hints actionable (for example: suggest `make up`, `make status`, `make seed`).
- Do not recommend `down -v` as default cleanup; recommend `make down`.
- Remove obsolete references/files when replacing defaults.

## Testing

- Run tests with: `go test ./...`
- In restricted/sandbox environments, use: `GOCACHE=/tmp/go-build go test ./...`
- For bug fixes/regressions, start with a failing test before changing runtime logic.
- Prefer integration-style tests (mock OpenSearch) for shipping-path behavior.
- Prefer package-external tests (`csv2opensearch_test`) to validate public behavior.
- Follow a red/green loop:
  1. Add/adjust test and confirm it fails for the expected reason.
  2. Implement the smallest fix.
  3. Run full suite (`go test ./...`) before concluding.

## When Changing Runtime Behavior

If modifying any of these, call out compatibility impact in commit/PR notes:

- CLI default `--host`
- Compose security mode (HTTP/no-auth vs HTTPS/auth)
- Default Make index variables (`INDEX`, `SEED_INDEX`, `DEV_INDEX`)
- Keep changes small and auditable: guard at boundaries (for example, batch size validation and empty-batch no-ops).
- Add regression tests in the same change when runtime behavior is modified.
