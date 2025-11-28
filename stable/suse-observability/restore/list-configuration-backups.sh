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
    echo "Did you set backup.enabled to true?"
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

while true; do
    # Get the pod name for the job (pending pods will exit kubectl log without an error, but also without showing the logs)
    POD_NAME=$(kubectl get pods --selector=job-name="${JOB_NAME}" --field-selector=status.phase!=Pending -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)

    if [ -z "${POD_NAME}" ]; then
        echo "=== Waiting for pod to start..."
        sleep 1
        continue
    fi
    break
done

kubectl logs "job/${JOB_NAME}"  --container=list --follow=true
