#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_STARTED=0

cleanup() {
  local exit_code=$?

  if [[ "$COMPOSE_STARTED" -eq 1 ]]; then
    (cd "$ROOT_DIR" && docker compose down >/dev/null 2>&1) || true
  fi

  exit "$exit_code"
}

trap cleanup EXIT

export GOCACHE="${GOCACHE:-/tmp/go-build}"
export INTEGRATION_DATABASE_URL="${INTEGRATION_DATABASE_URL:-postgres://project:projectpass@localhost:5432/projectdb}"

echo "Starting Postgres with Docker Compose..."
(cd "$ROOT_DIR" && docker compose up -d >/dev/null)
COMPOSE_STARTED=1

echo "Applying migrations and seed data..."
"$ROOT_DIR/bootstrap_local_db.sh" >/dev/null

echo "Running integration tests..."
(cd "$ROOT_DIR" && go test -tags=integration ./internal/project -v)
