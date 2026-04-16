#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-streaming}"
DB_PASSWORD="${DB_PASSWORD:-streaming}"
DB_NAME="${DB_NAME:-streaming}"

# If postgres container is running, prefer resetting from inside Docker.
use_docker_psql() {
  docker compose -f "$ROOT_DIR/docker-compose.yml" ps postgres >/dev/null 2>&1
}

run_sql() {
  sql="$1"
  if use_docker_psql; then
    docker compose -f "$ROOT_DIR/docker-compose.yml" exec -T postgres sh -c \
      "PGPASSWORD='$DB_PASSWORD' psql --username='$DB_USER' --dbname='postgres' --no-psqlrc -v ON_ERROR_STOP=1 -c \"$sql\""
  else
    PGPASSWORD="$DB_PASSWORD" psql \
      --host="$DB_HOST" \
      --port="$DB_PORT" \
      --username="$DB_USER" \
      --dbname="postgres" \
      --no-psqlrc \
      -v ON_ERROR_STOP=1 \
      -c "$sql"
  fi
}

echo "[hard-reset] terminating active connections to $DB_NAME"
run_sql "REVOKE CONNECT ON DATABASE \"$DB_NAME\" FROM PUBLIC;"
run_sql "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND pid <> pg_backend_pid();"

echo "[hard-reset] dropping database $DB_NAME (if exists)"
run_sql "DROP DATABASE IF EXISTS \"$DB_NAME\";"

echo "[hard-reset] creating database $DB_NAME"
run_sql "CREATE DATABASE \"$DB_NAME\";"
run_sql "GRANT ALL PRIVILEGES ON DATABASE \"$DB_NAME\" TO \"$DB_USER\";"

echo "[hard-reset] running migrations"
(cd "$ROOT_DIR" && go run ./cmd/migrate -action up)

echo "[hard-reset] seeding fixtures (-reset)"
(cd "$ROOT_DIR" && go run ./cmd/seed -reset)

echo "[hard-reset] done"

