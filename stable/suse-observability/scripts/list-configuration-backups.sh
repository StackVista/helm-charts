#!/usr/bin/env bash
set -Eeuo pipefail

export BACKUP_DIR=/settings-backup-data
echo "=== Listing StackGraph backups in local persistent volume..."
# shellcheck disable=SC2010
ls -t "${BACKUP_DIR}" | grep -v 'lost+found'

echo "==="

if [ "$BACKUP_CONFIGURATION_UPLOAD_REMOTE" == "true" ]; then
  export AWS_ACCESS_KEY_ID
  AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
  export AWS_SECRET_ACCESS_KEY
  AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

  echo "=== Listing settings backups in storage bucket \"${BACKUP_CONFIGURATION_BUCKET_NAME}\"..."
  sts-toolbox aws s3 ls --endpoint "http://${MINIO_ENDPOINT}" --region minio --bucket "${BACKUP_CONFIGURATION_BUCKET_NAME}" --prefix "${BACKUP_CONFIGURATION_S3_PREFIX}"
  echo "==="
fi
