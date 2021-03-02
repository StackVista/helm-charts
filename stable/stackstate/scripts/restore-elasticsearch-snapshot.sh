#!/usr/bin/env bash
set -Eeuo pipefail

echo "=== Restoring ElasticSearch snapshot \"${SNAPSHOT_NAME}\" from snapshot repository \"${BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_REPOSITORY_NAME}\" in bucket \"${BACKUP_ELASTICSEARCH_BUCKET_NAME}\"..."
SC=$(curl --request POST "http://${ELASTICSEARCH_ENDPOINT}/_snapshot/${BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_REPOSITORY_NAME}/${SNAPSHOT_NAME}/_restore?wait_for_completion=true&pretty" \
    --silent --output /dev/stderr --write-out "%{http_code}")
if [ "$SC" -ne 200 ]; then exit 1; fi
echo "==="
