#!/usr/bin/env bash
set -o pipefail

# shellcheck disable=SC2034
GREEN='\033[0;32m'
RED='\033[0;31m'
NO_COLOR='\033[0m'

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# if Macos, use gsed
sed_bin="sed"
if [ "$(uname)" == "Darwin" ] ;
then
  sed_bin="gsed"
fi

function usage() {
  cat <<EOF >&2
Change the registry and repository in the helm chart values file. This is meant as a pre-publishing step when the chart
gets published to an alternative repository.

Usage:
  $0 -g docker_registry -p docker_repository

Arguments:
    -g : Docker registry to change to
    -p : Docker repository to change to
    -h : Show this help text
EOF
}

# Parse options
while getopts ":g:p:h" opt; do
  case ${opt} in
    g)
        docker_registry_to=${OPTARG}
        ;;
    p)
        docker_repo_to=${OPTARG}
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

helm_chart_dir=$(realpath "$dir/..")

# Check if all required options are provided
if [[ -z "${docker_registry_to}" ]]; then
  echo -e "${RED}Error: -g option is required.${NO_COLOR}" >&2
  usage
  exit 1
fi

if [[ -z "${docker_repo_to}" ]]; then
  echo -e "${RED}Error: -p option is required.${NO_COLOR}" >&2
  usage
  exit 1
fi

function updateDockerImagesInValues() {
  echo "Updating values.yaml to use new docker images"
  ${sed_bin} -i "s#registry: [^/:]\+\$#registry: $docker_registry_to#g" "$helm_chart_dir/values.yaml"
  ${sed_bin} -i "s#repository: [^/:]\+\/[^/:]\+\/\([^/:]\+\)\$#repository: $docker_registry_to/$docker_repo_to/\1#g" "$helm_chart_dir/values.yaml"
  ${sed_bin} -i "s#repository: [^/:]\+\/\([^/:]\+\)\$#repository: $docker_repo_to/\1#g" "$helm_chart_dir/values.yaml"
  ${sed_bin} -i "s#spotlightRepository: [^/:]\+\/\([^/:]\+\)\$#spotlightRepository: $docker_repo_to/\1#g" "$helm_chart_dir/values.yaml"

  # Set the global image registry
  DOCKER_REGISTRY_TO="$docker_registry_to" yq -i -e  '.global.imageRegistry = env(DOCKER_REGISTRY_TO)' "$helm_chart_dir/values.yaml"

  # Update elasticsearch
  yq e -i ".elasticsearch.imageRegistry = \"$docker_registry_to\"" "$helm_chart_dir/values.yaml"
  yq e -i ".elasticsearch.imageRepository = \"$docker_repo_to/elasticsearch\"" "$helm_chart_dir/values.yaml"
  yq e -i ".elasticsearch.prometheus-elasticsearch-exporter.image.registry = \"$docker_registry_to\"" "$helm_chart_dir/values.yaml"
  yq e -i ".elasticsearch.prometheus-elasticsearch-exporter.image.repository = \"$docker_repo_to/elasticsearch-exporter\"" "$helm_chart_dir/values.yaml"
  # Update hbase
  yq e -i ".hbase.all.image.registry = \"$docker_registry_to\"" "$helm_chart_dir/values.yaml"
  yq e -i ".hbase.wait.image.registry = \"$docker_registry_to\"" "$helm_chart_dir/values.yaml"
  yq e -i ".hbase.zookeeper.image.registry = \"$docker_registry_to\"" "$helm_chart_dir/values.yaml"
  yq e -i ".hbase.stackgraph.image.repository = \"$docker_repo_to/stackgraph-hbase\"" "$helm_chart_dir/values.yaml"
  yq e -i ".hbase.console.image.repository = \"$docker_repo_to/stackgraph-console\"" "$helm_chart_dir/values.yaml"
  yq e -i ".hbase.wait.image.repository = \"$docker_repo_to/wait\"" "$helm_chart_dir/values.yaml"
  yq e -i ".hbase.hbase.master.image.repository = \"$docker_repo_to/hbase-master\"" "$helm_chart_dir/values.yaml"
  yq e -i ".hbase.hbase.regionserver.image.repository = \"$docker_repo_to/hbase-regionserver\"" "$helm_chart_dir/values.yaml"
  yq e -i ".hbase.hdfs.image.repository = \"$docker_repo_to/hadoop\"" "$helm_chart_dir/values.yaml"
  yq e -i ".hbase.tephra.image.repository = \"$docker_repo_to/tephra-server\"" "$helm_chart_dir/values.yaml"
  yq e -i ".hbase.zookeeper.image.repository = \"$docker_repo_to/zookeeper\"" "$helm_chart_dir/values.yaml"
  # Update Victoria Metrics
  yq e -i ".victoria-metrics-0.server.image.registry = \"$docker_registry_to\"" "$helm_chart_dir/values.yaml"
  yq e -i ".victoria-metrics-0.server.image.repository = \"$docker_repo_to/victoria-metrics\"" "$helm_chart_dir/values.yaml"
  yq e -i ".victoria-metrics-1.server.image.registry = \"$docker_registry_to\"" "$helm_chart_dir/values.yaml"
  yq e -i ".victoria-metrics-1.server.image.repository = \"$docker_repo_to/victoria-metrics\"" "$helm_chart_dir/values.yaml"
  yq e -i ".victoria-metrics-0.backup.vmbackup.image.registry = \"$docker_registry_to\"" "$helm_chart_dir/values.yaml"
  yq e -i ".victoria-metrics-0.backup.vmbackup.image.repository = \"$docker_repo_to/vmbackup\"" "$helm_chart_dir/values.yaml"
  yq e -i ".victoria-metrics-1.backup.vmbackup.image.registry = \"$docker_registry_to\"" "$helm_chart_dir/values.yaml"
  yq e -i ".victoria-metrics-1.backup.vmbackup.image.repository = \"$docker_repo_to/vmbackup\"" "$helm_chart_dir/values.yaml"
  yq e -i ".victoria-metrics-0.backup.setupCron.image.registry = \"$docker_registry_to\"" "$helm_chart_dir/values.yaml"
  yq e -i ".victoria-metrics-0.backup.setupCron.image.repository = \"$docker_repo_to/container-tools\"" "$helm_chart_dir/values.yaml"
  yq e -i ".victoria-metrics-1.backup.setupCron.image.registry = \"$docker_registry_to\"" "$helm_chart_dir/values.yaml"
  yq e -i ".victoria-metrics-1.backup.setupCron.image.repository = \"$docker_repo_to/container-tools\"" "$helm_chart_dir/values.yaml"

  echo "Updating documentations (README.md)"
  (cd  "$helm_chart_dir" && helm-docs)
}

updateDockerImagesInValues
