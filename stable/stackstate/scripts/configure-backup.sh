#!/usr/bin/env bash
set -Eeuo pipefail

export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"
export MC_HOST_minio="http://${AWS_ACCESS_KEY_ID}:${AWS_SECRET_ACCESS_KEY}@${MINIO_ENDPOINT}"

if [ "${BACKUP_STACKGRAPH_ENABLED}" == "true" ]; then
    echo "=== Creating MinIO bucket \"${BACKUP_STACKGRAPH_BUCKET_NAME}\"..."
    mc mb --ignore-existing "minio/${BACKUP_STACKGRAPH_BUCKET_NAME}"
fi

if [ "${BACKUP_ELASTICSEARCH_ENABLED}" == "true" ]; then
    echo "=== Creating MinIO bucket \"${BACKUP_ELASTICSEARCH_BUCKET_NAME}\"..."
    mc mb --ignore-existing "minio/${BACKUP_ELASTICSEARCH_BUCKET_NAME}"

    echo "=== Configuring ElaticSearch snapshot repository \"${BACKUP_ELASTICSEARCH_SNAPSHOT_REPOSITORY_NAME}\" for bucket \"${BACKUP_ELASTICSEARCH_BUCKET_NAME}\"..."
    SC=$(curl --request PUT "http://${ELASTICSEARCH_ENDPOINT}/_snapshot/${BACKUP_ELASTICSEARCH_SNAPSHOT_REPOSITORY_NAME}?pretty" --header "Content-Type: application/json" --data "
    {
        \"type\": \"s3\",
        \"settings\": {
            \"bucket\": \"${BACKUP_ELASTICSEARCH_BUCKET_NAME}\",
            \"endpoint\": \"${MINIO_ENDPOINT}\",
            \"protocol\": \"http\",
            \"access_key\": \"${AWS_ACCESS_KEY_ID}\",
            \"secret_key\": \"${AWS_SECRET_ACCESS_KEY}\",
            \"path_style_access\": \"true\"
        }
    }" --silent --output /dev/stderr --write-out "%{http_code}")
    if [ "$SC" -ne 200 ]; then exit 1; fi

    echo "=== Configuring ElaticSearch snapshot lifecycle management policy \"${BACKUP_ELASTICSEARCH_SNAPSHOT_POLICY_NAME}\" for snapshot repository \"${BACKUP_ELASTICSEARCH_SNAPSHOT_REPOSITORY_NAME}\"..."
    SC=$(curl --request PUT "http://${ELASTICSEARCH_ENDPOINT}/_slm/policy/${BACKUP_ELASTICSEARCH_SNAPSHOT_POLICY_NAME}?pretty" --header "Content-Type: application/json" --data "
    {
        \"schedule\": \"${BACKUP_ELASTICSEARCH_SCHEDULE}\",
        \"name\": \"${BACKUP_ELASTICSEARCH_SNAPSHOT_NAME_TEMPLATE}\",
        \"repository\": \"${BACKUP_ELASTICSEARCH_SNAPSHOT_REPOSITORY_NAME}\",
        \"config\": {
            \"indices\": ${BACKUP_ELASTICSEARCH_INDICES},
            \"ignore_unavailable\": false,
            \"include_global_state\": false
        },
        \"retention\": {
            \"expire_after\": \"${BACKUP_ELASTICSEARCH_SNAPSHOT_RETENTION_EXPIRE_AFTER}\",
            \"min_count\": \"${BACKUP_ELASTICSEARCH_SNAPSHOT_RETENTION_MIN_COUNT}\",
            \"max_count\": \"${BACKUP_ELASTICSEARCH_SNAPSHOT_RETENTION_MAX_COUNT}\"
        }
    }" --silent --output /dev/stderr --write-out "%{http_code}")
    if [ "$SC" -ne 200 ]; then exit 1; fi
fi
