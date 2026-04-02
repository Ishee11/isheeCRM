#!/usr/bin/env bash

set -euo pipefail

ENV_FILE="${ENV_FILE:-.env}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose-postgres.yml}"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-migrations}"

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "Env file not found: ${ENV_FILE}" >&2
  exit 1
fi

if [[ ! -d "${MIGRATIONS_DIR}" ]]; then
  echo "Migrations directory not found: ${MIGRATIONS_DIR}" >&2
  exit 1
fi

POSTGRES_USER="$(grep -E '^POSTGRES_USER=' "${ENV_FILE}" | cut -d= -f2-)"
POSTGRES_DB="$(grep -E '^POSTGRES_DB=' "${ENV_FILE}" | cut -d= -f2-)"

if [[ -z "${POSTGRES_USER}" || -z "${POSTGRES_DB}" ]]; then
  echo "POSTGRES_USER and POSTGRES_DB must be set in ${ENV_FILE}" >&2
  exit 1
fi

docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" up -d postgres

until docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" exec -T postgres \
  pg_isready -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" >/dev/null 2>&1; do
  echo "Waiting for postgres to become ready"
  sleep 2
done

for migration in "${MIGRATIONS_DIR}"/*.sql; do
  echo "Applying ${migration}"
  docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" exec -T postgres \
    psql -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -v ON_ERROR_STOP=1 < "${migration}"
done

echo "Migrations applied successfully"
