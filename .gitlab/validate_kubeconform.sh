#!/usr/bin/env bash

[[ -n "${TRACE+x}" ]] && set -x

set -e

charts=$(ct list-changed --config test/ct.yaml)

for chartPath in "${charts[@]}"; do
  if [ "${chartPath}" == "stable/suse-observability-values" ]; then
    printf "\nNOT running 'kubeconform' on '%s' because this Helm chart does not output Kubernetes objects, but a values file.\n" "${chartPath}"
  else
    printf "\nRunning 'kubeconform' on '%s'...\n" "${chartPath}"

    if [ -e "${chartPath}/linter_values.yaml" ]; then
      valuesFile=("--values" "${chartPath}/linter_values.yaml")
    else
      valuesFile=()
    fi

    helm template --api-versions networking.k8s.io/v1/Ingress --api-versions policy/v1/PodDisruptionBudget --api-versions "batch/v1/CronJob" "${valuesFile[@]}" "${chartPath}" | kubeconform --skip ServiceMonitor,PrometheusRule,Issuer,Certificate -
  fi
done
