#!/usr/bin/env bash
set -euo pipefail

# Ensure we are in the goserver directory
CDPATH="" cd -- "$(dirname -- "$0")"

if ! command -v nilaway &> /dev/null; then
    echo "nilaway not found, installing..."
    go install go.uber.org/nilaway/cmd/nilaway@latest
fi

echo "Running nilaway..."
nilaway ./...
