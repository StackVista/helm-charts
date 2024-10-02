#!/usr/bin/env bash
set -Eeuo pipefail

JOB_NAME_TEMPLATE=configuration-list-backups
JOB_NAME=configuration-list-backups-$(date +%Y%m%dt%H%M%S)
JOB_YAML_DIR=$(mktemp -d /tmp/sts-restore-XXXXXX)
JOB_YAML_FILE="${JOB_YAML_DIR}/job-${JOB_NAME}.yaml"

CM_NAME="$(kubectl get configmap -l stackstate.com/backup-scripts=true -o jsonpath='{.items[0].metadata.name}')"
if [ -z "${CM_NAME}" ]; then
    echo "=== Configmap not found. Exiting..."
    exit 1
fi

if (! (kubectl get configmap "${CM_NAME}" -o jsonpath="{.data.job-${JOB_NAME_TEMPLATE}\.yaml}" | sed -e "s/${JOB_NAME_TEMPLATE}/${JOB_NAME}/" > "${JOB_YAML_FILE}")) || [ ! -s "${JOB_YAML_FILE}" ]; then
    echo "Did you set backup.enabled and backup.configuration.restore.enabled to true?"
    exit 1
fi

# Ensure the job is deleted when the script exits
cleanup() {
  echo "=== Cleaning up job ${JOB_NAME}"
  kubectl delete job "${JOB_NAME}" || true
}
trap cleanup EXIT

kubectl create -f "${JOB_YAML_FILE}"
rm -rf "${JOB_YAML_DIR}"

while ! kubectl logs "job/${JOB_NAME}"  --container=list >/dev/null 2>/dev/null ; do echo "Waiting for job to start..."; sleep 2; done; kubectl logs "job/${JOB_NAME}"  --container=list --follow=true
