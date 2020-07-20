#!/usr/bin/env bash

[[ -n "${TRACE+x}" ]] && set -x

set -e

for chartPath in $(ct list-changed --config test/ct.yaml); do
  if [ "${chartPath}" == "stable/stackstate" ]; then
    printf "\nNOT running 'kubeval' on '%s' because kubeval cannot validate v1/List objects...\n" "${chartPath}"
  else
    printf "\nRunning 'kubeval' on '%s'...\n" "${chartPath}"

    if [ -e "${chartPath}/linter_values.yaml" ]; then
      valuesFile="--values ${chartPath}/linter_values.yaml"
    else
      valuesFile=""
    fi

    helm template "${valuesFile}" "${chartPath}" | kubeval --skip-kinds ServiceMonitor -
  fi
done
