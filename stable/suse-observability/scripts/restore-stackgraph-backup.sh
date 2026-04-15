#!/usr/bin/env bash
set -Eeuo pipefail

TMP_DIR=/tmp-data

export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

echo "=== Downloading StackGraph backup \"${BACKUP_FILE}\" from bucket \"${BACKUP_STACKGRAPH_BUCKET_NAME}\"..."

sts-toolbox aws s3 cp --endpoint "http://${S3_ENDPOINT}" --region us-east-1 "s3://${BACKUP_STACKGRAPH_BUCKET_NAME}/${BACKUP_STACKGRAPH_S3_PREFIX}${BACKUP_FILE}" "${TMP_DIR}/${BACKUP_FILE}"

echo "=== Importing StackGraph data from \"${BACKUP_FILE}\"..."
/opt/docker/bin/stackstate-server -Dlogback.configurationFile=/opt/docker/etc_log/logback.xml -import "${TMP_DIR}/${BACKUP_FILE}" "${FORCE_DELETE}"
echo "==="
