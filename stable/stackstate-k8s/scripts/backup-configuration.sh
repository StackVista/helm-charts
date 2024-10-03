#!/usr/bin/env bash
set -Eeuo pipefail

export BACKUP_DIR=/settings-backup-data
BACKUP_CONFIGURATION_MAX_LOCAL_FILES=${BACKUP_CONFIGURATION_MAX_LOCAL_FILES:-10}

echo "=== Removing oldest backups from local storage \"${BACKUP_DIR}\", keeping at most $BACKUP_CONFIGURATION_MAX_LOCAL_FILES..."
# shellcheck disable=SC2012
current_count=$(ls -1t "${BACKUP_DIR}" | wc -l)
if [ "$current_count" -gt "$BACKUP_CONFIGURATION_MAX_LOCAL_FILES" ]; then
  files_to_remove=$((current_count - BACKUP_CONFIGURATION_MAX_LOCAL_FILES))
  # shellcheck disable=SC2012
  ls -1t "${BACKUP_DIR}" | tail -n "$files_to_remove" | xargs -I {} rm -f "${BACKUP_DIR}/{}"
fi

echo "=== Listing contents of \"${BACKUP_DIR}\"..."
ls -lat "${BACKUP_DIR:?}"

eval "BACKUP_FILE=\"${BACKUP_DIR}/${BACKUP_CONFIGURATION_SCHEDULED_BACKUP_NAME_TEMPLATE}\""

echo "=== Creating settings backup to \"${BACKUP_FILE}\"..."
/opt/docker/bin/settings-backup -Dlogback.configurationFile=/opt/docker/etc_log/logback.groovy -create "${BACKUP_FILE}"

grep '_version:' "${BACKUP_FILE}" > /dev/null || (
  echo "Exported file is probably not in STY format, exiting..."
  exit 1
)

if [ "$BACKUP_CONFIGURATION_UPLOAD_REMOTE" == "true" ]; then
    export AWS_ACCESS_KEY_ID
    AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
    export AWS_SECRET_ACCESS_KEY
    AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

    echo "=== Uploading backup \"${BACKUP_FILE}\" to bucket \"${BACKUP_CONFIGURATION_BUCKET_NAME}\"..."
    sts-toolbox aws s3 cp --endpoint "http://${MINIO_ENDPOINT}" --region minio "${BACKUP_FILE}" "s3://${BACKUP_CONFIGURATION_BUCKET_NAME}/${BACKUP_CONFIGURATION_S3_PREFIX}${BACKUP_FILE}"

    echo "=== Expiring old settings backups..."
    export BACKUP_BUCKET_NAME=${BACKUP_CONFIGURATION_BUCKET_NAME}
    export S3_PREFIX=${BACKUP_CONFIGURATION_S3_PREFIX}
    export BACKUP_SCHEDULED_BACKUP_NAME_PARSE_REGEXP=${BACKUP_CONFIGURATION_SCHEDULED_BACKUP_NAME_PARSE_REGEXP}
    export BACKUP_SCHEDULED_BACKUP_DATETIME_PARSE_FORMAT=${BACKUP_CONFIGURATION_SCHEDULED_BACKUP_DATETIME_PARSE_FORMAT}
    export BACKUP_SCHEDULED_BACKUP_RETENTION_TIME_DELTA=${BACKUP_CONFIGURATION_SCHEDULED_BACKUP_RETENTION_TIME_DELTA}
    /backup-restore-scripts/expire-s3-backups.sh
fi
