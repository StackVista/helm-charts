#!/usr/bin/env bash
set -Eeuo pipefail

export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

# Fetches timestamp of latest successful backup of given VM instance
function fetchBackupDate() {
  BUCKET=$1
  INSTANCE_NAME=$2
  set +e
  BACKUP_DATE=$(aws --endpoint-url "http://${MINIO_ENDPOINT}" --region minio s3api head-object --bucket "$BUCKET" --key "$INSTANCE_NAME/backup_complete.ignore" --query "LastModified")
  BACKUP_EXISTS=$?
  set -e
  if [ $BACKUP_EXISTS -eq 0 ]; then
    echo "$INSTANCE_NAME: $BACKUP_DATE"
  elif [ $BACKUP_EXISTS -eq 254 ]; then
    echo "Backup not found for '$INSTANCE_NAME'"
  else
    echo "Unexpected status code $BACKUP_EXISTS"
  fi
}

echo "=== Fetching timestamps of last completed incremental backups"
fetchBackupDate "$BACKUP_VICTORIA_METRICS_0_BUCKET_NAME" "victoria-metrics-0"
fetchBackupDate "$BACKUP_VICTORIA_METRICS_1_BUCKET_NAME" "victoria-metrics-1"
echo "==="
