#!/usr/bin/env bash
set -Eeuo pipefail

# This script is an entrypoint of a sidecar container of ClickHouse.
# We have to start if backups are enabled and if "yes" then we can start supercronic - to create backups periodically,
# otherwise we have to execute infinitive No-op like `sleep inf` (to not restart container).
# Why?
# - we can use the official helm chart and add sidecar configuration instead of forking it
# - eventually we want to backup only the first instance of each replica - so with we need that kind of script to check if a Pod is the first one from the StatefulSet.

if [ "${BACKUP_CLICKHOUSE_ENABLED}" == "true" ]; then
  if [[ "${CLICKHOUSE_REPLICA_ID}" == *-0 ]]; then
    clickhouse-backup --config /etc/clickhouse-backup.yaml server
  else
    echo "This is a stub container doing nothing. For backup logs please check the first Pod of the StatefulSet where the backups are performed."
    sleep inf
  fi
else
  echo "Backup is disabled"
  sleep inf
fi
