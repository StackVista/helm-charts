#!/usr/bin/env bash
set -Eeuo pipefail

BACKUP_FILE=${1:?Backup file name is required}

JOB_NAME_TEMPLATE=configuration-upload-backup
JOB_NAME=configuration-upload-$(date +%Y%m%dt%H%M%S)
JOB_YAML_DIR=$(mktemp -d /tmp/sts-upload-XXXXXX)
JOB_YAML_FILE="${JOB_YAML_DIR}/job-${JOB_NAME}.yaml"

BACKUP_DIR=/settings-backup-data
BACKUP_FILE_NAME=$(basename "${BACKUP_FILE}")
UPLOADED_FILE="${BACKUP_DIR}/${BACKUP_FILE_NAME}"

CM_NAME="$(kubectl get configmap -l stackstate.com/backup-scripts=true -o jsonpath='{.items[0].metadata.name}')"
if [ -z "${CM_NAME}" ]; then
    echo "=== Configmap not found. Exiting..."
    exit 1
fi

if (! (kubectl get configmap "${CM_NAME}" -o jsonpath="{.data.job-${JOB_NAME_TEMPLATE}\.yaml}"  | sed -e "s/${JOB_NAME_TEMPLATE}/${JOB_NAME}/" > "${JOB_YAML_FILE}")) || [ ! -s "${JOB_YAML_FILE}" ]; then
    echo "Missing job templates."
    exit 1
fi

# Ensure the job is deleted when the script exits
cleanup() {
  echo "=== Cleaning up job ${JOB_NAME}"
  kubectl delete job "${JOB_NAME}" || true
}
trap cleanup EXIT

echo "=== Starting upload job..."
kubectl create -f "${JOB_YAML_FILE}"
rm -rf "${JOB_YAML_DIR}"

while true; do
    # Get the pod name for the job
    POD_NAME=$(kubectl get pods --selector=job-name="${JOB_NAME}" --field-selector=status.phase==Running -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)

    if [ -z "${POD_NAME}" ]; then
        echo "=== Waiting for pod to start..."
        sleep 1
        continue
    fi

    LOGS=$(kubectl logs "pod/${POD_NAME}" --container=upload)

    if echo "$LOGS" | grep -q "=== Waiting for backup file to be uploaded..."; then
        echo "Pod is ready for upload..."
        break
    fi

    sleep 1
done

# Copy the backup file from the the local machine to the pod
kubectl cp "${BACKUP_FILE}" "${POD_NAME}:${UPLOADED_FILE}" --container=upload >/dev/null 2>/dev/null

echo "Backup file $BACKUP_FILE uploaded"
