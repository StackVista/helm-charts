#!/usr/bin/env bash
set -Eeuo pipefail

# The script creates an incremental backup of ClickHouse
# If there isn't previous backup (it is the first one) then execute the full one

now="$(date +'%Y-%m-%dT%H-%M-%S')"

set +e
latest="$(LOG_LEVEL=error clickhouse-backup --config /etc/clickhouse-backup.yaml list remote latest)"
latest_backup_exists=$?
set -e

if [[ -z "${BACKUP_TABLES}" ]]; then
  tables=""
else
  tables="--tables ${BACKUP_TABLES}"
fi

if [ $latest_backup_exists -eq 0 ]; then
  backup_name="incremental_$now"
  # shellcheck disable=SC2086
  clickhouse-backup --config /etc/clickhouse-backup.yaml create_remote --diff-from-remote "$latest" ${tables} "$backup_name"
else
  echo "Not found any previous backup, starting full backup"
  backup_name="full_$now"
  # shellcheck disable=SC2086
  clickhouse-backup --config /etc/clickhouse-backup.yaml create_remote ${tables} "$backup_name"
fi
