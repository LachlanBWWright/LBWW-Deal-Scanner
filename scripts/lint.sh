#!/usr/bin/env bash
set -euo pipefail

readonly SCRIPT_DIR="$(CDPATH='' cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=_common.sh
source "$SCRIPT_DIR/_common.sh"

require_command go
require_command pnpm

run_step "Typechecking frontend" pnpm --dir "$REPO_ROOT/frontend" run typecheck
run_step "Linting frontend" pnpm --dir "$REPO_ROOT/frontend" run lint

run_step "Checking Go formatting" bash -c '
  cd "$1"
  unformatted="$(gofmt -l .)"
  if [[ -n "$unformatted" ]]; then
    echo "The following Go files need gofmt:" >&2
    echo "$unformatted" >&2
    exit 1
  fi
' _ "$REPO_ROOT/goserver"

run_step "Vetting Go code" go -C "$REPO_ROOT/goserver" vet ./...
run_step "Checking Go nil safety" bash -c 'cd "$1" && go run go.uber.org/nilaway/cmd/nilaway ./...' _ "$REPO_ROOT/goserver"
