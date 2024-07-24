#!/usr/bin/env bash
set -Eeuo pipefail

if [ "$#" -le 0 ]; then
  echo "Required backup name to restore, like 'incremental_2024-06-14T15-45-00'"
  exit 1
fi

BACKUP_NAME=$1

echo "=== Scaling down all OTel collectors"
collectorReplicas=$(kubectl get statefulset "stackstate-otel-collector" --output=jsonpath="{.spec.replicas}")
kubectl scale statefulsets "stackstate-otel-collector" --replicas=0

echo "=== Allowing pods to terminate"
sleep 15

echo "=== Starting the restore operation"
kubectl exec "stackstate-clickhouse-shard0-0" -c backup -- bash -c "curl -s -X POST http://localhost:7171/backup/download/${BACKUP_NAME}?callback=http://localhost:7171/backup/restore/${BACKUP_NAME}"

echo "=== Restore operation is async, waiting until completion ..."
until kubectl exec "stackstate-clickhouse-shard0-0" -c backup -- bash -c "curl -s localhost:7171/backup/actions?filter=restore | jq 'select(.command==\"restore ${BACKUP_NAME}\")' | jq -s . | jq -e -r '.[-1].status == \"success\"'"
do
  echo "waiting..."
  sleep 5
done

# ClickHouse has a problem with Mark cache after deletion of a table, it can throw an error 'Unknown codec family code'
# The solution is to drop the cache on ALL nodes
for clickhouse_pod in $(kubectl get pods -l app.kubernetes.io/component=clickhouse -o json | jq -r '.items[].metadata.name')
do
  kubectl exec "$clickhouse_pod" -c clickhouse -- bash -c "/app/post_restore.sh"
done

echo "=== Scaling up all OTel collectors"
kubectl scale statefulsets "stackstate-otel-collector" --replicas="${collectorReplicas}"
echo "=== Make sure that there is at least on replica of opentelemetry-collector"
