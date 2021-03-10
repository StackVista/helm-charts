#!/usr/bin/env bash
set -Eeuo pipefail

JOB_NAME_TEMPLATE=elasticsearch-list-snapshots
JOB_NAME=elasticsearch-list-snapshots-$(date +%Y%m%dt%H%M%S)
JOB_YAML_DIR=$(mktemp -d /tmp/sts-restore-XXXXXX)
JOB_YAML_FILE="${JOB_YAML_DIR}/job-${JOB_NAME}.yaml"

if (! (kubectl get configmap stackstate-backup-restore-scripts -o jsonpath="{.data.job-${JOB_NAME_TEMPLATE}\.yaml}" | sed -e "s/${JOB_NAME_TEMPLATE}/${JOB_NAME}/" > "${JOB_YAML_FILE}")) || [ ! -s "${JOB_YAML_FILE}" ]; then
    echo "Did you set backup.enabled and backup.elasticsearch.restore.enabled to true?"
    exit 1
fi

kubectl create -f "${JOB_YAML_FILE}"
rm -rf "${JOB_YAML_DIR}"

while ! kubectl logs "job/${JOB_NAME}" --container=list >/dev/null 2>/dev/null ; do echo "Waiting for job to start..."; sleep 2; done; kubectl logs "job/${JOB_NAME}" --container=list --follow=true
kubectl delete job "${JOB_NAME}"
