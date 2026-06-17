#!/usr/bin/env bash
set -euo pipefail

DEPLOY_ROOT="${DEPLOY_ROOT:-/opt/dealscanner}"
DEPLOY_ENV_FILE="${DEPLOY_ROOT}/deploy.env"
APP_ENV_FILE="${DEPLOY_ROOT}/app.env"
COMPOSE_FILE="${DEPLOY_ROOT}/docker-compose.prod.yml"

if [[ -z "${DEPLOY_IMAGE:-}" ]]; then
  echo "DEPLOY_IMAGE is required" >&2
  exit 1
fi

if [[ -z "${DEPLOY_IMAGE_TAG:-}" ]]; then
  echo "DEPLOY_IMAGE_TAG is required" >&2
  exit 1
fi

if [[ ! -f "${DEPLOY_ENV_FILE}" ]]; then
  echo "Missing deploy env file: ${DEPLOY_ENV_FILE}" >&2
  exit 1
fi

if [[ ! -f "${APP_ENV_FILE}" ]]; then
  echo "Missing app env file: ${APP_ENV_FILE}" >&2
  exit 1
fi

# shellcheck disable=SC1090
source "${DEPLOY_ENV_FILE}"

if [[ -z "${GHCR_USERNAME:-}" || -z "${GHCR_TOKEN:-}" ]]; then
  echo "GHCR_USERNAME and GHCR_TOKEN are required in ${DEPLOY_ENV_FILE}" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
source "${APP_ENV_FILE}"
set +a

if [[ -z "${TURSO_DATABASE_URL:-}" || -z "${TURSO_AUTH_TOKEN:-}" ]]; then
  echo "TURSO_DATABASE_URL and TURSO_AUTH_TOKEN are required in ${APP_ENV_FILE}" >&2
  exit 1
fi

HEALTHCHECK_URL="${HEALTHCHECK_URL:-http://127.0.0.1:3000/}"

cd "${DEPLOY_ROOT}"

DOCKER=(docker)
if ! docker info >/dev/null 2>&1; then
  DOCKER=(sudo docker)
fi

if ! "${DOCKER[@]}" compose version >/dev/null 2>&1; then
  echo "Docker Compose plugin is required. Install docker-compose-plugin on the VM." >&2
  exit 1
fi

printf '%s\n' "${GHCR_TOKEN}" | "${DOCKER[@]}" login ghcr.io --username "${GHCR_USERNAME}" --password-stdin

export DEPLOY_IMAGE
export DEPLOY_IMAGE_TAG

"${DOCKER[@]}" compose -f "${COMPOSE_FILE}" pull app
"${DOCKER[@]}" compose -f "${COMPOSE_FILE}" up -d --remove-orphans app

curl --fail --silent --show-error "${HEALTHCHECK_URL}" >/dev/null

"${DOCKER[@]}" image prune -f >/dev/null 2>&1 || true
