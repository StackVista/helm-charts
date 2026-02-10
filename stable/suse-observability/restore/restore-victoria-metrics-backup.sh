#!/usr/bin/env bash
set -Eeuo pipefail

if [ "$#" -le 1 ]; then
  echo "Required arguments with %instance_name% %s3_prefix%, like victoria-metrics-0 victoria-metrics-0-1701186840"
  exit 1
fi

INSTANCE_NAME=$1
S3_PREFIX=$2

JOB_NAME_TEMPLATE=victoria-metrics-restore-backup
JOB_NAME=victoria-metrics-restore-backup-$(date +%Y%m%dt%H%M%S)
JOB_YAML_DIR=$(mktemp -d /tmp/sts-restore-XXXXXX)
JOB_YAML_FILE="${JOB_YAML_DIR}/job-${JOB_NAME}.yaml"

CM_NAME="$(kubectl get configmap -l stackstate.com/backup-scripts=true -o jsonpath='{.items[0].metadata.name}')"
if [ -z "${CM_NAME}" ]; then
    echo "=== Configmap not found. Exiting..."
    exit 1
fi

# Checks if restore is enabled for this specific instance
BACKUP_ENABLED=$(kubectl get configmap "${CM_NAME}" -o jsonpath="{.data.restore-${INSTANCE_NAME}-enabled}")
if [ "$BACKUP_ENABLED" != "true" ]; then
  echo "Did you set backup.enabled to true?"
  exit 1
fi

if (! (kubectl get configmap "${CM_NAME}" -o jsonpath="{.data.job-${JOB_NAME_TEMPLATE}\.yaml}"  | sed -e "s#${JOB_NAME_TEMPLATE}#${JOB_NAME}#" -e "s#REPLACE_ME_VICTORIA_METRICS_INSTANCE_NAME#${INSTANCE_NAME}#" -e "s#REPLACE_ME_VICTORIA_METRICS_S3_PREFIX#${S3_PREFIX}#" > "${JOB_YAML_FILE}")) || [ ! -s "${JOB_YAML_FILE}" ]; then
    echo "Did you set backup.enabled to true?"
    exit 1
fi

echo "=== Scaling down the Victoria Metrics instance"
kubectl scale statefulsets "suse-observability-$INSTANCE_NAME" --replicas=0

echo "=== Allowing pods to terminate"
sleep 15

echo "=== Starting restore job"
kubectl create -f "${JOB_YAML_FILE}"
rm -rf "${JOB_YAML_DIR}"

echo "=== Restore job started. Follow the logs with the following command:"
echo "kubectl logs --follow job/${JOB_NAME}"

echo "=== After the job has completed clean up the job and scale up the Victoria Metrics instance pods again with the following commands:"
echo "kubectl delete job ${JOB_NAME}"
echo "kubectl scale statefulsets suse-observability-$INSTANCE_NAME --replicas=1"
