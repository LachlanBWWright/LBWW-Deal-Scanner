#!/usr/bin/env bash
set -euo pipefail

host="${1:-dealvm}"
repository_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
destination_file="${DATABASE_FILE:-${repository_dir}/dealscanner-vm-backup.db}"
remote_dir="/home/ubuntu/dealscanner"
remote_database="${remote_dir}/data/dealscanner.db"
remote_snapshot="${remote_dir}/data/dealscanner.download.db"

# SQLite's online backup command produces a consistent file even when the app
# is writing in WAL mode; copying the database file directly would not.
ssh "${host}" "if ! command -v sqlite3 >/dev/null; then sudo apt-get update && sudo apt-get install -y sqlite3; fi; test -f '${remote_database}'; rm -f '${remote_snapshot}'; sqlite3 '${remote_database}' \".backup '${remote_snapshot}'\"; chmod 600 '${remote_snapshot}'"
rsync -az --partial --progress "${host}:${remote_snapshot}" "${destination_file}"
ssh "${host}" "rm -f '${remote_snapshot}'"

echo "Downloaded a consistent SQLite backup to ${destination_file}."
