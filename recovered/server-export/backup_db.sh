#!/bin/bash
set -euo pipefail

DB_NAME="${DB_NAME:-isheecrm}"
CONTAINER_NAME="${CONTAINER_NAME:-postgres-container}"
BACKUP_DIR="${BACKUP_DIR:-/root/db-backups}"
REMOTE_DEST="${REMOTE_DEST:-gdrive:/backups}"
TIMESTAMP="$(date +%Y%m%d%H%M%S)"
BACKUP_NAME="${DB_NAME}_${TIMESTAMP}.dump"
BACKUP_PATH="${BACKUP_DIR}/${BACKUP_NAME}"
CONTAINER_TMP="/tmp/${BACKUP_NAME}"

mkdir -p "${BACKUP_DIR}"

if ! docker inspect -f '{{.State.Running}}' "${CONTAINER_NAME}" 2>/dev/null | grep -qx true; then
  echo "container ${CONTAINER_NAME} is not running" >&2
  exit 1
fi

docker exec "${CONTAINER_NAME}" sh -lc "rm -f '${CONTAINER_TMP}' && pg_dump -U postgres -F c '${DB_NAME}' > '${CONTAINER_TMP}'"
docker cp "${CONTAINER_NAME}:${CONTAINER_TMP}" "${BACKUP_PATH}"
docker exec "${CONTAINER_NAME}" rm -f "${CONTAINER_TMP}"

docker run --rm -v "${BACKUP_DIR}:/work" postgres:17.6-bookworm sh -lc "pg_restore -l '/work/${BACKUP_NAME}' >/dev/null"

if command -v rclone >/dev/null 2>&1; then
  rclone copy "${BACKUP_PATH}" "${REMOTE_DEST}"
fi

echo "backup saved: ${BACKUP_PATH}"
