#!/usr/bin/env sh
set -eo

if [ "$#" -eq 0 ]; then
  echo "Required argument with instance name, like victoria-metrics-0"
  exit 1
fi
INSTANCE_NAME=$1
INSTANCE_NAME_AS_ENV_NAME=$(echo "$INSTANCE_NAME" | tr '[:lower:]' '[:upper:]' | tr '-' '_') # Env vars (like bucket name) are uppercase and uses underscores

export AWS_ACCESS_KEY_ID
AWS_ACCESS_KEY_ID="$(cat /aws-keys/accesskey)"
export AWS_SECRET_ACCESS_KEY
AWS_SECRET_ACCESS_KEY="$(cat /aws-keys/secretkey)"

BUCKET_NAME_ENV_NAME="BACKUP_${INSTANCE_NAME_AS_ENV_NAME}_BUCKET_NAME"
BUCKET_NAME=$(printenv "$BUCKET_NAME_ENV_NAME")

/vmrestore-prod -storageDataPath=/storage -src="s3://$BUCKET_NAME/$INSTANCE_NAME" -customS3Endpoint="http://$MINIO_ENDPOINT"
