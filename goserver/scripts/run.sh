#!/usr/bin/env bash
set -euo pipefail

# Ensure we are in the goserver directory (one level up from scripts/)
CDPATH="" cd -- "$(dirname -- "$0")/.."

echo "Starting DealScanner..."
go run cmd/dealscanner/main.go
