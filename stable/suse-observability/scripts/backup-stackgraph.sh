#!/usr/bin/env bash
set -Eeuo pipefail

function uploadFileToS3() {
    srcFile=$1
    destObject=$2
    s3_endpoint=$3
    echo "=== Uploading StackGraph backup \"${srcFile}\" to bucket \"${destObject}\"..."
    sts-toolbox aws s3 cp --endpoint "${s3_endpoint}" --region us-east-1 "${srcFile}" "${destObject}"
}

export TMP_DIR=/tmp-data

export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

echo "=== Removing old StackGraph backups from temporary storage \"${TMP_DIR}\"..."
(cd "${TMP_DIR:?}" && rm -rf ./sts-backup-*)
echo "=== Listing contents of \"${TMP_DIR}\"..."
ls -lat "${TMP_DIR:?}"
eval "BACKUP_FILE=\"${BACKUP_STACKGRAPH_SCHEDULED_BACKUP_NAME_TEMPLATE}\""
echo "=== Exporting StackGraph data to \"${BACKUP_FILE}\"..."
/opt/docker/bin/stackstate-server -Dlogback.configurationFile=/opt/docker/etc_log/logback.xml -export "${TMP_DIR}/${BACKUP_FILE}"


if [ ! -f "${TMP_DIR}/${BACKUP_FILE}" ]; then
    echo "=== Export failed. Backup file \"${TMP_DIR}/${BACKUP_FILE}\" does not exist."
    exit 1
fi

# StackPacks backup (they work best when created right after the settings backup such that available stackpacks are in sync with the settings)
echo "=== Creating StackPacks backup..."
STACKPACKS_BACKUP_FILE="${BACKUP_FILE}.stackpacks.zip"
echo "=== Exporting StackPacks data to \"${STACKPACKS_BACKUP_FILE}\"..."
/opt/docker/bin/stack-packs-backup -Dlogback.configurationFile=/opt/docker/etc_log/logback.xml -create "${TMP_DIR}/${STACKPACKS_BACKUP_FILE}" -remote "${BACKUP_STACKPACKS_SERVICE_URL}"

if [ ! -f "${TMP_DIR}/${STACKPACKS_BACKUP_FILE}" ]; then
    echo "=== StackPacks export failed. Backup file \"${TMP_DIR}/${STACKPACKS_BACKUP_FILE}\" does not exist."
fi

if [ "${BACKUP_STACKGRAPH_ARCHIVE_SPLIT_SIZE:-0}" == "0" ]; then
    # shellcheck disable=SC2153
    uploadFileToS3 "${TMP_DIR}/${BACKUP_FILE}" "s3://${BACKUP_STACKGRAPH_BUCKET_NAME}/${BACKUP_STACKGRAPH_S3_PREFIX}${BACKUP_FILE}" "http://${S3_ENDPOINT}"
else
    # Split Stackgraph dump, sts-backup-20240223-0915.graph -> sts-backup-20240223-0915.graph.00, sts-backup-20240223-0915.graph.01, sts-backup-20240223-0915.graph.XX
    (cd "${TMP_DIR}" && split --verbose -d -b "${BACKUP_STACKGRAPH_ARCHIVE_SPLIT_SIZE}" "${TMP_DIR}/${BACKUP_FILE}" "$(basename "${BACKUP_FILE}").")
    find ${TMP_DIR} -name "${BACKUP_FILE}.*" | while read -r file
    do
        uploadFileToS3 "${file}" "s3://${BACKUP_STACKGRAPH_BUCKET_NAME}/${BACKUP_STACKGRAPH_S3_PREFIX}$(basename "${file}")" "http://${S3_ENDPOINT}"
    done
fi

echo "=== Expiring old StackGraph backups..."
export BACKUP_BUCKET_NAME=${BACKUP_STACKGRAPH_BUCKET_NAME}
export S3_PREFIX=${BACKUP_STACKGRAPH_S3_PREFIX}
export BACKUP_SCHEDULED_BACKUP_NAME_PARSE_REGEXP=${BACKUP_STACKGRAPH_SCHEDULED_BACKUP_NAME_PARSE_REGEXP}
export BACKUP_SCHEDULED_BACKUP_DATETIME_PARSE_FORMAT=${BACKUP_STACKGRAPH_SCHEDULED_BACKUP_DATETIME_PARSE_FORMAT}
export BACKUP_SCHEDULED_BACKUP_RETENTION_TIME_DELTA=${BACKUP_STACKGRAPH_SCHEDULED_BACKUP_RETENTION_TIME_DELTA}
/backup-restore-scripts/expire-s3-backups.sh

uploadFileToS3 "${TMP_DIR}/${STACKPACKS_BACKUP_FILE}" "s3://${BACKUP_STACKGRAPH_BUCKET_NAME}/${BACKUP_STACKGRAPH_STACKPACKS_S3_PREFIX}${STACKPACKS_BACKUP_FILE}" "http://${MINIO_ENDPOINT}"

echo "=== Expiring old StackPacks backups..."
export BACKUP_BUCKET_NAME=${BACKUP_STACKGRAPH_BUCKET_NAME}
export S3_PREFIX=${BACKUP_STACKGRAPH_STACKPACKS_S3_PREFIX}
export BACKUP_SCHEDULED_BACKUP_NAME_PARSE_REGEXP=${BACKUP_STACKGRAPH_SCHEDULED_BACKUP_NAME_PARSE_REGEXP}
export BACKUP_SCHEDULED_BACKUP_DATETIME_PARSE_FORMAT=${BACKUP_STACKGRAPH_SCHEDULED_BACKUP_DATETIME_PARSE_FORMAT}
export BACKUP_SCHEDULED_BACKUP_RETENTION_TIME_DELTA=${BACKUP_STACKGRAPH_SCHEDULED_BACKUP_RETENTION_TIME_DELTA}
/backup-restore-scripts/expire-s3-backups.sh
