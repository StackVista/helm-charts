#!/bin/bash

set -ex

echo "Applying update to configmap to set the router mode to active"

kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ template "common.fullname.short" . }}-router-automatic
  namespace: {{ .Release.Namespace }}
{{ template "stackstate.router.configmap.data" (merge (dict "RouterState" "active") .) }}
EOF

# We proactively restart the deployment when it is updated. This is strictly speaking not necessary because the content changes will be
# picked up automatically, however:
# - Updating the FS based on a config map can 30 seconds, so a long time
# - We would like to actively reset connections, to make all clients go in to maintenance
# - After the restart we are sure the configmap is applied.
# shellcheck disable=SC2140
if kubectl get deployment "{{ template "common.fullname.short" . }}-router" -n "{{ .Release.Namespace }}"; then
  echo "Restarting router"
  kubectl rollout restart "deployment/{{ template "common.fullname.short" . }}-router" -n "{{ .Release.Namespace }}"

  echo "Waiting for rollout to complete..."
  while ! kubectl rollout status "deployment/{{ template "common.fullname.short" . }}-router" -n "{{ .Release.Namespace }}"; do
    echo "."
    sleep 1
  done
else
  echo "Deployment not yet found, continuing."
fi

echo "Router mode set to active"
