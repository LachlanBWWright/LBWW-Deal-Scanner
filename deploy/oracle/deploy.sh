#!/usr/bin/env bash
set -euo pipefail

DEPLOY_ROOT="${DEPLOY_ROOT:-/opt/dealscanner}"
APP_ENV_FILE="${DEPLOY_ROOT}/app.env"
COMPOSE_ENV_FILE="${DEPLOY_ROOT}/compose.env"
COMPOSE_FILE="${DEPLOY_ROOT}/docker-compose.prod.yml"

if [[ -z "${DEPLOY_IMAGE:-}" ]]; then
  echo "DEPLOY_IMAGE is required" >&2
  exit 1
fi

if [[ -z "${DEPLOY_IMAGE_TAG:-}" ]]; then
  echo "DEPLOY_IMAGE_TAG is required" >&2
  exit 1
fi

if [[ ! -f "${APP_ENV_FILE}" ]]; then
  echo "Missing app env file: ${APP_ENV_FILE}" >&2
  exit 1
fi

if [[ -z "${GHCR_USERNAME:-}" ]]; then
  echo "GHCR_USERNAME is required" >&2
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

if ! command -v docker >/dev/null 2>&1; then
  if [[ -f /etc/os-release ]]; then
    # shellcheck disable=SC1091
    source /etc/os-release
  fi

  if [[ "${ID:-}" != "ubuntu" && "${ID_LIKE:-}" != *"debian"* ]]; then
    echo "Docker is not installed, and automatic install is only supported on Ubuntu/Debian." >&2
    exit 1
  fi

  sudo apt-get update
  sudo apt-get install -y ca-certificates curl gnupg
  sudo install -m 0755 -d /etc/apt/keyrings
  sudo rm -f /etc/apt/keyrings/docker.gpg
  curl -fsSL "https://download.docker.com/linux/${ID}/gpg" | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
  sudo chmod a+r /etc/apt/keyrings/docker.gpg

  codename="${VERSION_CODENAME:-}"
  if [[ -z "${codename}" ]]; then
    echo "VERSION_CODENAME is required to install Docker automatically." >&2
    exit 1
  fi

  echo \
    "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/${ID} ${codename} stable" \
    | sudo tee /etc/apt/sources.list.d/docker.list >/dev/null

  sudo apt-get update
  sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  sudo systemctl enable --now docker
fi

DOCKER=(docker)
if ! docker info >/dev/null 2>&1; then
  DOCKER=(sudo docker)
fi

if ! "${DOCKER[@]}" compose version >/dev/null 2>&1; then
  echo "Docker Compose plugin is required. Install docker-compose-plugin on the VM." >&2
  exit 1
fi

if ! IFS= read -r GHCR_TOKEN || [[ -z "${GHCR_TOKEN}" ]]; then
  echo "GHCR token is required on standard input" >&2
  exit 1
fi

printf '%s\n' "${GHCR_TOKEN}" | "${DOCKER[@]}" login ghcr.io --username "${GHCR_USERNAME}" --password-stdin
unset GHCR_TOKEN

export DEPLOY_IMAGE
export DEPLOY_IMAGE_TAG

printf 'DEPLOY_IMAGE=%s\nDEPLOY_IMAGE_TAG=%s\n' \
  "${DEPLOY_IMAGE}" \
  "${DEPLOY_IMAGE_TAG}" > "${COMPOSE_ENV_FILE}"
chmod 600 "${COMPOSE_ENV_FILE}"

"${DOCKER[@]}" compose --env-file "${APP_ENV_FILE}" --env-file "${COMPOSE_ENV_FILE}" -f "${COMPOSE_FILE}" pull app
"${DOCKER[@]}" compose --env-file "${APP_ENV_FILE}" --env-file "${COMPOSE_ENV_FILE}" -f "${COMPOSE_FILE}" up \
  -d \
  --force-recreate \
  --remove-orphans \
  app

healthcheck_attempts="${HEALTHCHECK_ATTEMPTS:-30}"
healthcheck_interval_seconds="${HEALTHCHECK_INTERVAL_SECONDS:-2}"

for ((attempt = 1; attempt <= healthcheck_attempts; attempt += 1)); do
  if curl --fail --silent --show-error --connect-timeout 2 --max-time 5 "${HEALTHCHECK_URL}" >/dev/null; then
    echo "Application is ready."
    "${DOCKER[@]}" image prune -f >/dev/null 2>&1 || true
    exit 0
  fi

  if ((attempt < healthcheck_attempts)); then
    echo "Waiting for application readiness (${attempt}/${healthcheck_attempts})..."
    sleep "${healthcheck_interval_seconds}"
  fi
done

echo "Application failed to become ready at ${HEALTHCHECK_URL}" >&2
"${DOCKER[@]}" compose --env-file "${APP_ENV_FILE}" --env-file "${COMPOSE_ENV_FILE}" -f "${COMPOSE_FILE}" ps >&2
"${DOCKER[@]}" compose --env-file "${APP_ENV_FILE}" --env-file "${COMPOSE_ENV_FILE}" -f "${COMPOSE_FILE}" logs --tail 100 app >&2
exit 1
