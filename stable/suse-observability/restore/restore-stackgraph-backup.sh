#!/usr/bin/env bash
set -Eeuo pipefail

BACKUP_FILE=$1

FORCE_DELETE=""
if [ "${2:-}" = "-force" ]; then
  FORCE_DELETE=$2
fi;

JOB_NAME_TEMPLATE=stackgraph-restore-backup
JOB_NAME=stackgraph-restore-$(date +%Y%m%dt%H%M%S)
JOB_YAML_DIR=$(mktemp -d /tmp/sts-restore-XXXXXX)
JOB_YAML_FILE="${JOB_YAML_DIR}/job-${JOB_NAME}.yaml"
PVC_YAML_FILE="${JOB_YAML_DIR}/pvc-${JOB_NAME}.yaml"

CM_NAME="$(kubectl get configmap -l stackstate.com/backup-scripts=true -o jsonpath='{.items[0].metadata.name}')"
if [ -z "${CM_NAME}" ]; then
    echo "=== Configmap not found. Exiting..."
    exit 1
fi

if (! (kubectl get configmap "${CM_NAME}" -o jsonpath="{.data.job-${JOB_NAME_TEMPLATE}\.yaml}"  | sed -e "s/${JOB_NAME_TEMPLATE}/${JOB_NAME}/" -e "s/REPLACE_ME_BACKUP_FILE_REPLACE_ME/${BACKUP_FILE}/" -e "s/REPLACE_ME_FORCE_DELETE_REPLACE_ME/${FORCE_DELETE}/" > "${JOB_YAML_FILE}")) || [ ! -s "${JOB_YAML_FILE}" ]; then
    echo "Did you set backup.enabled to true?"
    exit 1
fi

if (! (kubectl get configmap "${CM_NAME}" -o jsonpath="{.data.pvc-${JOB_NAME_TEMPLATE}\.yaml}")) | sed -e "s/${JOB_NAME_TEMPLATE}/${JOB_NAME}/" > "${PVC_YAML_FILE}" || [ ! -s "${PVC_YAML_FILE}" ]; then
    echo "Did you set backup.enabled to true?"
    exit 1
fi

echo "=== Scaling down deployments for pods that connect to StackGraph"
kubectl scale --replicas=0 deployments --selector=stackstate.com/connects-to-stackgraph=true

echo "=== Allowing pods to terminate"
sleep 15

echo "=== Starting restore job"
kubectl create -f "${PVC_YAML_FILE}" -f "${JOB_YAML_FILE}"
rm -rf "${JOB_YAML_DIR}"

echo "=== Restore job started. Follow the logs with the following command:"
echo "kubectl logs --follow job/${JOB_NAME}"

echo "=== After job has completed clean up the job and scale up the StackState pods again with the following commands:"
echo "kubectl delete job,pvc ${JOB_NAME}"
echo "./restore/scale-up.sh"
