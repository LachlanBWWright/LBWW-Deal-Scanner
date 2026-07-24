#!/usr/bin/env bash
set -euo pipefail

readonly SCRIPT_DIR="$(CDPATH='' cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

"$SCRIPT_DIR/lint.sh"
"$SCRIPT_DIR/test.sh"
"$SCRIPT_DIR/build.sh"
