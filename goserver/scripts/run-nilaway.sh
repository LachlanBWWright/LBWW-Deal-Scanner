#!/usr/bin/env bash
set -euo pipefail

# Ensure we are in the goserver directory (one level up from scripts/)
CDPATH="" cd -- "$(dirname -- "$0")/.."

# Ensure go bin directory is in PATH
export PATH="$(go env GOPATH)/bin:$PATH"

if ! command -v nilaway &> /dev/null; then
    echo "nilaway not found, installing..."
    go install go.uber.org/nilaway/cmd/nilaway@latest
fi

echo "Running nilaway..."
nilaway ./...
