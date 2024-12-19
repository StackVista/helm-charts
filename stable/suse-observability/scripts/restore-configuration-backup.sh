#!/usr/bin/env bash
set -Eeuo pipefail

export BACKUP_DIR=/settings-backup-data
export TMP_DIR=/tmp

RESTORE_FILE="${BACKUP_DIR}/${BACKUP_FILE}"

if [ "$BACKUP_CONFIGURATION_UPLOAD_REMOTE" == "true" ] && [ ! -f "${RESTORE_FILE}" ]; then
  export AWS_ACCESS_KEY_ID
  AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
  export AWS_SECRET_ACCESS_KEY
  AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

  echo "=== Downloading Settings backup \"${BACKUP_FILE}\" from bucket \"${BACKUP_CONFIGURATION_BUCKET_NAME}\"..."
  sts-toolbox aws s3 --endpoint "http://${MINIO_ENDPOINT}" --region minio cp "s3://${BACKUP_CONFIGURATION_BUCKET_NAME}/${BACKUP_CONFIGURATION_S3_PREFIX}${BACKUP_FILE}" "${TMP_DIR}/${BACKUP_FILE}"
  RESTORE_FILE="${TMP_DIR}/${BACKUP_FILE}"
fi

 if [ ! -f "${RESTORE_FILE}" ]; then
  echo "=== Backup file \"${RESTORE_FILE}\" not found, exiting..."
  exit 1
 fi

echo "=== Restoring settings backup from \"${BACKUP_FILE}\"..."
/opt/docker/bin/settings-backup -Dlogback.configurationFile=/opt/docker/etc_log/logback.xml -restore "${RESTORE_FILE}"
echo "==="
