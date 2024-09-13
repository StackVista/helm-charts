#!/usr/bin/env bash
set -Eeuo pipefail

# The script creates an incremental backup of ClickHouse
# If there isn't previous backup (it is the first one) then execute the full one

now="$(date +'%Y-%m-%dT%H-%M-%S')"

latest_remote_backup="$(curl -s http://suse-observability-clickhouse-backup:7171/backup/list/remote | jq -s . | jq -r '.[-1].name')"

if [[ "$latest_remote_backup" = "null" ]]; then
  echo "not found any backup, create the first full backup"
  backup_name="full_$now"
  # shellcheck disable=SC2140
  curl -v -s -X POST --fail-with-body http://suse-observability-clickhouse-backup:7171/backup/create?name="$backup_name"\&table="$BACKUP_TABLES"\&callback="http://localhost:7171/backup/upload/$backup_name"
else
  echo "create an incremental backup, the parent backup $latest_remote_backup}"
  backup_name="incremental_$now"
  # shellcheck disable=SC2140
  curl -v -s -X POST --fail-with-body http://suse-observability-clickhouse-backup:7171/backup/create?name="$backup_name"\&table="$BACKUP_TABLES"\&callback="http://localhost:7171/backup/upload/$backup_name?diff-from-remote=$latest_remote_backup"
fi
