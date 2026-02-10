#!/usr/bin/env bash
set -Eeuo pipefail

show_usage() {
    echo "Usage: $0 SNAPSHOT_NAME [INDICES_TO_RESTORE] [--delete-all-indices]"
    echo "   or: $0 SNAPSHOT_NAME --delete-all-indices"
    echo ""
    echo "Arguments:"
    echo "  SNAPSHOT_NAME           Name of the snapshot to restore (required)"
    echo "  INDICES_TO_RESTORE      Comma-separated list of indices to restore (optional)"
    echo "  --delete-all-indices    Delete all indices before restore (optional)"
    echo ""
    echo "Examples:"
    echo "  $0 snapshot-2024"
    echo "  $0 snapshot-2024 index1,index2"
    echo "  $0 snapshot-2024 --delete-all-indices"
    echo "  $0 snapshot-2024 index1,index2 --delete-all-indices"
    exit 1
}

# Validate arguments
if [ $# -lt 1 ] || [ $# -gt 3 ]; then
    echo "Error: Invalid number of arguments"
    show_usage
fi

SNAPSHOT_NAME=$1
INDICES_TO_RESTORE=""
DELETE_ALL_INDICES="false"

# Parse arguments based on count and values
if [ $# -eq 2 ]; then
    if [ "$2" = "--delete-all-indices" ]; then
        # Case: SNAPSHOT_NAME --delete-all-indices
        INDICES_TO_RESTORE=""
        DELETE_ALL_INDICES="true"
    else
        # Case: SNAPSHOT_NAME INDICES_TO_RESTORE
        INDICES_TO_RESTORE="$2"
        DELETE_ALL_INDICES="false"
    fi
elif [ $# -eq 3 ]; then
    if [ "$3" = "--delete-all-indices" ]; then
        # Case: SNAPSHOT_NAME INDICES_TO_RESTORE --delete-all-indices
        INDICES_TO_RESTORE="$2"
        DELETE_ALL_INDICES="true"
    else
        echo "Error: Third argument must be --delete-all-indices"
        show_usage
    fi
fi

# Warn user if delete-all-indices is enabled
if [ "$DELETE_ALL_INDICES" = "true" ]; then
    echo "WARNING: All indices will be deleted before restore!"
    read -r -p "Are you sure you want to continue? (yes/no): " confirmation
    if [ "$confirmation" != "yes" ]; then
        echo "Restore cancelled."
        exit 0
    fi
fi

JOB_NAME_TEMPLATE=elasticsearch-restore-snapshot
JOB_NAME=elasticsearch-restore-$(date +%Y%m%dt%H%M%S)
JOB_YAML_DIR=$(mktemp -d /tmp/sts-restore-XXXXXX)
JOB_YAML_FILE="${JOB_YAML_DIR}/job-${JOB_NAME}.yaml"

CM_NAME="$(kubectl get configmap -l stackstate.com/backup-scripts=true -o jsonpath='{.items[0].metadata.name}')"
if [ -z "${CM_NAME}" ]; then
    echo "=== Configmap not found. Exiting..."
    exit 1
fi

if (! (kubectl get configmap "${CM_NAME}" -o jsonpath="{.data.job-${JOB_NAME_TEMPLATE}\.yaml}"  | sed -e "s/${JOB_NAME_TEMPLATE}/${JOB_NAME}/" -e "s/REPLACE_ME_SNAPSHOT_NAME_REPLACE_ME/${SNAPSHOT_NAME}/" -e "s/REPLACE_ME_INDICES_TO_RESTORE_REPLACE_ME/${INDICES_TO_RESTORE}/" -e "s/REPLACE_ME_DELETE_ALL_INDICES_REPLACE_ME/${DELETE_ALL_INDICES}/"> "${JOB_YAML_FILE}")) || [ ! -s "${JOB_YAML_FILE}" ]; then
    echo "Did you set backup.enabled to true?"
    exit 1
fi

echo "=== Starting restore job"
kubectl create -f "${JOB_YAML_FILE}"
rm -rf "${JOB_YAML_DIR}"

echo "=== Restore job started. Follow the logs and clean up the job with the following commands:"
echo "kubectl logs --follow job/${JOB_NAME}"
echo "kubectl delete job/${JOB_NAME}"
