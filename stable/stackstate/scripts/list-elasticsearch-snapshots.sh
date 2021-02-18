#!/usr/bin/env bash
set -Eeuo pipefail

echo "=== Listing ElasticSearch snapshots in snapshot repository \"${BACKUP_ELASTICSEARCH_SNAPSHOT_REPOSITORY_NAME}\" in bucket \"${BACKUP_ELASTICSEARCH_BUCKET_NAME}\"..."
curl --request GET "http://${ELASTICSEARCH_ENDPOINT}/_snapshot/${BACKUP_ELASTICSEARCH_SNAPSHOT_REPOSITORY_NAME}/_all" --fail --silent --show-error | jq -r ".snapshots[] | .snapshot"
echo "==="
