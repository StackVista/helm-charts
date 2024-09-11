#!/usr/bin/env bash
set -Eeuo pipefail

kubectl exec "suse-observability-clickhouse-shard0-0" -c backup -- bash -c "LOG_LEVEL=error clickhouse-backup --config /etc/clickhouse-backup.yaml list remote"
