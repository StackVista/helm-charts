#!/usr/bin/env bash
set -Eeuo pipefail

export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

echo "=== Listing StackGraph backups in bucket \"${BACKUP_STACKGRAPH_BUCKET_NAME}\"..."
aws --endpoint-url "http://${MINIO_ENDPOINT}" s3api list-objects-v2 --bucket "${BACKUP_STACKGRAPH_BUCKET_NAME}" --query "Contents[].[Key]" --output text
