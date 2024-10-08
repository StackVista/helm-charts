#!/usr/bin/env bash
set -Eeuo pipefail

BACKUP_FILE=${1:?Backup file name is required}

JOB_NAME_TEMPLATE=configuration-download-backup
JOB_NAME=configuration-download-$(date +%Y%m%dt%H%M%S)
JOB_YAML_DIR=$(mktemp -d /tmp/sts-download-XXXXXX)
JOB_YAML_FILE="${JOB_YAML_DIR}/job-${JOB_NAME}.yaml"

TMP_DIR=/tmp
RESTORE_FILE="${TMP_DIR}/${BACKUP_FILE}"

CM_NAME="$(kubectl get configmap -l stackstate.com/backup-scripts=true -o jsonpath='{.items[0].metadata.name}')"
if [ -z "${CM_NAME}" ]; then
    echo "=== Configmap not found. Exiting..."
    exit 1
fi

if (! (kubectl get configmap "${CM_NAME}" -o jsonpath="{.data.job-${JOB_NAME_TEMPLATE}\.yaml}"  | sed -e "s/${JOB_NAME_TEMPLATE}/${JOB_NAME}/" -e "s/REPLACE_ME_BACKUP_FILE_REPLACE_ME/${BACKUP_FILE}/" > "${JOB_YAML_FILE}")) || [ ! -s "${JOB_YAML_FILE}" ]; then
    echo "Missing job templates."
    exit 1
fi

# Ensure the job is deleted when the script exits
cleanup() {
  echo "=== Cleaning up job ${JOB_NAME}"
  kubectl delete job "${JOB_NAME}" || true
}
trap cleanup EXIT

echo "=== Starting download job..."
kubectl create -f "${JOB_YAML_FILE}"
rm -rf "${JOB_YAML_DIR}"

while true; do
    # Get the pod name for the job
    POD_NAME=$(kubectl get pods --selector=job-name="${JOB_NAME}" --field-selector=status.phase!=Pending -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)

    if [ -z "${POD_NAME}" ]; then
        echo "=== Waiting for pod to start..."
        sleep 1
        continue
    fi

    LOGS=$(kubectl logs "pod/${POD_NAME}" --container=download)

    if echo "$LOGS" | grep -q "=== Waiting for backup file to be downloaded..."; then
        echo "Backup file ${RESTORE_FILE} found, downloading..."
        break
    elif echo "$LOGS" | grep -q "=== Backup file not found"; then
        echo "=== Backup file not found ${RESTORE_FILE}, exiting."
        exit 1
    fi

    sleep 1
done

# Copy the backup file from the pod to the local machine
kubectl cp "${POD_NAME}:${RESTORE_FILE}" "${BACKUP_FILE}" --container=download >/dev/null 2>/dev/null

echo "Backup file $BACKUP_FILE downloaded"
