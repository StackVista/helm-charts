#!/usr/bin/env bash
set -Eeuo pipefail

export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

echo "=== Listing StackGraph backups in bucket \"${BACKUP_CONFIGURATION_BUCKET_NAME}\"..."
sts-toolbox aws s3 ls --endpoint "http://${MINIO_ENDPOINT}" --region minio --bucket "${BACKUP_CONFIGURATION_BUCKET_NAME}" --prefix "${BACKUP_CONFIGURATION_S3_PREFIX}"
echo "==="
