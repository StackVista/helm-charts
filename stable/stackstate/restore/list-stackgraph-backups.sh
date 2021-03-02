#!/usr/bin/env bash
set -Eeuo pipefail

JOB_NAME_TEMPLATE=stackgraph-list-backups
JOB_NAME=stackgraph-list-backups-$(date +%Y%m%dt%H%M%S)
JOB_YAML_DIR=$(mktemp -d /tmp/sts-restore-XXXXXX)
JOB_YAML_FILE="${JOB_YAML_DIR}/job-${JOB_NAME}.yaml"

if (! (kubectl get configmap stackstate-backup-restore-scripts -o jsonpath="{.data.job-${JOB_NAME_TEMPLATE}\.yaml}" | sed -e "s/${JOB_NAME_TEMPLATE}/${JOB_NAME}/" > "${JOB_YAML_FILE}")) || [ ! -s "${JOB_YAML_FILE}" ]; then
    echo "Did you set backup.stackGraph.restore.enabled to true?"
    exit 1
fi

kubectl create -f "${JOB_YAML_FILE}"
while ! kubectl logs "job/${JOB_NAME}"  --container=list >/dev/null 2>/dev/null ; do echo "Waiting for job to start..."; sleep 2; done; kubectl logs "job/${JOB_NAME}"  --container=list --follow=true
kubectl delete job "${JOB_NAME}"
rm -rf "${JOB_YAML_DIR}"
