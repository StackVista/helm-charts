#!/usr/bin/env bash
set -Eeuo pipefail

# The script creates a full backup of ClickHouse

now="$(date +'%Y-%m-%dT%H-%M-%S')"
backup_name="full_$now"

if [[ -z "${BACKUP_TABLES}" ]]; then
  tables=""
else
  tables="--tables ${BACKUP_TABLES}"
fi

# shellcheck disable=SC2086
clickhouse-backup --config /etc/clickhouse-backup.yaml create_remote ${tables} "$backup_name"
