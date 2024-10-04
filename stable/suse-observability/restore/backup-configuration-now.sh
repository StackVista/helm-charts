#!/usr/bin/env bash
set -Eeuo pipefail

CRONJOB_NAME=$(kubectl get cronjob -l app.kubernetes.io/component=backup-settings -l app.kubernetes.io/name=suse-observability -o name)
JOB_NAME="suse-observability-manual-$(date +%Y%m%dt%H%M%S)"

kubectl create job --from="${CRONJOB_NAME}" "${JOB_NAME}"

while ! kubectl logs "job/${JOB_NAME}" >/dev/null 2>/dev/null ; do echo "Waiting for job to start..."; sleep 2; done; kubectl logs "job/${JOB_NAME}"   --follow=true
