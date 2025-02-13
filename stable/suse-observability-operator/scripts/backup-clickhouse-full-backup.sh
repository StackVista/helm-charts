#!/usr/bin/env bash
set -Eeuo pipefail

# The script creates a full backup of ClickHouse

now="$(date +'%Y-%m-%dT%H-%M-%S')"
backup_name="full_$now"

# shellcheck disable=SC2140
curl -v -s -X POST --fail-with-body http://suse-observability-clickhouse-backup:7171/backup/create?name="$backup_name"\&table="$BACKUP_TABLES"\&callback="http://localhost:7171/backup/upload/$backup_name"
