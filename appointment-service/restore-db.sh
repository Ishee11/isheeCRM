#!/usr/bin/env bash

set -euo pipefail

CALLER_DIR="$(pwd)"
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${PROJECT_DIR}"

if [[ $# -lt 1 ]]; then
  echo "Usage: $0 /absolute/or/relative/path/to/backup.dump"
  exit 1
fi

DUMP_PATH="$1"
if [[ "${DUMP_PATH}" != /* ]]; then
  DUMP_PATH="${CALLER_DIR}/${DUMP_PATH}"
fi
ENV_FILE="${ENV_FILE:-.env}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose-postgres.yml}"
APP_COMPOSE_FILE="${APP_COMPOSE_FILE:-docker-compose-app.yml}"
STOP_APP_BEFORE_RESTORE="${STOP_APP_BEFORE_RESTORE:-true}"
START_APP_AFTER_RESTORE="${START_APP_AFTER_RESTORE:-true}"

if [[ ! -f "${DUMP_PATH}" ]]; then
  echo "Dump file not found: ${DUMP_PATH}"
  exit 1
fi

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "Missing ${ENV_FILE}. Copy .env.example and fill real values."
  exit 1
fi

set -a
source "${ENV_FILE}"
set +a

if [[ -z "${POSTGRES_USER:-}" || -z "${POSTGRES_DB:-}" ]]; then
  echo "POSTGRES_USER and POSTGRES_DB are required in ${ENV_FILE}"
  exit 1
fi

APP_NETWORK_NAME="${APP_NETWORK_NAME:-appointment-service_app-network}"

if ! docker network inspect "${APP_NETWORK_NAME}" >/dev/null 2>&1; then
  docker network create "${APP_NETWORK_NAME}" >/dev/null
fi

if [[ "${STOP_APP_BEFORE_RESTORE}" == "true" && -f "${APP_COMPOSE_FILE}" ]]; then
  docker compose --env-file "${ENV_FILE}" -f "${APP_COMPOSE_FILE}" stop app || true
fi

docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" up -d postgres

until docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" exec -T postgres \
  pg_isready -U "${POSTGRES_USER}" -d postgres >/dev/null 2>&1; do
  echo "Waiting for postgres to become ready"
  sleep 2
done

docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" exec -T postgres \
  psql -U "${POSTGRES_USER}" -d postgres -v ON_ERROR_STOP=1 \
  -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '${POSTGRES_DB}' AND pid <> pg_backend_pid();"

docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" exec -T postgres \
  psql -U "${POSTGRES_USER}" -d postgres -v ON_ERROR_STOP=1 \
  -c "DROP DATABASE IF EXISTS \"${POSTGRES_DB}\";"

if [[ "${DUMP_PATH}" == *.sql ]]; then
  docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" exec -T postgres \
    psql -U "${POSTGRES_USER}" -d postgres -v ON_ERROR_STOP=1 < "${DUMP_PATH}"
else
  docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" exec -T postgres \
    pg_restore -U "${POSTGRES_USER}" -d postgres --clean --if-exists --create --no-owner --no-privileges < "${DUMP_PATH}"
fi

docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" exec -T postgres \
  psql -U "${POSTGRES_USER}" -d postgres -v ON_ERROR_STOP=1 \
  -c "ALTER DATABASE \"${POSTGRES_DB}\" REFRESH COLLATION VERSION;"

docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" exec -T postgres \
  psql -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -v ON_ERROR_STOP=1 \
  -c "REINDEX DATABASE \"${POSTGRES_DB}\";"

if [[ "${START_APP_AFTER_RESTORE}" == "true" && -f "${APP_COMPOSE_FILE}" ]]; then
  docker compose --env-file "${ENV_FILE}" -f "${APP_COMPOSE_FILE}" up -d
fi

echo "Restore completed successfully from ${DUMP_PATH}"
