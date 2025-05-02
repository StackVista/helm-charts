#!/usr/bin/env bash
set -o pipefail

# shellcheck disable=SC2034
GREEN='\033[0;32m'
RED='\033[0;31m'
NO_COLOR='\033[0m'

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

function usage() {
  cat <<EOF >&2
Get the list of Docker images used by the SUSE Observability Agent chart

Usage:
  $0 -f suse-observability-agent-X.Y.Z.tgz

Arguments:
    -f : TGZ archive with the SUSE Observability Agent chart
    -d : Directory with SUSE Observability Agent chart
    -h : Show this help text

One of -f or -d can be set, the script's default behavior is to use the directory relative to installation.
EOF
}

# Parse options
while getopts ":f:d:h" opt; do
  case ${opt} in
    f)
      helm_chart_archive=${OPTARG}
      ;;
    d)
      helm_chart_dir=${OPTARG}
      ;;
    h)
      usage
      exit 0
      ;;
    :)
      echo -e "${RED}Option -${OPTARG} requires an argument.${NO_COLOR}" >&2
      usage
      exit 1
      ;;
    *)
      echo -e "${RED}Unimplemented option: -${OPTARG}${NO_COLOR}" >&2
      usage
      exit 1
      ;;
  esac
done


# Check if all required options are provided
if [[ -z "${helm_chart_archive}" ]] && [[ -z "$helm_chart_dir" ]]; then
  helm_chart_dir=$(realpath "$dir/..")
  if [ ! -f "${helm_chart_dir}/Chart.yaml" ]; then
    echo -e "${RED}Error: -f or -d option is required. No chart found relative to script path at ${helm_chart_dir}${NO_COLOR}" >&2
    usage
    exit 1
  fi
fi

# Check if the archive exists
if [[ -n "${helm_chart_archive}" ]] && [ ! -f "${helm_chart_archive}" ]; then
  echo -e "${RED}Helm chart archive not found${NO_COLOR}: ${helm_chart_archive}" >&2
  exit 1
fi

# Check if the dir exists
if [[ -n "${helm_chart_dir}" ]] && [ ! -d "${helm_chart_dir}" ]; then
  echo -e "${RED}Helm chart directory not found${NO_COLOR}: ${helm_chart_dir}" >&2
  exit 1
fi

# Helm values to enable non-default features and get their images.
helm_values="httpHeaderInjectorWebhook.enabled=true,stackstate.apiKey=APIKEY,logsAgent.enabled=true,stackstate.cluster.name=dummy-cluster,stackstate.url=http://dummy.stackstate.io,kubernetes-rbac-agent.enabled=true"
helm_release=release

if [[ -z ${helm_chart_archive} ]]; then
  helm_chart="${helm_chart_dir}"
else
  helm_chart="${helm_chart_archive}"
fi

# Render the manifests from the Helm chart, skipping known warnings.
helm_manifests=$(helm template ${helm_release} "${helm_chart}" --set "${helm_values}" 2>/dev/stdout | grep -Ev "coalesce.go:")
# shellcheck disable=SC2181
if [ $? -ne 0 ]; then
  echo "${helm_manifests}"
  echo -e "${RED}Failed to render from Helm chart${NO_COLOR}" >&2
  exit 1
fi

# Extract images from the Helm manifests
echo "${helm_manifests}" | grep image: | sed -E 's/^.*image: ['\''"]?([^'\''"]*)['\''"]?.*$/\1/' | sort | uniq
