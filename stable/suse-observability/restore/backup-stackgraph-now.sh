#!/usr/bin/env bash
set -Eeuo pipefail

CRONJOB_NAME=$(kubectl get cronjob -l 'app.kubernetes.io/component=backup-sg,app.kubernetes.io/name=suse-observability' -o name)
JOB_NAME="suse-observability-backup-sg-manual-$(date +%Y%m%dt%H%M%S)"

kubectl create job --from="${CRONJOB_NAME}" "${JOB_NAME}"

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

kubectl logs "job/${JOB_NAME}"   --follow=true
