#!/usr/bin/env bash
set -euo pipefail

readonly SCRIPT_DIR="$(CDPATH='' cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=_common.sh
source "$SCRIPT_DIR/_common.sh"

require_command go

run_step "Running Go tests" go -C "$REPO_ROOT/goserver" test ./...
