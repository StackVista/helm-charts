#!/usr/bin/env bash
set -Eeuo pipefail

TMP_DIR=/tmp-data

export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

echo "=== Downloading StackGraph backup \"${BACKUP_FILE}\" from bucket \"${BACKUP_STACKGRAPH_BUCKET_NAME}\"..."
aws --endpoint-url "http://${MINIO_ENDPOINT}" --region minio s3 cp "s3://${BACKUP_STACKGRAPH_BUCKET_NAME}/${BACKUP_FILE}" "${TMP_DIR}/${BACKUP_FILE}"

echo "=== Importing StackGraph data from \"${BACKUP_FILE}\"..."
/opt/docker/bin/stackstate-server -Dlogback.configurationFile=/opt/docker/etc_log/logback.groovy -import "${TMP_DIR}/${BACKUP_FILE}" "${FORCE_DELETE}"
echo "==="
