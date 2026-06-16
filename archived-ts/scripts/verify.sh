#!/usr/bin/env bash
set -euo pipefail

pnpm run build
pnpm run lint
pnpm run test
