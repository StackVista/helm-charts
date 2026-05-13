#!/bin/bash

set -euxo pipefail

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# shellcheck disable=SC1091
source "$dir/util.sh"

# The script appends "-pre" to the current helm chart version
# Usage ./gitlab/modify_chart_to_prerelease_version.sh chart_name
# - chart_name - name of the chart under stable/ to update
# The flow:
#    -  there is "current" version like x.y.z then "-pre" suffix is added, e.g. 1.5.2 => 1.5.2-pre

chart=$1
chart_path="stable/$chart/Chart.yaml"
current_version=$(yq ".version" "$chart_path")

if [[ "$current_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  yq -i eval ".version=\"${current_version}-pre\"" "$chart_path"
else
  echo "Not supported chart version, expected x.y.z (got '$current_version')"
  exit 1
fi

update_chart_version_in_readme_file "stable/$chart"
