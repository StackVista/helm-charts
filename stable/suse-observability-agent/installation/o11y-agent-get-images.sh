#!/usr/bin/env bash
set -o pipefail

# shellcheck disable=SC2034
GREEN='\033[0;32m'
RED='\033[0;31m'
NO_COLOR='\033[0m'


function usage() {
  cat <<EOF >&2
Get the list of Docker images used by the SUSE Observability Agent chart

Usage:
  $0 -f suse-observability-agent-X.Y.Z.tgz

Arguments:
    -f : TGZ archive with the SUSE Observability Agent chart (required)
    -h : Show this help text
EOF
}

# Parse options
while getopts ":f:h" opt; do
  case ${opt} in
    f)
      helm_chart_archive=${OPTARG}
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
if [[ -z "${helm_chart_archive}" ]]; then
  echo -e "${RED}Error: -f option is required.${NO_COLOR}" >&2
  usage
  exit 1
fi

# Check if the archive exists
if [ ! -f "${helm_chart_archive}" ]; then
  echo -e "${RED}Helm chart archive not found${NO_COLOR}: ${helm_chart_archive}" >&2
  exit 1
fi

# Helm values to enable non-deafult features and get their images.
helm_values="http-header-injector-webhook.enabled=true,stackstate.apiKey=APIKEY,logsAgent.enabled=true,stackstate.cluster.name=dummy-cluster,stackstate.url=http://dummy.stackstate.io"
helm_release=release

# Render the manifests from the Helm chart, skipping known warnings.
helm_manifests=$(helm template ${helm_release} "${helm_chart_archive}" --set "${helm_values}" 2>/dev/stdout | grep -Ev "coalesce.go:")
# shellcheck disable=SC2181
if [ $? -ne 0 ]; then
  echo "${helm_manifests}"
  echo -e "${RED}Failed to render from Helm chart${NO_COLOR}" >&2
  exit 1
fi

# Extract images from the Helm manifests
echo "${helm_manifests}" | grep image: | sed -E 's/^.*image: ['\''"]?([^'\''"]*)['\''"]?.*$/\1/' | sort | uniq
