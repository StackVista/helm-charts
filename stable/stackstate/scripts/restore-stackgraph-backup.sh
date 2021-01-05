#!/usr/bin/env bash
set -Eeuo pipefail

export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

echo "=== Downloading StackGraph backup \"${BACKUP_FILE}\" from bucket \"${BACKUP_STACKGRAPH_BUCKET_NAME}\"..."
aws --endpoint-url "http://${MINIO_ENDPOINT}" s3 cp "s3://${BACKUP_STACKGRAPH_BUCKET_NAME}/${BACKUP_FILE}" "/tmp/${BACKUP_FILE}"

echo "=== Importing StackGraph data from \"${BACKUP_FILE}\"..."
/opt/docker/bin/stackstate-server -import "/tmp/${BACKUP_FILE}"
