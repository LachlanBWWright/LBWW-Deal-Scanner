#!/usr/bin/env bash

# Shared helpers for repository command scripts. This file is sourced, not run.
set -euo pipefail

readonly REPO_ROOT="$(CDPATH='' cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"

require_command() {
  local command_name="$1"

  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "Required command not found: $command_name" >&2
    return 1
  fi
}

run_step() {
  local description="$1"
  shift

  echo
  echo "==> $description"
  "$@"
}
