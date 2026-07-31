#!/usr/bin/env bash
set -euo pipefail

host="${1:-dealvm}"
repository_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
database_file="${DATABASE_FILE:-${repository_dir}/lbww-scanner-instance.db}"
remote_dir="/home/ubuntu/dealscanner"
remote_database="${remote_dir}/data/dealscanner.db"
incoming_database="${remote_database}.incoming"

if [[ ! -f "${database_file}" ]]; then
  echo "SQLite database does not exist: ${database_file}" >&2
  exit 1
fi

# This is intentionally separate from normal schema migration because it
# replaces production data. Validate the incoming file before stopping writes.
rsync -az --partial --progress "${database_file}" "${host}:${incoming_database}"
ssh "${host}" "REMOTE_DATABASE='${incoming_database}' python3 - <<'PY'
import os
import sqlite3

database = sqlite3.connect(os.environ['REMOTE_DATABASE'], timeout=60)
result = database.execute('PRAGMA integrity_check').fetchone()
database.close()
if result is None or result[0] != 'ok':
    raise SystemExit(f'incoming database integrity check failed: {result}')
print('Incoming database integrity check passed.')
PY"

ssh "${host}" "sudo systemctl stop dealscanner; mkdir -p '${remote_dir}/data'; if test -f '${remote_database}'; then timestamp=\$(date -u +%Y%m%dT%H%M%SZ); cp '${remote_database}' '${remote_database}.before-upload-'\${timestamp}; fi; mv '${incoming_database}' '${remote_database}'; rm -f '${remote_database}-wal' '${remote_database}-shm'; chmod 600 '${remote_database}'"

echo "Uploaded ${database_file} to ${host}:${remote_database}."
echo "Run scripts/migrate-dealvm.sh ${host} to migrate, verify, and restart the service."
