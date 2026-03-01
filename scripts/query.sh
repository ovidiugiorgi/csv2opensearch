#!/usr/bin/env bash
set -euo pipefail

host=""
index=""
query=""
size=""
dev_index=""

while (($# > 0)); do
  case "$1" in
    --host=*)
      host="${1#*=}"
      ;;
    --host)
      host="${2:-}"
      shift
      ;;
    --index=*)
      index="${1#*=}"
      ;;
    --index)
      index="${2:-}"
      shift
      ;;
    --query=*)
      query="${1#*=}"
      ;;
    --query)
      query="${2:-}"
      shift
      ;;
    --size=*)
      size="${1#*=}"
      ;;
    --size)
      size="${2:-}"
      shift
      ;;
    --dev-index=*)
      dev_index="${1#*=}"
      ;;
    --dev-index)
      dev_index="${2:-}"
      shift
      ;;
    *)
      echo "error: unknown argument: $1" >&2
      exit 1
      ;;
  esac
  shift
done

if [[ -z "$host" || -z "$index" || -z "$size" || -z "$dev_index" ]]; then
  echo "error: missing required arguments." >&2
  exit 1
fi

if [[ ! "$size" =~ ^[0-9]+$ ]]; then
  echo "error: N/QUERY_SIZE must be a positive integer (got '$size')." >&2
  exit 1
fi

if ((size <= 0)); then
  echo "error: N/QUERY_SIZE must be greater than 0 (got '$size')." >&2
  exit 1
fi

json_query=${query//\\/\\\\}
json_query=${json_query//\"/\\\"}

echo "OpenSearch Dev Tools query:"
echo "POST /$index/_search"
echo "{"
echo "  \"size\": $size,"
echo "  \"query\": {"
echo "    \"query_string\": {"
echo "      \"query\": \"$json_query\""
echo "    }"
echo "  }"
echo "}"
echo

if ! curl -sS -o /dev/null "$host" >/dev/null 2>&1; then
  echo "error: unable to reach OpenSearch at $host" >&2
  echo "hint: cluster may be down; run 'make up' and then 'make status'." >&2
  exit 1
fi

if ! curl -sS --fail-with-body -X POST "$host/$index/_search?pretty" \
  -H "Content-Type: application/json" \
  -d "{\"query\":{\"query_string\":{\"query\":\"$json_query\"}},\"size\":$size}"; then
  code=$?
  echo
  echo "error: search request failed (curl exit $code)." >&2
  echo "hint: current query index is '$index'." >&2
  echo "hint: run 'make seed' to populate the demo index, or query another one with e.g. make query INDEX=$dev_index" >&2
  exit "$code"
fi
