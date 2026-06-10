#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

is_port_in_use() {
  local port="$1"
  if command -v ss >/dev/null 2>&1; then
    ss -ltn "( sport = :$port )" | tail -n +2 | grep -q .
    return
  fi

  if command -v lsof >/dev/null 2>&1; then
    lsof -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1
    return
  fi

  # If neither tool is available, assume free and let process startup report.
  return 1
}

find_free_port() {
  local port="$1"
  while is_port_in_use "$port"; do
    port=$((port + 1))
  done
  echo "$port"
}

cleanup() {
  trap - INT TERM EXIT
  if [[ -n "${SERVER_PID:-}" ]]; then
    kill "$SERVER_PID" 2>/dev/null || true
  fi
  if [[ -n "${FRONTEND_PID:-}" ]]; then
    kill "$FRONTEND_PID" 2>/dev/null || true
  fi
  wait 2>/dev/null || true
}

trap cleanup INT TERM EXIT

run_with_prefix() {
  local name="$1"
  local dir="$2"
  shift 2

  cd "$ROOT_DIR/$dir"
  "$@" 2>&1 | sed "s/^/[$name] /"
}

SERVER_PORT="${API_PORT:-3001}"
FRONTEND_PORT="${FRONTEND_PORT:-5173}"
API_SECRET_VALUE="${API_SECRET:-dealscanner-dev-secret}"
ENABLE_TESTING_VALUE="${ENABLE_TESTING_API:-true}"

SERVER_PORT="$(find_free_port "$SERVER_PORT")"
FRONTEND_PORT="$(find_free_port "$FRONTEND_PORT")"

echo "Starting archived TypeScript backend and frontend. Press Ctrl+C to stop both."
echo "Using API_PORT=$SERVER_PORT and frontend port $FRONTEND_PORT"
echo "Testing API enabled: $ENABLE_TESTING_VALUE"

run_with_prefix "archived-ts" "archived-ts" env API_PORT="$SERVER_PORT" API_SECRET="$API_SECRET_VALUE" ENABLE_TESTING_API="$ENABLE_TESTING_VALUE" pnpm run dev &
SERVER_PID=$!

run_with_prefix "frontend" "frontend" env API_PORT="$SERVER_PORT" VITE_API_SECRET="$API_SECRET_VALUE" pnpm run dev -- --port "$FRONTEND_PORT" &
FRONTEND_PID=$!

wait -n "$SERVER_PID" "$FRONTEND_PID"

echo "A dev process exited, shutting down the other process..."
cleanup
