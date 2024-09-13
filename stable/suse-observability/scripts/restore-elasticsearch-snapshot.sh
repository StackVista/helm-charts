#!/usr/bin/env bash
set -Eeuo pipefail
if [ -z "${INDICES_TO_RESTORE}" ] ; then
    INDICES_TO_RESTORE=${BACKUP_ELASTICSEARCH_SCHEDULED_INDICES}
fi
echo "=== Restoring ElasticSearch snapshot \"${SNAPSHOT_NAME}\" (indices = \"${INDICES_TO_RESTORE}\") from snapshot repository \"${BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_REPOSITORY_NAME}\" in bucket \"${BACKUP_ELASTICSEARCH_BUCKET_NAME}\"..."
SC=$(curl --request POST "http://${ELASTICSEARCH_ENDPOINT}/_snapshot/${BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_REPOSITORY_NAME}/${SNAPSHOT_NAME}/_restore?wait_for_completion=true&pretty" \
    --data "{\"indices\": \"${INDICES_TO_RESTORE}\"}" -H 'Content-Type: application/json' \
    --silent --output /dev/stderr --write-out "%{http_code}")
if [ "$SC" -ne 200 ]; then exit 1; fi
echo "==="
