#!/usr/bin/env bash
set -Eeuo pipefail

function uploadFileToS3() {
    srcFile=$1
    destObject=$2
    s3_endpoint=$3
    echo "=== Uploading backup \"${srcFile}\" to bucket \"${destObject}\"..."
    sts-toolbox aws s3 cp --endpoint "${s3_endpoint}" --region us-east-1 "${srcFile}" "${destObject}"
}

export TMP_DIR=/tmp-data

mkdir -p "${TMP_DIR}"

export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

export BACKUP_V2_LOCATION="${BACKUP_STACKGRAPH_BUCKET_NAME}/${BACKUP_STACKGRAPH_S3_PREFIX}v2/"

# StackPacks backup (they work best when created right after the settings backup such that available stackpacks are in sync with the settings)
echo "=== Creating StackPacks backup..."
eval "BACKUP_FILE=\"${BACKUP_STACKGRAPH_SCHEDULED_BACKUP_NAME_TEMPLATE}.v2\""
STACKPACKS_BACKUP_FILE="${BACKUP_FILE}.stackpacks.zip"
echo "=== Exporting StackPacks data to \"${STACKPACKS_BACKUP_FILE}\"..."
/opt/docker/bin/stack-packs-backup -Dlogback.configurationFile=/opt/docker/etc_log/logback.xml -create "${TMP_DIR}/${STACKPACKS_BACKUP_FILE}" -remote "${BACKUP_STACKPACKS_SERVICE_URL}"

if [ ! -f "${TMP_DIR}/${STACKPACKS_BACKUP_FILE}" ]; then
    echo "=== StackPacks export failed. Backup file \"${TMP_DIR}/${STACKPACKS_BACKUP_FILE}\" does not exist."
fi

# shellcheck disable=SC2153
uploadFileToS3 "${TMP_DIR}/${STACKPACKS_BACKUP_FILE}" "s3://${BACKUP_V2_LOCATION}${BACKUP_STACKPACKS_DIR}${STACKPACKS_BACKUP_FILE}" "http://${S3_ENDPOINT}"

echo "=== Expiring old StackPacks backups..."
export BACKUP_BUCKET_NAME=${BACKUP_STACKGRAPH_BUCKET_NAME}
export S3_PREFIX="${BACKUP_STACKGRAPH_S3_PREFIX}v2/${BACKUP_STACKPACKS_DIR}"
export BACKUP_SCHEDULED_BACKUP_NAME_PARSE_REGEXP=${BACKUP_STACKGRAPH_SCHEDULED_BACKUP_NAME_PARSE_REGEXP}
export BACKUP_SCHEDULED_BACKUP_DATETIME_PARSE_FORMAT=${BACKUP_STACKGRAPH_SCHEDULED_BACKUP_DATETIME_PARSE_FORMAT}
export BACKUP_SCHEDULED_BACKUP_RETENTION_TIME_DELTA=${BACKUP_STACKGRAPH_SCHEDULED_BACKUP_RETENTION_TIME_DELTA}
/backup-restore-scripts/expire-s3-backups.sh


echo "=== Exporting StackGraph data to $BACKUP_FILE..."

# We need to set all config per-bucket we access. This is due to the way hadoop-aws interprets its config, it expects the
# config per-bucket. Because we allow configuring the bucket name, all that config goes here.

TYPESAFE_ESCAPED_BUCKET=$(echo "${BACKUP_STACKGRAPH_BUCKET_NAME}" | sed 's/_/___/g; s/-/__/g; s/\./_/g')
# HACK: We configure hbase here through typesafe config. However, typesafe does not support a key being both object and
# string (as in `endpoint = "string"` and `endpoint.region = "us-east-1"`
# We build a little hack there, which allows postfixing endpoint as endpoint__, which gets stripped when transforming typesafe to hbase conf
AWS_BUCKET_ENDPOINT_VAR="CONFIG_FORCE_fs_s3a_bucket_${TYPESAFE_ESCAPED_BUCKET}_endpoint___"
export "${AWS_BUCKET_ENDPOINT_VAR}=http://${S3_ENDPOINT}"
AWS_BUCKET_REGION_VAR="CONFIG_FORCE_fs_s3a_bucket_${TYPESAFE_ESCAPED_BUCKET}_endpoint_region"
export "${AWS_BUCKET_REGION_VAR}=us-east-1"
AWS_BUCKET_ACCESS_KEY_VAR="CONFIG_FORCE_fs_s3a_bucket_${TYPESAFE_ESCAPED_BUCKET}_access_key"
export "${AWS_BUCKET_ACCESS_KEY_VAR}=$(cat /aws-keys/accesskey)"
AWS_BUCKET_SECRET_KEY_VAR="CONFIG_FORCE_fs_s3a_bucket_${TYPESAFE_ESCAPED_BUCKET}_secret_key"
export "${AWS_BUCKET_SECRET_KEY_VAR}=$(cat /aws-keys/secretkey)"

AWS_BUCKET_PATH_STYLE_VAR="CONFIG_FORCE_fs_s3a_bucket_${TYPESAFE_ESCAPED_BUCKET}_path_style_access"
export "${AWS_BUCKET_PATH_STYLE_VAR}=true"
AWS_BUCKET_CONNECTION_SSL_ENABLED_VAR="CONFIG_FORCE_fs_s3a_bucket_${TYPESAFE_ESCAPED_BUCKET}_connection_ssl_enabled"
export "${AWS_BUCKET_CONNECTION_SSL_ENABLED_VAR}=false"
AWS_BUCKET_AWS_CREDENTIALS_PROVIDER_VAR="CONFIG_FORCE_fs_s3a_bucket_${TYPESAFE_ESCAPED_BUCKET}_aws_credentials_provider"
export "${AWS_BUCKET_AWS_CREDENTIALS_PROVIDER_VAR}=org.apache.hadoop.fs.s3a.SimpleAWSCredentialsProvider"
AWS_BUCKET_AWS_CREDENTIALS_PROVIDER_VAR="CONFIG_FORCE_fs_s3a_bucket_${TYPESAFE_ESCAPED_BUCKET}_buffer_dir"
export "${AWS_BUCKET_AWS_CREDENTIALS_PROVIDER_VAR}=/tmp-data"

# Turn the retention date into multiple days
DAYS=$(( ($(date +%s) - $(date -d "${BACKUP_SCHEDULED_BACKUP_RETENTION_TIME_DELTA}" +%s) + 86399) / 86400 ))

echo "Keeping backups for '$DAYS' days."
export CONFIG_FORCE_stackgraph_backup_snapshotRetentionDays="$DAYS"

/opt/docker/bin/stackstate-server -Dlogback.configurationFile=/opt/docker/etc_log/logback.xml -export-v2-incremental "s3a://${BACKUP_V2_LOCATION}" -backup-name "$BACKUP_FILE"
