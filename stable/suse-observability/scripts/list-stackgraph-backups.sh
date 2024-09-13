#!/usr/bin/env bash
set -Eeuo pipefail

export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

echo "=== Listing StackGraph backups in bucket \"${BACKUP_STACKGRAPH_BUCKET_NAME}\"..."
if [ "${BACKUP_STACKGRAPH_ARCHIVE_SPLIT_SIZE:-0}" == "0" ]; then
  # Exclude all multipart arhives if present
  sts-toolbox aws s3 ls --endpoint "http://${MINIO_ENDPOINT}" --region minio --bucket "${BACKUP_STACKGRAPH_BUCKET_NAME}" --prefix "${BACKUP_STACKGRAPH_S3_PREFIX}" | grep -v "\.[0-9]\+$"
else
  # Only show the first file of the multipart arhive
  sts-toolbox aws s3 ls --endpoint "http://${MINIO_ENDPOINT}" --region minio --bucket "${BACKUP_STACKGRAPH_BUCKET_NAME}" --prefix "${BACKUP_STACKGRAPH_S3_PREFIX}" | grep "\.00$"
fi
echo "==="
