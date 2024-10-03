#!/usr/bin/env bash
set -Eeuo pipefail

CRONJOB_NAME="suse-observability-backup-conf"
JOB_NAME="${CRONJOB_NAME}-manual-$(date +%Y%m%dt%H%M%S)"

kubectl create job --from=cronjob/"${CRONJOB_NAME}" "${JOB_NAME}"

while ! kubectl logs "job/${JOB_NAME}" >/dev/null 2>/dev/null ; do echo "Waiting for job to start..."; sleep 2; done; kubectl logs "job/${JOB_NAME}"   --follow=true
