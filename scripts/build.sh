#!/usr/bin/env bash
set -euo pipefail

readonly SCRIPT_DIR="$(CDPATH='' cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=_common.sh
source "$SCRIPT_DIR/_common.sh"

require_command go
require_command pnpm

build_dir="$(mktemp -d)"
cleanup() {
  rm -rf "$build_dir"
}
trap cleanup EXIT

run_step "Building production frontend" pnpm --dir "$REPO_ROOT/frontend" run build -- --outDir "$build_dir/frontend"
run_step "Building production Go binary" go -C "$REPO_ROOT/goserver" build -o "$build_dir/dealscanner" ./cmd/dealscanner

echo
echo "Production frontend and Go server built successfully."
