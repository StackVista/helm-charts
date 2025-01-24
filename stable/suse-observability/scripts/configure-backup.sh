#!/usr/bin/env bash
set -Eeuo pipefail

export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"
export MC_HOST_minio="http://${AWS_ACCESS_KEY_ID}:${AWS_SECRET_ACCESS_KEY}@${MINIO_ENDPOINT}"
MC_WITH_CONFIG_DIR="/work/.mc"
if [ "${BACKUP_STACKGRAPH_RESTORE_ENABLED}" == "true" ] || [ "${BACKUP_STACKGRAPH_SCHEDULED_ENABLED}" == "true" ]; then
    echo "=== Testing for existence of MinIO bucket \"${BACKUP_STACKGRAPH_BUCKET_NAME}\"..."
    if ! mc -C "$MC_WITH_CONFIG_DIR" ls "minio/${BACKUP_STACKGRAPH_BUCKET_NAME}/${BACKUP_STACKGRAPH_S3_PREFIX}" >/dev/null ; then
        if [ "${BACKUP_STACKGRAPH_SCHEDULED_ENABLED}" == "true" ]; then
            echo "=== Creating MinIO bucket \"${BACKUP_STACKGRAPH_BUCKET_NAME}\"..."
            mc -C "$MC_WITH_CONFIG_DIR" mb "minio/${BACKUP_STACKGRAPH_BUCKET_NAME}"
        else
            echo "=== ERROR: MinIO bucket \"${BACKUP_STACKGRAPH_BUCKET_NAME}\" does not exist. Restore functionality for StackGraph will not function correctly"
            exit 1
        fi
    fi
fi

if [ "${BACKUP_CONFIGURATION_RESTORE_ENABLED}" == "true" ] || [ "${BACKUP_CONFIGURATION_SCHEDULED_ENABLED}" == "true" ]; then
    echo "=== Testing for existence of MinIO bucket \"${BACKUP_CONFIGURATION_BUCKET_NAME}\"..."
    if ! mc -C "$MC_WITH_CONFIG_DIR" ls "minio/${BACKUP_CONFIGURATION_BUCKET_NAME}/${BACKUP_CONFIGURATION_S3_PREFIX}" >/dev/null ; then
        if [ "${BACKUP_CONFIGURATION_SCHEDULED_ENABLED}" == "true" ]; then
            echo "=== Creating MinIO bucket \"${BACKUP_CONFIGURATION_BUCKET_NAME}\"..."
            mc -C "$MC_WITH_CONFIG_DIR" mb "minio/${BACKUP_CONFIGURATION_BUCKET_NAME}"
        else
            echo "=== ERROR: MinIO bucket \"${BACKUP_CONFIGURATION_BUCKET_NAME}\" does not exist. Restore functionality for configuration will not function correctly"
            exit 1
        fi
    fi
fi

if [ "${BACKUP_VICTORIA_METRICS_0_RESTORE_ENABLED}" == "true" ] || [ "${BACKUP_VICTORIA_METRICS_0_ENABLED}" == "true" ]; then
    echo "=== Testing for existence of MinIO bucket \"${BACKUP_VICTORIA_METRICS_0_BUCKET_NAME}\"..."
    if ! mc -C "$MC_WITH_CONFIG_DIR" ls "minio/${BACKUP_VICTORIA_METRICS_0_BUCKET_NAME}/${BACKUP_VICTORIA_METRICS_1_S3_PREFIX}" >/dev/null ; then
        if [ "${BACKUP_VICTORIA_METRICS_0_ENABLED}" == "true" ]; then
            echo "=== Creating MinIO bucket \"${BACKUP_VICTORIA_METRICS_0_BUCKET_NAME}\"..."
            mc -C "$MC_WITH_CONFIG_DIR" mb "minio/${BACKUP_VICTORIA_METRICS_0_BUCKET_NAME}"
        else
            echo "=== ERROR: MinIO bucket \"${BACKUP_VICTORIA_METRICS_0_BUCKET_NAME}\" does not exist. Restore functionality for VictoriaMetrics_0 will not function correctly"
            exit 1
        fi
    fi
fi

if [ "${BACKUP_VICTORIA_METRICS_1_RESTORE_ENABLED}" == "true" ] || [ "${BACKUP_VICTORIA_METRICS_1_ENABLED}" == "true" ]; then
    echo "=== Testing for existence of MinIO bucket \"${BACKUP_VICTORIA_METRICS_1_BUCKET_NAME}\"..."
    if ! mc -C "$MC_WITH_CONFIG_DIR" ls "minio/${BACKUP_VICTORIA_METRICS_1_BUCKET_NAME}/${BACKUP_VICTORIA_METRICS_1_S3_PREFIX}" >/dev/null ; then
        if [ "${BACKUP_VICTORIA_METRICS_1_ENABLED}" == "true" ]; then
            echo "=== Creating MinIO bucket \"${BACKUP_VICTORIA_METRICS_1_BUCKET_NAME}\"..."
            mc -C "$MC_WITH_CONFIG_DIR" mb "minio/${BACKUP_VICTORIA_METRICS_1_BUCKET_NAME}"
        else
            echo "=== ERROR: MinIO bucket \"${BACKUP_VICTORIA_METRICS_1_BUCKET_NAME}\" does not exist. Restore functionality for VictoriaMetrics_1 will not function correctly"
            exit 1
        fi
    fi
fi

if [ "${BACKUP_CLICKHOUSE_RESTORE_ENABLED}" == "true" ] || [ "${BACKUP_CLICKHOUSE_SCHEDULED_ENABLED}" == "true" ]; then
    echo "=== Testing for existence of MinIO bucket \"${BACKUP_CLICKHOUSE_BUCKET_NAME}\"..."
    if ! mc -C "$MC_WITH_CONFIG_DIR" ls "minio/${BACKUP_CLICKHOUSE_BUCKET_NAME}/${BACKUP_CLICKHOUSE_S3_PREFIX}" >/dev/null ; then
        if [ "${BACKUP_CLICKHOUSE_SCHEDULED_ENABLED}" == "true" ]; then
            echo "=== Creating MinIO bucket \"${BACKUP_CLICKHOUSE_BUCKET_NAME}\"..."
            mc -C "$MC_WITH_CONFIG_DIR" mb "minio/${BACKUP_CLICKHOUSE_BUCKET_NAME}"
        else
            echo "=== ERROR: MinIO bucket \"${BACKUP_CLICKHOUSE_BUCKET_NAME}\" does not exist. Restore functionality for ClickHouse will not function correctly"
            exit 1
        fi
    fi
fi

if [ "${BACKUP_ELASTICSEARCH_RESTORE_ENABLED}" == "true" ] || [ "${BACKUP_ELASTICSEARCH_SCHEDULED_ENABLED}" == "true" ]; then
    echo "=== Testing for existence of MinIO bucket \"${BACKUP_ELASTICSEARCH_BUCKET_NAME}\"..."
    if ! mc -C "$MC_WITH_CONFIG_DIR" ls "minio/${BACKUP_ELASTICSEARCH_BUCKET_NAME}/${BACKUP_ELASTICSEARCH_S3_PREFIX}" >/dev/null ; then
        if [ "${BACKUP_ELASTICSEARCH_SCHEDULED_ENABLED}" == "true" ]; then
            echo "=== Creating MinIO bucket \"${BACKUP_ELASTICSEARCH_BUCKET_NAME}\"..."
            mc -C "$MC_WITH_CONFIG_DIR" mb "minio/${BACKUP_ELASTICSEARCH_BUCKET_NAME}"
        else
            echo "=== ERROR: MinIO bucket \"${BACKUP_ELASTICSEARCH_BUCKET_NAME}\" does not exist. Restore functionality for ElasticSearch will not function correctly"
            exit 1
        fi
    fi

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

    if [ "${BACKUP_ELASTICSEARCH_SCHEDULED_ENABLED}" == "true" ]; then
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
    fi
fi
