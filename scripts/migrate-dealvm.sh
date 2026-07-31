#!/usr/bin/env bash
set -euo pipefail

host="${1:-dealvm}"
if [[ $# -gt 0 ]]; then
  shift
fi

remote_dir="/home/ubuntu/dealscanner"
remote_database="${remote_dir}/data/dealscanner.db"
release_dir="${remote_dir}/current"
leave_stopped=false

usage() {
  cat <<'EOF'
Usage: scripts/migrate-dealvm.sh [host] [options]

Options:
  --release PATH    Release containing migrate-db (default: current release)
  --leave-stopped   Do not restart dealscanner after a successful migration
  -h, --help        Show this help

The script stops application writes, creates a timestamped SQLite online
backup, runs the release's migrate-db binary, verifies integrity, and restarts
the service unless --leave-stopped is supplied. A failed migration leaves the
database and backup in place and restarts the previously-running service.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --release)
      if [[ $# -lt 2 ]]; then
        echo "--release requires a path" >&2
        exit 2
      fi
      release_dir=$2
      shift 2
      ;;
    --leave-stopped)
      leave_stopped=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

log() {
  printf '[%s] %s\n' "$(date '+%H:%M:%S')" "$*"
}

service_was_active=false
migration_succeeded=false

restore_service_on_exit() {
  local exit_code=$?
  if [[ "${migration_succeeded}" != true && "${service_was_active}" == true ]]; then
    log "Migration failed; restarting the previously-running service"
    ssh "${host}" "sudo systemctl start dealscanner" || true
  fi
  exit "${exit_code}"
}
trap restore_service_on_exit EXIT

log "Checking migration inputs on ${host}"
ssh "${host}" \
  "test -f '${remote_database}' && test -x '${release_dir}/migrate-db'"

if ssh "${host}" "systemctl is-active --quiet dealscanner"; then
  service_was_active=true
fi

log "Stopping Dealscanner to prevent database writes"
ssh "${host}" "sudo systemctl stop dealscanner"

backup_path="$(
  ssh "${host}" "REMOTE_DATABASE='${remote_database}' python3 - <<'PY'
import datetime
import os
import sqlite3

source_path = os.environ['REMOTE_DATABASE']
timestamp = datetime.datetime.now(datetime.timezone.utc).strftime('%Y%m%dT%H%M%SZ')
backup_path = f'{source_path}.backup-{timestamp}'

source = sqlite3.connect(source_path, timeout=60)
destination = sqlite3.connect(backup_path)
with destination:
    source.backup(destination)
result = destination.execute('PRAGMA integrity_check').fetchone()
destination.close()
source.close()

if result is None or result[0] != 'ok':
    raise SystemExit(f'backup integrity check failed: {result}')

os.chmod(backup_path, 0o600)
print(backup_path)
PY"
)"
log "Created verified backup: ${backup_path}"

log "Running schema migration from ${release_dir}"
ssh "${host}" \
  "cd '${release_dir}' && DATABASE_URL='file:${remote_database}' TURSO_AUTH_TOKEN='' ./migrate-db"

log "Verifying migrated database integrity and foreign keys"
ssh "${host}" "REMOTE_DATABASE='${remote_database}' python3 - <<'PY'
import os
import sqlite3

database = sqlite3.connect(os.environ['REMOTE_DATABASE'], timeout=60)
integrity = database.execute('PRAGMA integrity_check').fetchone()
foreign_keys = database.execute('PRAGMA foreign_key_check').fetchall()
database.close()

if integrity is None or integrity[0] != 'ok':
    raise SystemExit(f'integrity check failed: {integrity}')
if foreign_keys:
    raise SystemExit(f'foreign key check returned {len(foreign_keys)} violation(s)')
print('Database verification passed.')
PY"

migration_succeeded=true
if [[ "${leave_stopped}" == true ]]; then
  log "Migration completed; service intentionally left stopped"
elif [[ "${service_was_active}" == true ]]; then
  log "Restarting Dealscanner"
  ssh "${host}" "sudo systemctl start dealscanner"
fi

trap - EXIT
log "Migration completed successfully; rollback backup: ${backup_path}"
