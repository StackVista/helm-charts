#!/usr/bin/env bash

[[ -n "${TRACE+x}" ]] && set -x

set -euo pipefail

if [ "$#" -ne 1 ] || [ -z "$1" ]; then
  echo "Usage: $(basename "$0") <chart-path>" >&2
  echo "Example: $(basename "$0") stable/suse-observability" >&2
  exit 2
fi

chartPath="${1%/}"

if [ ! -f "${chartPath}/Chart.yaml" ]; then
  echo "Error: '${chartPath}' is not a chart directory (no Chart.yaml found)" >&2
  exit 2
fi

if [ "${chartPath}" == "stable/suse-observability-values" ]; then
  printf "\nNOT running 'kubeconform' on '%s' because this Helm chart does not output Kubernetes objects, but a values file.\n" "${chartPath}"
  exit 0
fi

printf "\nRunning 'kubeconform' on '%s'...\n" "${chartPath}"

if [ -e "${chartPath}/linter_values.yaml" ]; then
  valuesFile=("--values" "${chartPath}/linter_values.yaml")
else
  valuesFile=()
fi

helm template --api-versions networking.k8s.io/v1/Ingress --api-versions policy/v1/PodDisruptionBudget --api-versions "batch/v1/CronJob" "${valuesFile[@]}" "${chartPath}" | kubeconform --skip ServiceMonitor,PrometheusRule,Issuer,Certificate -
