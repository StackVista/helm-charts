#!/bin/bash

set -ex

echo "Applying update to configmap to set the router mode to active"

kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "stackstate.router.mode.automatic.configmap.fullname" . }}
  namespace: {{ .Release.Namespace }}
{{ template "stackstate.router.configmap.data" (merge (dict "RouterState" "active") .) }}
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
  # Poll with the existing get/list permissions; the hook Role does not grant watch.
  while ! kubectl rollout status deployment -n "{{ .Release.Namespace }}" -l "$router_selector" --watch=false; do
    echo "."
    router_deployments=$(kubectl get deployments -n "{{ .Release.Namespace }}" -l "$router_selector" -o name)
    if [[ -z "$router_deployments" ]]; then
      echo "Deployment went away, exiting"
      exit 0
    fi
    sleep 1
  done
else
  echo "Deployment not yet found, continuing."
fi

echo "Router mode set to active"
