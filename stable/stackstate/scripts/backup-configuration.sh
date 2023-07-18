#!/usr/bin/env bash
set -Eeuo pipefail

export TMP_DIR=/tmp

export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

eval "BACKUP_FILE=\"${BACKUP_CONFIGURATION_SCHEDULED_BACKUP_NAME_TEMPLATE}\""

sts settings describe --namespace "" --url "${STACKSTATE_ROUTER_ENDPOINT}" --file "${TMP_DIR}/${BACKUP_FILE}"

grep '"_version":' "${TMP_DIR}/${BACKUP_FILE}" > /dev/null || (
  echo "Exported file is probably not in STJ format, exiting..."
  exit 1
)

echo "=== Uploading backup \"${BACKUP_FILE}\" to bucket \"${BACKUP_CONFIGURATION_BUCKET_NAME}\"..."
aws --endpoint-url "http://${MINIO_ENDPOINT}" --region minio s3 cp "${TMP_DIR}/${BACKUP_FILE}" "s3://${BACKUP_CONFIGURATION_BUCKET_NAME}/${BACKUP_FILE}"

echo "=== Expiring old backups..."
export BACKUP_BUCKET_NAME=$BACKUP_CONFIGURATION_BUCKET_NAME
export BACKUP_SCHEDULED_BACKUP_NAME_PARSE_REGEXP=$BACKUP_CONFIGURATION_SCHEDULED_BACKUP_NAME_PARSE_REGEXP
export BACKUP_SCHEDULED_BACKUP_DATETIME_PARSE_FORMAT=$BACKUP_CONFIGURATION_SCHEDULED_BACKUP_DATETIME_PARSE_FORMAT
export BACKUP_SCHEDULED_BACKUP_RETENTION_TIME_DELTA=$BACKUP_CONFIGURATION_SCHEDULED_BACKUP_RETENTION_TIME_DELTA
/backup-restore-scripts/expire-s3-backups.py
