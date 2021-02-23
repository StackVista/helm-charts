#!/usr/bin/env bash
set -Eeuo pipefail

BACKUP_FILE=$1

JOB_NAME_TEMPLATE=stackgraph-restore-backup
JOB_NAME=stackgraph-restore-$(date +%Y%m%dt%H%M%S)
JOB_YAML_DIR=$(mktemp -d /tmp/sts-restore-XXXXXX)
JOB_YAML_FILE="${JOB_YAML_DIR}/job-${JOB_NAME}.yaml"

kubectl get configmap stackstate-backup-restore-scripts -o jsonpath="{.data.job-${JOB_NAME_TEMPLATE}\.yaml}"  | sed -e "s/${JOB_NAME_TEMPLATE}/${JOB_NAME}/" -e "s/REPLACE_ME_BACKUP_FILE_REPLACE_ME/${BACKUP_FILE}/" > "${JOB_YAML_FILE}"
kubectl create -f "${JOB_YAML_FILE}"
while ! kubectl logs job/"${JOB_NAME}" --container=restore >/dev/null 2>/dev/null ; do echo "Waiting for job to start..."; sleep 2; done; kubectl logs "job/${JOB_NAME}"  --container=restore --follow=true
kubectl delete job "${JOB_NAME}"
rm -rf "${JOB_YAML_DIR}"
