#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONTAINER_NAME="${CONTAINER_NAME:-project-man-postgres}"
DB_USER="${DB_USER:-project}"
DB_NAME="${DB_NAME:-projectdb}"
GOOSE_DRIVER="${GOOSE_DRIVER:-postgres}"
GOOSE_VERSION="${GOOSE_VERSION:-v3.24.1}"
GOOSE_DBSTRING="${GOOSE_DBSTRING:-postgres://project:projectpass@localhost:5432/projectdb?sslmode=disable}"

run_sql_file() {
  local file_path="$1"
  echo "Applying $(basename "$file_path")"
  docker exec -i "$CONTAINER_NAME" psql -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 <"$file_path"
}

echo "Running goose migrations..."
(
  cd "$ROOT_DIR"
  go run "github.com/pressly/goose/v3/cmd/goose@$GOOSE_VERSION" \
    -dir "$ROOT_DIR/db/migrations" \
    "$GOOSE_DRIVER" \
    "$GOOSE_DBSTRING" \
    up
)

for seed in "$ROOT_DIR"/db/seeds/*.sql; do
  run_sql_file "$seed"
done

echo "Database bootstrap complete."
