#!/usr/bin/env bash
set -Eeuo pipefail

SNAPSHOT_NAME=$1

JOB_NAME_TEMPLATE=elasticsearch-restore-snapshot
JOB_NAME=elasticsearch-restore-$(date +%Y%m%dt%H%M%S)
JOB_YAML_DIR=$(mktemp -d /tmp/sts-restore-XXXXXX)
JOB_YAML_FILE="${JOB_YAML_DIR}/job-${JOB_NAME}.yaml"

if (! (kubectl get configmap stackstate-backup-restore-scripts -o jsonpath="{.data.job-${JOB_NAME_TEMPLATE}\.yaml}"  | sed -e "s/${JOB_NAME_TEMPLATE}/${JOB_NAME}/" -e "s/REPLACE_ME_SNAPSHOT_NAME_REPLACE_ME/${SNAPSHOT_NAME}/" > "${JOB_YAML_FILE}")) || [ ! -s "${JOB_YAML_FILE}" ]; then
    echo "Did you set backup.enabled and backup.elasticsearch.restore.enabled to true?"
    exit 1
fi

echo "=== Starting restore job"
kubectl create -f "${JOB_YAML_FILE}"
rm -rf "${JOB_YAML_DIR}"

echo "=== Restore job started. Follow the logs and clean up the job with the following commands:"
echo "kubectl logs --follow job/${JOB_NAME}"
echo "kubectl delete job/${JOB_NAME}"
