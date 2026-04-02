#!/usr/bin/env bash

set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${PROJECT_DIR}"

ENV_FILE="${ENV_FILE:-.env}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose-postgres.yml}"
BACKUP_DIR_DEFAULT="${PROJECT_DIR}/backups"
RETENTION_DAYS_DEFAULT="7"

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "Missing ${ENV_FILE}. Copy .env.example and fill real values."
  exit 1
fi

set -a
source "${ENV_FILE}"
set +a

BACKUP_DIR="${POSTGRES_BACKUP_DIR:-${BACKUP_DIR_DEFAULT}}"
RETENTION_DAYS="${POSTGRES_BACKUP_RETENTION_DAYS:-${RETENTION_DAYS_DEFAULT}}"

if [[ -z "${POSTGRES_USER:-}" || -z "${POSTGRES_DB:-}" ]]; then
  echo "POSTGRES_USER and POSTGRES_DB are required in ${ENV_FILE}"
  exit 1
fi

mkdir -p "${BACKUP_DIR}"

docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" up -d postgres

until docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" exec -T postgres \
  pg_isready -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" >/dev/null 2>&1; do
  echo "Waiting for postgres to become ready"
  sleep 2
done

timestamp="$(date +%Y%m%d%H%M%S)"
backup_path="${BACKUP_DIR}/${POSTGRES_DB}_${timestamp}.dump"

docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" exec -T postgres \
  pg_dump -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -Fc --no-owner --no-privileges > "${backup_path}"

if command -v sha256sum >/dev/null 2>&1; then
  sha256sum "${backup_path}" > "${backup_path}.sha256"
fi

find "${BACKUP_DIR}" -maxdepth 1 -type f -name "${POSTGRES_DB}_*.dump" -mtime +"${RETENTION_DAYS}" -delete
find "${BACKUP_DIR}" -maxdepth 1 -type f -name "${POSTGRES_DB}_*.dump.sha256" -mtime +"${RETENTION_DAYS}" -delete

echo "Backup created: ${backup_path}"
