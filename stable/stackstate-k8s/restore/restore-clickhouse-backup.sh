#!/usr/bin/env bash
set -Eeuo pipefail

if [ "$#" -le 0 ]; then
  echo "Required backup name to restore, like 'incremental_2024-06-14T15-45-00'"
  exit 1
fi

BACKUP_NAME=$1

kubectl exec "stackstate-clickhouse-shard0-0" -c backup -- bash -c "clickhouse-backup --config /etc/clickhouse-backup.yaml restore_remote ${BACKUP_NAME}"
echo "Data restored"
