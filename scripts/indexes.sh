#!/usr/bin/env bash
set -euo pipefail

host=""

while (($# > 0)); do
  case "$1" in
    --host=*)
      host="${1#*=}"
      ;;
    --host)
      host="${2:-}"
      shift
      ;;
    *)
      echo "error: unknown argument: $1" >&2
      exit 1
      ;;
  esac
  shift
done

if [[ -z "$host" ]]; then
  echo "error: missing required argument --host." >&2
  exit 1
fi

if ! curl -sS -o /dev/null "$host" >/dev/null 2>&1; then
  echo "error: unable to reach OpenSearch at $host" >&2
  echo "hint: cluster may be down; run 'make up' and then 'make status'." >&2
  exit 1
fi

echo "index docs.count store.size status"
curl -sS --fail-with-body "$host/_cat/indices?h=index,docs.count,store.size,status&s=index" | /usr/bin/grep -E -v '^\.' || true
