#!/usr/bin/env bash
set -Eeuo pipefail

CRONJOB_NAME=$(get cronjob -l app.kubernetes.io/component=backup-settings -l app.kubernetes.io/name=stackstate-k8s)
JOB_NAME="${stackstate-backup-conf}-manual-$(date +%Y%m%dt%H%M%S)"

kubectl create job --from="${CRONJOB_NAME}" "${JOB_NAME}"

while ! kubectl logs "job/${JOB_NAME}" >/dev/null 2>/dev/null ; do echo "Waiting for job to start..."; sleep 2; done; kubectl logs "job/${JOB_NAME}"   --follow=true
