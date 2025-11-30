#!/usr/bin/env bash
# Quick helper for hitting the /timeentries REST API.

set -euo pipefail

BASE_URL=${BASE_URL:-http://localhost:8080}

usage() {
  cat <<'TXT'
Usage: rest-tests.sh [create|list|delete <ID>]
  create  – posts a demo time entry and prints the JSON response
  list    – fetches every /timeentries record
  delete  – removes an entry by ID (shows status headers)

Environment:
  BASE_URL  Base URL for the API (default http://localhost:8080)
TXT
}

create_entry() {
  curl -sS -X POST "$BASE_URL/timeentries" \
    -H "Content-Type: application/json" \
    -d '{
      "Description": "REST tester entry",
      "Project": "CLI Demo",
      "StartTime": "2024-05-17T08:00:00Z",
      "EndTime":   "2024-05-17T10:30:00Z"
    }'
}

list_entries() {
  if command -v jq >/dev/null 2>&1; then
    curl -sS "$BASE_URL/timeentries" | jq .
  else
    curl -sS "$BASE_URL/timeentries"
  fi
}

delete_entry() {
  local id=$1
  curl -sS -i -X DELETE "$BASE_URL/timeentries/$id"
}

case "${1:-}" in
  create)
    create_entry
    ;;
  list)
    list_entries
    ;;
  delete)
    shift
    if [[ $# -ne 1 ]]; then
      echo "delete requires an ID argument" >&2
      exit 1
    fi
    delete_entry "$1"
    ;;
  *)
    usage
    exit 1
    ;;
esac
