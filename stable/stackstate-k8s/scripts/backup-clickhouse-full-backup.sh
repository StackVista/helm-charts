#!/usr/bin/env bash
set -Eeuo pipefail

# The script creates a full backup of ClickHouse

now="$(date +'%Y-%m-%dT%H-%M-%S')"
backup_name="full_$now"
clickhouse-backup --config /etc/clickhouse-backup.yaml create_remote "$backup_name"
