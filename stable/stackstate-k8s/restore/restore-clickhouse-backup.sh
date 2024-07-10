#!/usr/bin/env bash
set -Eeuo pipefail

if [ "$#" -le 0 ]; then
  echo "Required backup name to restore, like 'incremental_2024-06-14T15-45-00'"
  exit 1
fi

BACKUP_NAME=$1

collectorReplicas=$(kubectl get statefulset "stackstate-otel-collector" --output=jsonpath="{.spec.replicas}")
kubectl scale statefulsets "stackstate-otel-collector" --replicas=0
set +e
kubectl exec "stackstate-clickhouse-shard0-0" -c backup -- bash -c "clickhouse-backup --config /etc/clickhouse-backup.yaml restore_remote ${BACKUP_NAME}"
restoreStatus=$?
set -e
kubectl scale statefulsets "stackstate-otel-collector" --replicas="${collectorReplicas}"

if [ $restoreStatus -eq 0 ]
then
  echo "Data restored"
  exit 0
else
  echo "[ERROR] Restore failed, please check logs above"
  exit 1
fi
