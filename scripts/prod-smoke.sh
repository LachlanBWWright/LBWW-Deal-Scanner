#!/usr/bin/env bash
set -euo pipefail

readonly SCRIPT_DIR="$(CDPATH='' cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=_common.sh
source "$SCRIPT_DIR/_common.sh"

require_command curl
require_command docker

readonly IMAGE_NAME="${IMAGE_NAME:-dealscanner:smoke}"
readonly CONTAINER_NAME="dealscanner-smoke-$$"
readonly HOST_PORT="${SMOKE_PORT:-3000}"
database_dir="$(mktemp -d)"

cleanup() {
  docker logs "$CONTAINER_NAME" 2>/dev/null || true
  docker rm --force "$CONTAINER_NAME" >/dev/null 2>&1 || true
  rm -rf "$database_dir"
}
trap cleanup EXIT

run_step "Building production Docker image" docker build --tag "$IMAGE_NAME" "$REPO_ROOT"

run_step "Starting production container" docker run --detach \
  --name "$CONTAINER_NAME" \
  --publish "127.0.0.1:${HOST_PORT}:3000" \
  --mount "type=bind,source=${database_dir},target=/var/lib/dealscanner" \
  --env DATABASE_URL=file:/var/lib/dealscanner/dealscanner.db \
  --env API_SECRET=local-smoke-test-secret \
  --env ENABLE_TESTING_API=false \
  --env SCHEDULED_SCANNER_MODE=false \
  --env CASH_CONVERTERS=false \
  --env CS_ITEMS=false \
  --env EBAY=false \
  --env GUMTREE=false \
  --env SALVOS=false \
  --env STEAM_QUERY=false \
  "$IMAGE_NAME"

echo
echo "==> Waiting for http://127.0.0.1:${HOST_PORT}/"
for attempt in $(seq 1 30); do
  if curl --fail --silent --show-error "http://127.0.0.1:${HOST_PORT}/" >/dev/null; then
    test -f "$database_dir/dealscanner.db"
    echo "Production container smoke test passed."
    exit 0
  fi

  if [[ "$(docker inspect --format '{{.State.Running}}' "$CONTAINER_NAME")" != "true" ]]; then
    echo "Production container exited before becoming ready." >&2
    exit 1
  fi

  sleep 1
done

echo "Production container did not become ready within 30 seconds." >&2
exit 1
