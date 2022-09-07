#!/usr/bin/env bash
set -Eeuo pipefail

TMP_DIR=/tmp

export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

echo "=== Downloading Configuration backup \"${BACKUP_FILE}\" from bucket \"${BACKUP_CONFIGURATION_BUCKET_NAME}\"..."
aws --endpoint-url "http://${MINIO_ENDPOINT}" s3 cp "s3://${BACKUP_CONFIGURATION_BUCKET_NAME}/${BACKUP_FILE}" "${TMP_DIR}/${BACKUP_FILE}"

echo "=== Importing Configuration data from \"${BACKUP_FILE}\"..."
sts settings apply --namespace "" --url "${STACKSTATE_ROUTER_ENDPOINT}" --file "${TMP_DIR}/${BACKUP_FILE}"
echo "==="
