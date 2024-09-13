#!/usr/bin/env bash

set -euo pipefail

OLDER_THAN=$(date -d "${BACKUP_SCHEDULED_BACKUP_RETENTION_TIME_DELTA}" +"%Y%m%d-%H%M")
if [ -z "${OLDER_THAN}" ]; then
  echo "Failed to calculate the backup expiration date. Exiting."
  exit 1
fi

sts-toolbox aws s3 ls --endpoint "http://${MINIO_ENDPOINT}" --region minio --bucket "${BACKUP_BUCKET_NAME}" --prefix "${S3_PREFIX}" | while read -r backup_file
do
  if [[ "${backup_file}" =~ ${BACKUP_SCHEDULED_BACKUP_NAME_PARSE_REGEXP} ]];
  then
    BACKUP_DATE="${BASH_REMATCH[1]}"
    if [[ "${BACKUP_DATE}" < "${OLDER_THAN}" ]];
    then
      sts-toolbox aws s3 delete --endpoint "http://${MINIO_ENDPOINT}" --region minio --bucket "${BACKUP_BUCKET_NAME}" --key "${S3_PREFIX}${backup_file}"
    else
        echo "${S3_PREFIX}${backup_file} is not older than ${OLDER_THAN}. Skipping deletion."
    fi
  fi
done
