#!/usr/bin/env bash
set -euo pipefail

host="${1:-dealvm}"
remote_dir="/home/ubuntu/dealscanner"
repository_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
image_name="dealscanner:dealvm-native-build"
artifact_dir="$(mktemp -d "${TMPDIR:-/tmp}/dealscanner-dealvm.XXXXXX")"
container_id=""

log() {
  printf '\n[%s] %s\n' "$(date '+%H:%M:%S')" "$*"
}

require_docker_engine() {
  local check_pid
  local elapsed=0

  docker info >/dev/null 2>&1 &
  check_pid=$!

  while kill -0 "${check_pid}" >/dev/null 2>&1; do
    if (( elapsed >= 20 )); then
      kill "${check_pid}" >/dev/null 2>&1 || true
      wait "${check_pid}" >/dev/null 2>&1 || true
      printf '\nDocker did not respond within 20 seconds. Start or restart Docker Desktop, then rerun this script.\n' >&2
      return 1
    fi
    sleep 1
    elapsed=$((elapsed + 1))
  done

  if ! wait "${check_pid}"; then
    printf '\nDocker Desktop is not ready. Start it and wait for the engine to finish loading, then rerun this script.\n' >&2
    return 1
  fi
}

run_with_heartbeat() {
  local description=$1
  shift
  local command_pid
  local elapsed=0

  "$@" &
  command_pid=$!

  while kill -0 "${command_pid}" >/dev/null 2>&1; do
    sleep 15
    elapsed=$((elapsed + 15))
    if kill -0 "${command_pid}" >/dev/null 2>&1; then
      log "${description} is still running (${elapsed}s elapsed)"
    fi
  done

  wait "${command_pid}"
}

fail() {
  local exit_code=$?
  printf '\n[%s] Deployment failed at line %s while running: %s (exit %s)\n' \
    "$(date '+%H:%M:%S')" "${BASH_LINENO[0]}" "${BASH_COMMAND}" "${exit_code}" >&2
  exit "${exit_code}"
}

trap fail ERR

cleanup() {
  log "Cleaning up temporary local deployment artifacts"
  if [[ -n "${container_id}" ]]; then
    docker rm -f "${container_id}" >/dev/null 2>&1 || true
  fi
  rm -rf "${artifact_dir}"
}
trap cleanup EXIT

cd "${repository_dir}"

if [[ ! -f "${repository_dir}/.env" ]]; then
  echo "Missing local environment file: ${repository_dir}/.env" >&2
  exit 1
fi

# Build a Linux/x86_64 release locally. The VM never builds application code or
# runs Docker; Docker Desktop is used only to make the Linux artifact portable.
log "Deploying Dealscanner to ${host}"
log "Local repository: ${repository_dir}"
log "Temporary artifact directory: ${artifact_dir}"
log "Checking the local Docker engine"
require_docker_engine
log "Docker engine is ready"
log "Checking SSH connectivity"
ssh -v "${host}" "printf 'Connected to %s as %s\n' \"\$(hostname)\" \"\$(whoami)\""

log "Building the Linux/amd64 release locally (nothing is built on the VM)"
run_with_heartbeat "Local Linux/amd64 build" \
  docker build --progress=plain --platform linux/amd64 --tag "${image_name}" .

log "Extracting the application, frontend, and Playwright driver from the local image"
container_id="$(docker create "${image_name}")"
docker cp "${container_id}:/app/dealscanner" "${artifact_dir}/dealscanner"
docker cp "${container_id}:/app/frontend/dist" "${artifact_dir}/frontend"
docker cp "${container_id}:/root/.cache/ms-playwright-go/1.57.0" "${artifact_dir}/playwright-driver"
chmod 755 "${artifact_dir}/dealscanner"
du -sh "${artifact_dir}"/*

log "Creating the remote application directory"
ssh "${host}" "mkdir -p '${remote_dir}'"

log "Uploading .env securely (its contents will not be printed)"
local_env_sha256="$(shasum -a 256 "${repository_dir}/.env" | awk '{print $1}')"
scp "${repository_dir}/.env" "${host}:${remote_dir}/.env.incoming"
remote_env_sha256="$(ssh "${host}" "sha256sum '${remote_dir}/.env.incoming' | awk '{print \$1}'")"
if [[ "${local_env_sha256}" != "${remote_env_sha256}" ]]; then
  printf 'Transferred .env checksum does not match the current local .env\n' >&2
  exit 1
fi
ssh "${host}" "mv '${remote_dir}/.env.incoming' '${remote_dir}/.env' && chmod 600 '${remote_dir}/.env'"
log "Verified and activated the current local .env"

log "Checking and installing required VM runtime packages"
ssh "${host}" "if ! command -v node >/dev/null; then sudo apt-get update && sudo apt-get install -y nodejs; fi; if ! command -v chromium >/dev/null; then sudo apt-get update && sudo apt-get install -y chromium-browser; if command -v chromium-browser >/dev/null; then sudo ln -sfn \"\$(command -v chromium-browser)\" /usr/bin/chromium; fi; fi; command -v chromium >/dev/null; mkdir -p '${remote_dir}/releases' '${remote_dir}/data' && chmod 700 '${remote_dir}/data'"

release_id="$(date -u +%Y%m%dT%H%M%SZ)"
release_dir="${remote_dir}/releases/${release_id}"

log "Uploading release ${release_id} to ${host}:${release_dir}"
rsync -az --delete --progress "${artifact_dir}/" "${host}:${release_dir}/"

log "Uploading the systemd service definition"
scp deploy/dealvm/dealscanner.service "${host}:${remote_dir}/dealscanner.service"

log "Activating the release and starting Dealscanner"
ssh "${host}" "ln -sfn '${release_dir}' '${remote_dir}/current' && sudo install -m 644 '${remote_dir}/dealscanner.service' /etc/systemd/system/dealscanner.service && sudo systemctl daemon-reload && sudo systemctl enable dealscanner && sudo systemctl restart dealscanner && sleep 3 && curl --fail --silent --show-error http://127.0.0.1:3000/ >/dev/null && sudo systemctl --no-pager --full status dealscanner"

log "Deployment succeeded; Dealscanner is responding on the VM at http://127.0.0.1:3000/"
