#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
COMPOSE_FILE=${COMPOSE_FILE:-"$ROOT_DIR/compose.yaml"}
TIMEOUT_SECONDS=${TIMEOUT_SECONDS:-180}
POLL_INTERVAL_SECONDS=${POLL_INTERVAL_SECONDS:-2}

if [ ! -f "$COMPOSE_FILE" ]; then
  echo "compose file not found: $COMPOSE_FILE" >&2
  exit 1
fi

if command -v nerdctl >/dev/null 2>&1; then
  COMPOSE_CMD="nerdctl compose -f $COMPOSE_FILE"
elif command -v docker >/dev/null 2>&1; then
  COMPOSE_CMD="docker compose -f $COMPOSE_FILE"
else
  echo "neither nerdctl nor docker is available in PATH" >&2
  exit 1
fi

echo "using compose command: $COMPOSE_CMD"
echo "starting stack..."
sh -c "$COMPOSE_CMD up -d"

deadline=$(( $(date +%s) + TIMEOUT_SECONDS ))

echo "waiting for OpenSearch on http://localhost:9200 ..."
while :; do
  if curl -fsS http://localhost:9200 >/dev/null 2>&1; then
    echo "OpenSearch is ready"
    break
  fi
  if [ "$(date +%s)" -ge "$deadline" ]; then
    echo "timeout waiting for OpenSearch" >&2
    sh -c "$COMPOSE_CMD ps" || true
    exit 1
  fi
  sleep "$POLL_INTERVAL_SECONDS"
done

echo "waiting for Dashboards on http://localhost:5601/api/status ..."
while :; do
  status=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:5601/api/status || true)
  if [ "$status" = "200" ]; then
    echo "Dashboards is ready"
    break
  fi
  if [ "$(date +%s)" -ge "$deadline" ]; then
    echo "timeout waiting for Dashboards (last status: $status)" >&2
    sh -c "$COMPOSE_CMD ps" || true
    exit 1
  fi
  sleep "$POLL_INTERVAL_SECONDS"
done

echo "stack is ready"
echo "OpenSearch: http://localhost:9200"
echo "Dashboards: http://localhost:5601"
