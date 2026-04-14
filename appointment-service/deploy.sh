#!/usr/bin/env bash

set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${PROJECT_DIR}"

ENV_FILE="${ENV_FILE:-.env}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.yml}"
HEALTHCHECK_RETRIES="${HEALTHCHECK_RETRIES:-30}"
HEALTHCHECK_DELAY_SECONDS="${HEALTHCHECK_DELAY_SECONDS:-2}"

if [[ ! -f "${ENV_FILE}" ]]; then
echo "Missing ${ENV_FILE}. Copy .env.example and fill real values."
exit 1
fi

# Загружаем переменные окружения

set -a
source "${ENV_FILE}"
set +a

# Теперь APP_PORT уже доступен

HEALTHCHECK_URL="${HEALTHCHECK_URL:-[http://127.0.0.1:${APP_PORT:-8080}/healthz}](http://127.0.0.1:${APP_PORT:-8080}/healthz})"

if [[ -z "${APP_IMAGE:-}" ]]; then
echo "APP_IMAGE is required in ${ENV_FILE}"
exit 1
fi

if [[ -z "${IMAGE_TAG:-}" ]]; then
echo "IMAGE_TAG is required in ${ENV_FILE}"
exit 1
fi

echo "Deploying ${APP_IMAGE}:${IMAGE_TAG}"

docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" pull
docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" up -d

attempt=1

healthcheck_cmd() {
if command -v curl >/dev/null 2>&1; then
curl -fsS "${HEALTHCHECK_URL}" >/dev/null
return
fi

if command -v wget >/dev/null 2>&1; then
wget -q -O /dev/null "${HEALTHCHECK_URL}"
return
fi

echo "Neither curl nor wget is installed on the server"
return 1
}

until healthcheck_cmd; do
if (( attempt >= HEALTHCHECK_RETRIES )); then
echo "Healthcheck failed after ${HEALTHCHECK_RETRIES} attempts"
docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" ps
exit 1
fi

echo "Waiting for app healthcheck (${attempt}/${HEALTHCHECK_RETRIES})"
attempt=$((attempt + 1))
sleep "${HEALTHCHECK_DELAY_SECONDS}"
done

echo "Deployment completed successfully"
