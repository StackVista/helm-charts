#!/usr/bin/env bash
set -Eeuo pipefail

SNAPSHOT_NAME=$1

JOB_NAME_TEMPLATE=elasticsearch-restore-snapshot
JOB_NAME=elasticsearch-restore-$(date +%Y%m%dt%H%M%S)
JOB_YAML_DIR=$(mktemp -d /tmp/sts-restore-XXXXXX)
JOB_YAML_FILE="${JOB_YAML_DIR}/job-${JOB_NAME}.yaml"

if (! (kubectl get configmap stackstate-backup-restore-scripts -o jsonpath="{.data.job-${JOB_NAME_TEMPLATE}\.yaml}"  | sed -e "s/${JOB_NAME_TEMPLATE}/${JOB_NAME}/" -e "s/REPLACE_ME_SNAPSHOT_NAME_REPLACE_ME/${SNAPSHOT_NAME}/" > "${JOB_YAML_FILE}")) || [ ! -s "${JOB_YAML_FILE}" ]; then
    echo "Did you set backup.elasticSearch.restore.enabled to true?"
    exit 1
fi

kubectl create -f "${JOB_YAML_FILE}"
while ! kubectl logs job/"${JOB_NAME}"  --container=restore >/dev/null 2>/dev/null ; do echo "Waiting for job to start..."; sleep 2; done; kubectl logs "job/${JOB_NAME}" --container=restore --follow=true
kubectl delete job "${JOB_NAME}"
rm -rf "${JOB_YAML_DIR}"
