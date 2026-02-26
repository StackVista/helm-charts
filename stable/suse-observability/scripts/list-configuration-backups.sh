#!/usr/bin/env bash
set -Eeuo pipefail

# Set up AWS credentials for S3Proxy access
export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

echo "=== Listing Settings backups in local settings bucket \"${S3_BUCKET_SETTINGS}\"..."
sts-toolbox aws s3 ls --endpoint "http://${S3_ENDPOINT}" --region us-east-1 --bucket "${S3_BUCKET_SETTINGS}" --prefix "" 2>/dev/null || echo "(no backups found)"
echo "==="

if [ "$BACKUP_CONFIGURATION_UPLOAD_REMOTE" == "true" ]; then
  echo "=== Listing settings backups in remote storage bucket \"${BACKUP_CONFIGURATION_BUCKET_NAME}\"..."
  sts-toolbox aws s3 ls --endpoint "http://${S3_ENDPOINT}" --region us-east-1 --bucket "${BACKUP_CONFIGURATION_BUCKET_NAME}" --prefix "${BACKUP_CONFIGURATION_S3_PREFIX}" 2>/dev/null || echo "(no backups found)"
  echo "==="
fi
