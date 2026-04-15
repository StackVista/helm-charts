#!/usr/bin/env bash
set -Eeuo pipefail

export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

echo "=== Listing StackGraph backups in bucket \"${BACKUP_STACKGRAPH_BUCKET_NAME}\"..."
sts-toolbox aws s3 ls --endpoint "http://${S3_ENDPOINT}" --region us-east-1 --bucket "${BACKUP_STACKGRAPH_BUCKET_NAME}" --prefix "${BACKUP_STACKGRAPH_S3_PREFIX}" | grep -v "\.[0-9]\+$"
echo "==="
