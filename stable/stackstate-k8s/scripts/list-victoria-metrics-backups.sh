#!/usr/bin/env bash
set -Eeuo pipefail

export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

# Fetches timestamp of latest successful backup of given VM instance
function fetchBackupDate() {
  BUCKET=$1
  INSTANCE_S3_PREFIX=$2
  INSTANCE_S3_ROOTDIR=$(dirname "${INSTANCE_S3_PREFIX}")

  # If INSTANCE_S3_PREFIX doesn't contain any slashes, dirname returns ".", which is not a valid S3 prefix
  if [ "${INSTANCE_S3_ROOTDIR}" == "." ]; then
    INSTANCE_S3_ROOTDIR=""
  else
    INSTANCE_S3_ROOTDIR="${INSTANCE_S3_ROOTDIR}/"
  fi

  INSTANCE_NAME=$(basename "${INSTANCE_S3_PREFIX}")
  set +e
  ALL_S3_prefixes=$(sts-toolbox aws s3 ls --endpoint "http://${MINIO_ENDPOINT}" --region minio --bucket "${BACKUP_VICTORIA_METRICS_0_BUCKET_NAME}" --prefix "${INSTANCE_S3_PREFIX}" )
  for S3_PREFIX in $ALL_S3_prefixes
  do
    if [ "$S3_PREFIX" != "None" ]; then
      BACKUP_DATE=$(sts-toolbox aws s3 head --endpoint "http://${MINIO_ENDPOINT}" --region minio --bucket "${BUCKET}" --key "${INSTANCE_S3_ROOTDIR}${S3_PREFIX}backup_complete.ignore" | jq -r .lastModified)
      BACKUP_EXISTS=$?
      set -e
      if [ ${BACKUP_EXISTS} -eq 0 ]; then
        echo "${INSTANCE_NAME} ${INSTANCE_S3_ROOTDIR}${S3_PREFIX::-1} ${BACKUP_DATE}"
      elif [ ${BACKUP_EXISTS} -eq 254 ]; then
        echo "Backup not found for '${S3_PREFIX::-1}'"
      else
        echo "Unexpected status code ${BACKUP_EXISTS} when querying ${S3_PREFIX::-1}"
      fi
    fi
  done
  set -e
}

echo "=== Fetching timestamps of last completed incremental backups"
fetchBackupDate "${BACKUP_VICTORIA_METRICS_0_BUCKET_NAME}" "${BACKUP_VICTORIA_METRICS_0_S3_PREFIX}"
fetchBackupDate "${BACKUP_VICTORIA_METRICS_1_BUCKET_NAME}" "${BACKUP_VICTORIA_METRICS_1_S3_PREFIX}"
echo "==="
