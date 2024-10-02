#!/usr/bin/env bash
set -Eeuo pipefail

export BACKUP_DIR=/settings-backup-data
export TMP_DIR=/tmp

RESTORE_FILE="${TMP_DIR}/${BACKUP_FILE}"

if [ -f "${BACKUP_DIR}/${BACKUP_FILE}" ]; then
  cp "${BACKUP_DIR}/${BACKUP_FILE}" "${RESTORE_FILE}"

elif [ "$BACKUP_CONFIGURATION_UPLOAD_REMOTE" == "true" ]; then
  export AWS_ACCESS_KEY_ID
  AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
  export AWS_SECRET_ACCESS_KEY
  AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

  echo "=== Downloading Settings backup \"${BACKUP_FILE}\" from bucket \"${BACKUP_CONFIGURATION_BUCKET_NAME}\"..."
  sts-toolbox aws s3 --endpoint "http://${MINIO_ENDPOINT}" --region minio cp "s3://${BACKUP_CONFIGURATION_BUCKET_NAME}/${BACKUP_CONFIGURATION_S3_PREFIX}${BACKUP_FILE}" "${RESTORE_FILE}"
fi

 if [ ! -f  "${RESTORE_FILE}" ]; then
  echo "=== Backup file not found (\"${RESTORE_FILE}\"), exiting..."
  exit 1
 fi

echo "=== Waiting for backup file to be downloaded..."
while true; do
  sleep 1
done
echo "==="
