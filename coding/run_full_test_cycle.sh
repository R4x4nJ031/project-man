#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPORT_PATH="${1:-}"
SERVER_PID=""
COMPOSE_STARTED=0

export GOCACHE="${GOCACHE:-/tmp/go-build}"
export PORT="${PORT:-18080}"
export JWT_SECRET="${JWT_SECRET:-supersecretkey}"
export JWT_ISSUER="${JWT_ISSUER:-project-man}"
export JWT_AUDIENCE="${JWT_AUDIENCE:-project-man-users}"
export DATABASE_URL="${DATABASE_URL:-postgres://project:projectpass@localhost:5432/projectdb}"
export BASE_URL="${BASE_URL:-http://127.0.0.1:$PORT}"

cleanup() {
  local exit_code=$?

  if [[ -n "$SERVER_PID" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi

  if [[ "$COMPOSE_STARTED" -eq 1 ]]; then
    (cd "$ROOT_DIR" && docker compose down >/dev/null 2>&1) || true
  fi

  exit "$exit_code"
}

trap cleanup EXIT

wait_for_postgres() {
  local attempts=30
  local i

  for ((i = 1; i <= attempts; i++)); do
    if docker exec project-man-postgres pg_isready -U project -d projectdb >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done

  echo "Postgres did not become ready in time." >&2
  return 1
}

wait_for_server() {
  local attempts=30
  local i

  for ((i = 1; i <= attempts; i++)); do
    if curl -fsS "$BASE_URL/healthz" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done

  echo "Server did not become ready at $BASE_URL." >&2
  return 1
}

echo "Starting Postgres with Docker Compose..."
(cd "$ROOT_DIR" && docker compose up -d)
COMPOSE_STARTED=1

echo "Waiting for Postgres..."
wait_for_postgres

echo "Applying migrations and seed data..."
"$ROOT_DIR/bootstrap_local_db.sh"

echo "Starting API server..."
: >"$ROOT_DIR/server.test.log"
(
  cd "$ROOT_DIR"
  go run ./cmd/server/main.go
) >"$ROOT_DIR/server.test.log" 2>&1 &
SERVER_PID=$!

echo "Waiting for API server..."
wait_for_server

echo "Running endpoint tests..."
if [[ -n "$REPORT_PATH" ]]; then
  "$ROOT_DIR/run_endpoint_tests.sh" "$REPORT_PATH"
else
  "$ROOT_DIR/run_endpoint_tests.sh"
fi

if [[ -n "$REPORT_PATH" ]]; then
  echo "Report written to: $REPORT_PATH"
fi
echo "Server log written to: $ROOT_DIR/server.test.log"
