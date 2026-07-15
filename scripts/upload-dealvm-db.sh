#!/usr/bin/env bash
set -euo pipefail

host="${1:-dealvm}"
repository_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
database_file="${DATABASE_FILE:-${repository_dir}/lbww-scanner-instance.db}"
remote_dir="/home/ubuntu/dealscanner"
remote_database="${remote_dir}/data/dealscanner.db"

if [[ ! -f "${database_file}" ]]; then
  echo "SQLite database does not exist: ${database_file}" >&2
  exit 1
fi

# Stop application writes before the one-time migration, then atomically replace
# the VM database. The previous database is retained as a rollback copy.
ssh "${host}" "sudo systemctl stop dealscanner || true; mkdir -p '${remote_dir}/data'; if test -f '${remote_database}'; then cp '${remote_database}' '${remote_database}.before-migration'; fi"
rsync -az --partial --progress "${database_file}" "${host}:${remote_database}.incoming"
ssh "${host}" "mv '${remote_database}.incoming' '${remote_database}'; rm -f '${remote_database}-wal' '${remote_database}-shm'; chmod 600 '${remote_database}'"

echo "Uploaded ${database_file} to ${host}:${remote_database}. Run scripts/deploy-dealvm.sh to start the native application."
