#!/usr/bin/env sh

[ -n "${TRACE+x}" ] && set -x

set -e

for chartPath in $(ct list-changed --config test/ct.yaml); do
  printf "\nRunning 'kubeval' on '%s'...\n" "${chartPath}"

  if [ -e "${chartPath}/linter_values.yaml" ]; then
    valuesFile="--values ${chartPath}/linter_values.yaml"
  else
    valuesFile=""
  fi

  helm template ${valuesFile} "${chartPath}" | kubeval --skip-kinds ServiceMonitor -
done
