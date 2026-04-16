#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
OUTPUT_FILE="${1:-$ROOT_DIR/internal/database/schema.sql}"

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-streaming}"
DB_PASSWORD="${DB_PASSWORD:-streaming}"
DB_NAME="${DB_NAME:-streaming}"
DB_DUMP_FROM_DOCKER="${DB_DUMP_FROM_DOCKER:-1}"

mkdir -p "$(dirname "$OUTPUT_FILE")"

TMP_FILE="$(mktemp)"
trap 'rm -f "$TMP_FILE"' EXIT

if [ "$DB_DUMP_FROM_DOCKER" = "1" ] && \
  docker compose -f "$ROOT_DIR/docker-compose.yml" ps postgres >/dev/null 2>&1; then
  docker compose -f "$ROOT_DIR/docker-compose.yml" exec -T postgres sh -c \
    "PGPASSWORD='$DB_PASSWORD' pg_dump --username='$DB_USER' --dbname='$DB_NAME' --schema-only --no-owner --no-privileges" \
    > "$TMP_FILE"
else
  PGPASSWORD="$DB_PASSWORD" pg_dump \
    --host="$DB_HOST" \
    --port="$DB_PORT" \
    --username="$DB_USER" \
    --dbname="$DB_NAME" \
    --schema-only \
    --no-owner \
    --no-privileges > "$TMP_FILE"
fi

mv "$TMP_FILE" "$OUTPUT_FILE"
echo "schema snapshot written to $OUTPUT_FILE"
