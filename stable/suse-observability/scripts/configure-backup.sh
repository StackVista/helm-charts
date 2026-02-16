#!/usr/bin/env bash
set -Eeuo pipefail

export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

echo "=== Configuring ElasticSearch snapshot repository \"${BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_REPOSITORY_NAME}\" for bucket \"${BACKUP_ELASTICSEARCH_BUCKET_NAME}\"..."
SC=$(curl --request PUT "http://${ELASTICSEARCH_ENDPOINT}/_snapshot/${BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_REPOSITORY_NAME}?pretty" --header "Content-Type: application/json" --data "
{
    \"type\": \"s3\",
    \"settings\": {
        \"bucket\": \"${BACKUP_ELASTICSEARCH_BUCKET_NAME}\",
        \"region\": \"minio\",
        \"endpoint\": \"${MINIO_ENDPOINT}\",
        \"base_path\": \"${BACKUP_ELASTICSEARCH_S3_PREFIX}\",
        \"protocol\": \"http\",
        \"access_key\": \"${AWS_ACCESS_KEY_ID}\",
        \"secret_key\": \"${AWS_SECRET_ACCESS_KEY}\",
        \"path_style_access\": \"true\"
    }
}" --silent --output /dev/stderr --write-out "%{http_code}")
if [ "$SC" -ne 200 ]; then exit 1; fi

echo "=== Configuring ElasticSearch snapshot lifecycle management policy \"${BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_POLICY_NAME}\" for snapshot repository \"${BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_REPOSITORY_NAME}\"..."
SC=$(curl --request PUT "http://${ELASTICSEARCH_ENDPOINT}/_slm/policy/${BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_POLICY_NAME}?pretty" --header "Content-Type: application/json" --data "
{
    \"schedule\": \"${BACKUP_ELASTICSEARCH_SCHEDULED_SCHEDULED}\",
    \"name\": \"${BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_NAME_TEMPLATE}\",
    \"repository\": \"${BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_REPOSITORY_NAME}\",
    \"config\": {
        \"indices\": \"${BACKUP_ELASTICSEARCH_SCHEDULED_INDICES}\",
        \"ignore_unavailable\": false,
        \"include_global_state\": false
    },
    \"retention\": {
        \"expire_after\": \"${BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_RETENTION_EXPIRE_AFTER}\",
        \"min_count\": \"${BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_RETENTION_MIN_COUNT}\",
        \"max_count\": \"${BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_RETENTION_MAX_COUNT}\"
    }
}" --silent --output /dev/stderr --write-out "%{http_code}")
if [ "$SC" -ne 200 ]; then exit 1; fi
