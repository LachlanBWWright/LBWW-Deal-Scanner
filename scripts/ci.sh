#!/usr/bin/env bash
set -euo pipefail

readonly SCRIPT_DIR="$(CDPATH='' cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=_common.sh
source "$SCRIPT_DIR/_common.sh"

require_command go
require_command pnpm
require_command docker

run_step "Installing locked frontend dependencies" pnpm --dir "$REPO_ROOT/frontend" install --frozen-lockfile
run_step "Downloading locked Go dependencies" go -C "$REPO_ROOT/goserver" mod download

"$SCRIPT_DIR/verify.sh"
"$SCRIPT_DIR/prod-smoke.sh"
