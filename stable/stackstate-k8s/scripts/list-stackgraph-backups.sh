#!/usr/bin/env bash
set -Eeuo pipefail

export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

echo "=== Listing StackGraph backups in bucket \"${BACKUP_STACKGRAPH_BUCKET_NAME}\"..."
if [ "${BACKUP_STACKGRAPH_ARCHIVE_SPLIT_SIZE:-0}" == "0" ]; then
  # Exclude all multipart arhives if present
  aws --endpoint-url "http://${MINIO_ENDPOINT}" --region minio s3api list-objects-v2 --bucket "${BACKUP_STACKGRAPH_BUCKET_NAME}" --query "Contents[].[Key]" --output text | grep -v "\.[0-9]\+$"
else
  # Only show the first file of the multipart arhive
  aws --endpoint-url "http://${MINIO_ENDPOINT}" --region minio s3api list-objects-v2 --bucket "${BACKUP_STACKGRAPH_BUCKET_NAME}" --query "Contents[].[Key]" --output text | grep "\.00$"
fi
echo "==="
