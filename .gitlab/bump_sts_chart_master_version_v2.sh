#!/bin/bash

set -euxo pipefail

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# shellcheck disable=SC1091
source "$dir/util.sh"

# The script "bumps" helm chart version
# Usage ./gitlab/bump_sts_chart_master_version_v2.sh chart_name [new_version]
# - chart_name - name directory with the chart to update
# - new_version - optional "released" version like x.y.z. The version will be saved to Chart.yaml file
# The flow:
# - if the script is executed with 'new_version' then that version will be stored in Chart.yaml file
# - if the script is executed without 'new_version' and
#    -  there is "released" version just now then a next "pre" version will be created, e.g. 1.5.2 => 1.5.3-pre.1
#    -  there is "pre" version just now then a next "pre" version will be created, e.g. 1.5.3-pre.1 => 1.5.3-pre.2

chart=$1
new_version=${2:-}
chart_path="$chart/Chart.yaml"

# Overrides chart version to value provided as a parameter
if [[ -n "$new_version" ]]; then
  if [[ "$new_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    yq -i eval ".version=\"$new_version\"" "$chart_path"
  else
    echo "Invalid new version, expected x.y.z"
    exit 1
  fi
else #Increment current chart version to the next one
  current_version=$(yq ".version" "$chart_path")

  # There is "pre" version like `x.y.z-pre.v` so the new version will be `x.y.z-pre.(v+1)`
  if [[ "$current_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+-pre\.[0-9]+$ ]]; then
    current_pre_counter=$(sed -E 's/^[0-9]+\.[0-9]+\.[0-9]+-pre\.([0-9]+)$/\1/' <<< "$current_version")
    echo "$current_pre_counter"
    new_pre_counter=$((current_pre_counter+1))
    sed -i -E "s/^(version: [0-9]+\.[0-9]+\.[0-9]+-pre\.)[0-9]+$/\1${new_pre_counter}/" "$chart_path"
  # There is "released" version like `x.y.z` so the new version will be `x.y.(z+1)-pre.1`
  elif [[ "$current_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    current_patch_counter=$(sed -E 's/^[0-9]+\.[0-9]+\.([0-9])$/\1/' <<< "$current_version")
    new_patch_counter=$((current_patch_counter+1))
    sed -i -E "s/^(version: [0-9]+\.[0-9]+\.)[0-9]+$/\1${new_patch_counter}-pre.1/" "$chart_path"
  else
    echo "Not supported chart version, you have to fix it first to run the script. Expected x.y.z or x.y.z-pre.v"
    exit 1
  fi
fi

new_version=$(yq ".version" "$chart_path")
git fetch --all
git add "$chart_path"
commit_changes "Updating '$chart' helm chart version to $new_version"
push_changes
