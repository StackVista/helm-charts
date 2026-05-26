#!/bin/bash

set -euxo pipefail

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# shellcheck disable=SC1091
source "$dir/util.sh"

# The script sets helm chart version
# Usage ./gitlab/set_chart_release_version.sh chart_name [new_version]
# - chart_name - name directory with the chart to update
# - new_version - released version like x.y.z. The version will be saved to Chart.yaml file

chart=$1
new_version=${2:-}
chart_path="$chart/Chart.yaml"

# Overrides chart version to value provided as a parameter
if [[ -n "$new_version" ]]; then
  if [[ "$new_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    yq -i eval ".version=\"$new_version\"" "$chart_path"
    update_chart_version_in_readme_file "$chart"
  else
    echo "Invalid new version, expected x.y.z"
    exit 1
  fi
fi
