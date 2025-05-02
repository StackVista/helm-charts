#!/usr/bin/env bash
set -o pipefail

# shellcheck disable=SC2034
GREEN='\033[0;32m'
RED='\033[0;31m'
NO_COLOR='\033[0m'

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

function usage() {
  cat <<EOF >&2
Get the list of Docker images used by the SUSE Observability chart

Usage:
  $0 [-f suse-observability-X.Y.Z.tgz]

Arguments:
    -f : TGZ archive with the SUSE Observability chart (optional)
    -h : Show this help text
EOF
}

helm_chart_archive=$(realpath "$dir/..")

# Parse options
while getopts ":f:h" opt; do
  case ${opt} in
    f)
      helm_chart_archive=${OPTARG}

      # Check if the archive exists
      if [ ! -f "${helm_chart_archive}" ]; then
        echo -e "${RED}Helm chart archive not found${NO_COLOR}: ${helm_chart_archive}" >&2
        exit 1
      fi
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

helm_release=release

images=()
function listImages() {
  tmp_file=/tmp/o11y-tenant-get-images
  helm_values_base="anomaly-detection.enabled=true,backup.enabled=true,victoria-metrics-0.backup.enabled=true,victoria-metrics-0.restore.enabled=true,minio.accessKey=ABCDEFGH,minio.secretKey=ABCDEFGHABCDEFGH,stackstate.baseUrl=http://dummy.stackstate.io,stackstate.admin.authentication.password=dummy,stackstate.authentication.adminPassword=dummy,stackstate.license.key=dummy,global.receiverApiKey=dummy,clickhouse.enabled=true,opentelemetry.enabled=true,kubernetes-rbac-agent.enabled=true"
  # hbase in Distributed mode
  helm_values="$helm_values_base"
  helm template "$helm_release" "$helm_chart_archive" --set "$helm_values" | grep image: | sed -E 's/^.*image: ['\''"]?([^'\''"]*)['\''"]?.*$/\1/' > "$tmp_file"
  while IFS='' read -r line; do images+=("$line"); done < "$tmp_file"

  # hbase in Mono mode
  helm_values="$helm_values_base,hbase.deployment.mode=Mono"
  helm template "$helm_release" "$helm_chart_archive" --set "$helm_values" | grep image: | sed -E 's/^.*image: ['\''"]?([^'\''"]*)['\''"]?.*$/\1/' > "$tmp_file"
  while IFS='' read -r line; do images+=("$line"); done < "$tmp_file"

  # Remove duplicates
  IFS=" " read -r -a images <<< "$(echo "${images[@]}" | tr ' ' '\n' | sort -u | tr '\n' ' ')"
  rm -f "$tmp_file"
}

listImages
for image in "${images[@]}"
do
  echo "$image"
done
