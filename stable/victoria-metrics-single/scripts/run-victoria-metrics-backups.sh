#!/usr/bin/env sh
set -euo

METRIC_NAME=_sts_generation
TMP_READ_GENERATION=/tmp/vm-generation.txt
wget --post-data="match[]=$METRIC_NAME" -O /tmp/vm-generation.txt http://127.0.0.1:8428/api/v1/export
READ_GENERATION_CODE=$?
if [ $READ_GENERATION_CODE -ne 0 ]; then
  echo "Error fetching $METRIC_NAME with exit code $READ_GENERATION_CODE"
  exit 1
fi

if [ -s $TMP_READ_GENERATION ]; then
  VM_GENERATION=$(cat $TMP_READ_GENERATION | jq -e '.values | last')
else
  # Generate timestamp as the vm_generation
  VM_GENERATION=$(date +"%Y%m%d%H%M%S")
  wget --post-data "$METRIC_NAME $VM_GENERATION" http://127.0.0.1:8428/api/v1/import/prometheus -O /dev/null
  WRITE_GENERATION_CODE=$?
  if [ $WRITE_GENERATION_CODE -ne 0 ]; then
    echo "Error persisting $METRIC_NAME with value $VM_GENERATION and exit code $WRITE_GENERATION_CODE"
    exit 1
  fi
fi

rm -rf $TMP_READ_GENERATION

/vmbackup-prod -storageDataPath=/storage/ -snapshot.createURL=http://localhost:8428/snapshot/create -dst=s3://"$BUCKET_NAME"/"$S3_PREFIX"-"$VM_GENERATION" -customS3Endpoint="$S3_ENDPOINT"
