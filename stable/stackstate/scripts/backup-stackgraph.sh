#!/usr/bin/env bash
set -Eeuo pipefail

export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

eval "BACKUP_FILE=\"${BACKUP_STACKGRAPH_BACKUP_NAME_TEMPLATE}\""
echo "=== Exporting StackGraph data to \"${BACKUP_FILE}\"..."
/opt/docker/bin/stackstate-server -export "/tmp/${BACKUP_FILE}"
echo "=== Uploading StackGraph backup \"${BACKUP_FILE}\" to bucket \"${BACKUP_STACKGRAPH_BUCKET_NAME}\"..."
aws --endpoint-url "http://${MINIO_ENDPOINT}" s3 cp "/tmp/${BACKUP_FILE}" "s3://${BACKUP_STACKGRAPH_BUCKET_NAME}/${BACKUP_FILE}"

echo "=== Expiring old StackGraph backups..."
/backup-restore-scripts/expire-stackgraph-backups.py
