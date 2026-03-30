#!/usr/bin/env bash
set -Eeuo pipefail

export TMP_BACKUP_DIR=/tmp/settings-backup
BACKUP_CONFIGURATION_MAX_LOCAL_FILES=${BACKUP_CONFIGURATION_MAX_LOCAL_FILES:-10}

# Ensure temp directory exists
mkdir -p "${TMP_BACKUP_DIR}"

# Set up AWS credentials for S3Proxy access
export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

echo "=== Removing oldest backups from settings bucket \"${S3_BUCKET_SETTINGS}\", keeping at most $BACKUP_CONFIGURATION_MAX_LOCAL_FILES..."

# List current backups and count them
current_files=$(sts-toolbox aws s3 ls --endpoint "http://${S3_ENDPOINT}" --region us-east-1 --bucket "${S3_BUCKET_SETTINGS}" --prefix "" 2>/dev/null | grep -E '\.sty$' | awk '{print $NF}' | sort || true)
current_count=$(echo "$current_files" | grep -c . || echo 0)

if [ "$current_count" -ge "$BACKUP_CONFIGURATION_MAX_LOCAL_FILES" ]; then
  files_to_remove=$((current_count - BACKUP_CONFIGURATION_MAX_LOCAL_FILES + 1))
  echo "$current_files" | head -n "$files_to_remove" | while read -r file; do
    echo "Removing old backup: $file"
    sts-toolbox aws s3 delete --endpoint "http://${S3_ENDPOINT}" --region us-east-1 --bucket "${S3_BUCKET_SETTINGS}" --key "${file}"
  done
fi

echo "=== Listing contents of settings bucket \"${S3_BUCKET_SETTINGS}\"..."
sts-toolbox aws s3 ls --endpoint "http://${S3_ENDPOINT}" --region us-east-1 --bucket "${S3_BUCKET_SETTINGS}" --prefix "" 2>/dev/null | awk '{print $NF}' || true

eval "BACKUP_FILE=\"${BACKUP_CONFIGURATION_SCHEDULED_BACKUP_NAME_TEMPLATE}\""
BACKUP_FILE_WITH_PATH="${TMP_BACKUP_DIR}/${BACKUP_FILE}"

echo "=== Creating settings backup to \"${BACKUP_FILE}\"..."
/opt/docker/bin/settings-backup -Dlogback.configurationFile=/opt/docker/etc_log/logback.xml -create "${BACKUP_FILE_WITH_PATH}"

grep '_version:' "${BACKUP_FILE_WITH_PATH}" > /dev/null || (
  echo "Exported file is probably not in STY format, exiting..."
  exit 1
)

# Always upload to settings-local-backup bucket (S3Proxy routes to local PVC)
echo "=== Uploading backup \"${BACKUP_FILE}\" to local settings bucket \"${S3_BUCKET_SETTINGS}\"..."
sts-toolbox aws s3 cp --endpoint "http://${S3_ENDPOINT}" --region us-east-1 "${BACKUP_FILE_WITH_PATH}" "s3://${S3_BUCKET_SETTINGS}/${BACKUP_FILE}"

# StackPacks backup (they work best when created right after the settings backup such that available stackpacks are in sync with the settings)
echo "=== Creating StackPacks backup..."
STACKPACKS_BACKUP_FILE="${BACKUP_FILE}.stackpacks.zip"
STACKPACKS_BACKUP_FILE_WITH_PATH="${TMP_BACKUP_DIR}/${STACKPACKS_BACKUP_FILE}"

echo "=== Exporting StackPacks data to \"${STACKPACKS_BACKUP_FILE}\"..."
/opt/docker/bin/stack-packs-backup -Dlogback.configurationFile=/opt/docker/etc_log/logback.xml -create "${STACKPACKS_BACKUP_FILE_WITH_PATH}" -remote "${BACKUP_STACKPACKS_SERVICE_URL}"

# Always upload stackpacks to local settings bucket (S3Proxy routes to local PVC)
echo "=== Uploading StackPacks backup \"${STACKPACKS_BACKUP_FILE}\" to local settings bucket \"${S3_BUCKET_SETTINGS}\"..."
sts-toolbox aws s3 cp --endpoint "http://${S3_ENDPOINT}" --region us-east-1 "${STACKPACKS_BACKUP_FILE_WITH_PATH}" "s3://${S3_BUCKET_SETTINGS}/${BACKUP_CONFIGURATION_STACKPACKS_S3_PREFIX}${STACKPACKS_BACKUP_FILE}"

# If remote backup is enabled, also upload to remote bucket
if [ "$BACKUP_CONFIGURATION_UPLOAD_REMOTE" == "true" ]; then
    echo "=== Uploading backup \"${BACKUP_FILE}\" to remote bucket \"${BACKUP_CONFIGURATION_BUCKET_NAME}\"..."
    sts-toolbox aws s3 cp --endpoint "http://${S3_ENDPOINT}" --region us-east-1 "${BACKUP_FILE_WITH_PATH}" "s3://${BACKUP_CONFIGURATION_BUCKET_NAME}/${BACKUP_CONFIGURATION_S3_PREFIX}${BACKUP_FILE}"

    echo "=== Expiring old settings backups from remote bucket..."
    export BACKUP_BUCKET_NAME=${BACKUP_CONFIGURATION_BUCKET_NAME}
    export S3_PREFIX=${BACKUP_CONFIGURATION_S3_PREFIX}
    export BACKUP_SCHEDULED_BACKUP_NAME_PARSE_REGEXP=${BACKUP_CONFIGURATION_SCHEDULED_BACKUP_NAME_PARSE_REGEXP}
    export BACKUP_SCHEDULED_BACKUP_DATETIME_PARSE_FORMAT=${BACKUP_CONFIGURATION_SCHEDULED_BACKUP_DATETIME_PARSE_FORMAT}
    export BACKUP_SCHEDULED_BACKUP_RETENTION_TIME_DELTA=${BACKUP_CONFIGURATION_SCHEDULED_BACKUP_RETENTION_TIME_DELTA}
    /backup-restore-scripts/expire-s3-backups.sh

    echo "=== Uploading StackPacks backup \"${STACKPACKS_BACKUP_FILE}\" to bucket \"${BACKUP_CONFIGURATION_BUCKET_NAME}\"..."
    sts-toolbox aws s3 cp --endpoint "http://${S3_ENDPOINT}" --region us-east-1 "${STACKPACKS_BACKUP_FILE_WITH_PATH}" "s3://${BACKUP_CONFIGURATION_BUCKET_NAME}/${BACKUP_CONFIGURATION_STACKPACKS_S3_PREFIX}${STACKPACKS_BACKUP_FILE}"

    echo "=== Expiring old StackPacks backups..."
    export BACKUP_BUCKET_NAME=${BACKUP_CONFIGURATION_BUCKET_NAME}
    export S3_PREFIX=${BACKUP_CONFIGURATION_STACKPACKS_S3_PREFIX}
    export BACKUP_SCHEDULED_BACKUP_NAME_PARSE_REGEXP=${BACKUP_CONFIGURATION_SCHEDULED_BACKUP_NAME_PARSE_REGEXP}
    export BACKUP_SCHEDULED_BACKUP_DATETIME_PARSE_FORMAT=${BACKUP_CONFIGURATION_SCHEDULED_BACKUP_DATETIME_PARSE_FORMAT}
    export BACKUP_SCHEDULED_BACKUP_RETENTION_TIME_DELTA=${BACKUP_CONFIGURATION_SCHEDULED_BACKUP_RETENTION_TIME_DELTA}
    /backup-restore-scripts/expire-s3-backups.sh
fi

# Clean up temp files
rm -f "${BACKUP_FILE_WITH_PATH}"
rm -f "${STACKPACKS_BACKUP_FILE_WITH_PATH}"
