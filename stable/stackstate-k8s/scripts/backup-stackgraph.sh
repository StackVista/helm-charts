#!/usr/bin/env bash
set -Eeuo pipefail

export TMP_DIR=/tmp-data

export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

echo "=== Removing old StackGraph backups from temporary storage \"${TMP_DIR}\"..."
(cd "${TMP_DIR:?}" && rm -rf ./sts-backup-*)
echo "=== Listing contents of \"${TMP_DIR}\"..."
ls -lat "${TMP_DIR:?}"
eval "BACKUP_FILE=\"${BACKUP_STACKGRAPH_SCHEDULED_BACKUP_NAME_TEMPLATE}\""
echo "=== Exporting StackGraph data to \"${BACKUP_FILE}\"..."
/opt/docker/bin/stackstate-server -Dlogback.configurationFile=/opt/docker/etc_log/logback.groovy -export "${TMP_DIR}/${BACKUP_FILE}"

if [ -f "${TMP_DIR}/${BACKUP_FILE}" ]; then
    echo "=== Uploading StackGraph backup \"${BACKUP_FILE}\" to bucket \"${BACKUP_STACKGRAPH_BUCKET_NAME}\"..."
    aws --endpoint-url "http://${MINIO_ENDPOINT}" --region minio s3 cp "${TMP_DIR}/${BACKUP_FILE}" "s3://${BACKUP_STACKGRAPH_BUCKET_NAME}/${BACKUP_FILE}"
    echo "=== Expiring old StackGraph backups..."
    export BACKUP_BUCKET_NAME=$BACKUP_STACKGRAPH_BUCKET_NAME
    export BACKUP_SCHEDULED_BACKUP_NAME_PARSE_REGEXP=$BACKUP_STACKGRAPH_SCHEDULED_BACKUP_NAME_PARSE_REGEXP
    export BACKUP_SCHEDULED_BACKUP_DATETIME_PARSE_FORMAT=$BACKUP_STACKGRAPH_SCHEDULED_BACKUP_DATETIME_PARSE_FORMAT
    export BACKUP_SCHEDULED_BACKUP_RETENTION_TIME_DELTA=$BACKUP_STACKGRAPH_SCHEDULED_BACKUP_RETENTION_TIME_DELTA
    /backup-restore-scripts/expire-s3-backups.py
else
    echo "=== Export failed. Backup file \"${TMP_DIR}/${BACKUP_FILE}\" does not exist."
    exit 1
fi
