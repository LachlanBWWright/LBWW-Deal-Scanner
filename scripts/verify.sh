#!/usr/bin/env bash
set -euo pipefail

npm run build
npm run lint
npm test
