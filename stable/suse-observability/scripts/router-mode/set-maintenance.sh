#!/bin/bash

set -ex

# We exit successfully because we do not want to fail the upgrade on maintenance mode.
trap "echo 'Timeout reached. Exiting successfully'; exit 0" SIGTERM

echo "Applying update to configmap to set router mode"

kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "stackstate.router.mode.automatic.configmap.fullname" . }}
  namespace: {{ .Release.Namespace }}
{{ template "stackstate.router.configmap.data" (merge (dict "RouterState" "maintenance") .) }}
EOF

# We proactively restart the deployment when it is updated. This is strictly speaking not necessary because the content changes will be
# picked up automatically, however:
# - Updating the FS based on a config map can 30 seconds, so a long time
# - We would like to actively reset connections, to make all clients go in to maintenance
# - After the restart we are sure the configmap is applied.
# Release labels identify the router before and after a Deployment rename.
router_selector="app.kubernetes.io/component=router,app.kubernetes.io/instance={{ .Release.Name }}"
router_deployments=$(kubectl get deployments -n "{{ .Release.Namespace }}" -l "$router_selector" -o name)
if [[ -n "$router_deployments" ]]; then
  echo "Restarting router"
  kubectl rollout restart deployment -n "{{ .Release.Namespace }}" -l "$router_selector"

  echo "Waiting for rollout to complete..."
  kubectl rollout status deployment -n "{{ .Release.Namespace }}" -l "$router_selector" --timeout=120s
else
  echo "Deployment not yet found, continuing."
fi

# We opted to not http-based dynamic checking of the 503. This is because:
# when switching from 'active' to 'automatic' mode, the pre-hook will run before the deployment upgrade, so
# the 'router-automatic' configmap will not have been picked up.
#
# A second reason is we do not want to block the pre-hook based on fixes that may be ni the main chart, otherwise we get stuck
# Leaving here for reference
#
#echo "Waiting for router mode to be applied"
#
#while [[ "$(curl -I {{ template "common.fullname.short" . }}-router:8080/api | head -n 1 | cut -d ' ' -f 2)" != "503" ]]; do
#  echo -n "."
#  sleep 1
#done

echo "Router mode enabled"
