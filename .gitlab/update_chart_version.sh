#!/usr/bin/env bash

set -euxo pipefail

CHART_PATH=${1?Please specify the chart you want to update}
DEPENDENCY_NAME=${2?Please specify the chart name}
DEPENDENCY_VERSION=${3?Please Specify the version of dependency}

bumping_chart() {
  chart_path=${1?Please specify the chart you want to update}
  dependency_chart_name=${2?Please specify the dependency chart name}
  dependency_version=${3?Please Specify the version of dependency}
  # Updates dependency version in the chart
  echo "Upgrading dependency '$dependency_chart_name' for '${chart_path}'."
  echo "New dependency version: ${dependency_version}."
  readme="${chart_path}/README.md"
  readme_out="${readme}.out"
  chart_yaml="${chart_path}/Chart.yaml"
  chart_lock="${chart_path}/Chart.lock"

  yq e "(.dependencies[] | select(.name == \"$dependency_chart_name\")).version = \"${dependency_version}\"" -i "${chart_yaml}"

  sed -E "s/$dependency_chart_name \| [^\|]* \|/$dependency_chart_name | $dependency_version |/" "${readme}" > "${readme_out}"
  mv "$readme_out" "$readme"

  helm dependency update "$chart_path"

  git add "${chart_yaml}" "${readme}" "${chart_lock}"
}

increment_helm_chart_version() {
  chart_path=${1?Please specify the chart you want to update}
  # Increments snapshot version of helm chart and updates readme

  readme="${chart_path}/README.md"
  readme_out="${readme}.out"
  chart_yaml="${chart_path}/Chart.yaml"

  chart_version=$(yq e .version "${chart_yaml}")

  if [[ "$chart_version" =~ ^(([0-9]+\.[0-9]+\.[0-9]+-snapshot\.)([0-9]+))$ ]];
  then
    snapshot_number="${BASH_REMATCH[3]}"
    prefix="${BASH_REMATCH[2]}"
    echo "SNAPSHOT_VERSION: ${snapshot_number}"
    ((incremented_number=snapshot_number+1))
    incremented_version="${prefix}${incremented_number}"

    echo "New chart version: ${incremented_version}"
    yq e ".version = \"${incremented_version}\"" -i "${chart_yaml}"
    sed -E "s/^Current chart version is .*$/Current chart version is \`${incremented_version}\`/" "${readme}" > "${readme_out}"
    mv "$readme_out" "$readme"

    git add "${chart_yaml}" "${readme}"
  else
    echo "Not incrementing snapshot number, the version is not snapshot: ${chart_version}"
  fi
}

bumping_chart "$CHART_PATH" "$DEPENDENCY_NAME" "$DEPENDENCY_VERSION"
increment_helm_chart_version "$CHART_PATH"
