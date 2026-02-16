#!/usr/bin/env bash
set -Eeuo pipefail

export TMP_DIR=/tmp

# Set up AWS credentials for S3Proxy access
export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

RESTORE_FILE="${TMP_DIR}/${BACKUP_FILE}"

# First try to download from local settings bucket
echo "=== Attempting to download Settings backup \"${BACKUP_FILE}\" from local settings bucket \"${S3_BUCKET_SETTINGS}\"..."
if sts-toolbox aws s3 --endpoint "${S3_ENDPOINT}" --region us-east-1 cp "s3://${S3_BUCKET_SETTINGS}/${BACKUP_FILE}" "${RESTORE_FILE}" 2>/dev/null; then
  echo "Downloaded from local settings bucket"
elif [ "$BACKUP_CONFIGURATION_UPLOAD_REMOTE" == "true" ]; then
  # Fall back to remote bucket if local not found and remote backup is enabled
  echo "=== Not found in local bucket, downloading from remote bucket \"${BACKUP_CONFIGURATION_BUCKET_NAME}\"..."
  sts-toolbox aws s3 --endpoint "${S3_ENDPOINT}" --region us-east-1 cp "s3://${BACKUP_CONFIGURATION_BUCKET_NAME}/${BACKUP_CONFIGURATION_S3_PREFIX}${BACKUP_FILE}" "${RESTORE_FILE}"
fi

if [ ! -f "${RESTORE_FILE}" ]; then
  echo "=== Backup file \"${BACKUP_FILE}\" not found in any location, exiting..."
  exit 1
fi

echo "=== Restoring settings backup from \"${BACKUP_FILE}\"..."
/opt/docker/bin/settings-backup -Dlogback.configurationFile=/opt/docker/etc_log/logback.xml -restore "${RESTORE_FILE}"
echo "==="

# Clean up
rm -f "${RESTORE_FILE}"
