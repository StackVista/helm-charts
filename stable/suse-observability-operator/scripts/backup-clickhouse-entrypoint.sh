#!/usr/bin/env bash
set -Eeuo pipefail

# This script is an entrypoint of a sidecar container of ClickHouse.
# We want to backup only the first instance of each replica - so with we need that kind of script to check if a Pod is the first one from the StatefulSet.

if [[ "${CLICKHOUSE_REPLICA_ID}" == *-0-0 ]]; then
  clickhouse-backup --config /etc/clickhouse-backup.yaml server
else
  echo "This is a stub container doing nothing. For backup logs please check the first Pod of the StatefulSet where the backups are performed."
  sleep inf
fi
