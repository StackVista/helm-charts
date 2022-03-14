#!/usr/bin/env bash

[[ -n "${TRACE+x}" ]] && set -x

set -e

charts=$(ct list-changed --config test/ct.yaml)

for chartPath in "${charts[@]}"; do
  if [ "${chartPath}" == "stable/stackstate" ]; then
    printf "\nNOT running 'kubeconform' on '%s' because kubeconform cannot validate v1/List objects...\n" "${chartPath}"
  else
    printf "\nRunning 'kubeconform' on '%s'...\n" "${chartPath}"

    if [ -e "${chartPath}/linter_values.yaml" ]; then
      valuesFile=("--values" "${chartPath}/linter_values.yaml")
    else
      valuesFile=()
    fi

    helm template --api-versions networking.k8s.io/v1/Ingress "${valuesFile[@]}" "${chartPath}" | kubeconform --skip ServiceMonitor,PrometheusRule -
  fi
done
