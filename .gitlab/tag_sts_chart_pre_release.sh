#!/bin/bash

set -euxo pipefail

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# shellcheck disable=SC1091
source "$dir/util.sh"

# The script tags the current head as a pre-release, if the provided chart was published as a prerelease.
# Usage ./gitlab/tag_sts_chart_pre_release.sh chart_name
# - chart - name directory with the chart to update
# The flow:
#    -  We just published a prerelease from the current head.
#    -  Push a tag representing the current pre-release

chart_name=$1
chart="stable/$chart_name"
chart_path="$chart/Chart.yaml"
current_version=$(yq ".version" "$chart_path")

# There is "pre" version like `x.y.z-pre.v` so the new version will be `x.y.z-pre.(v+1)`
if [[ "$current_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+-pre\.[0-9]+$ ]]; then
  echo "Found pre-release version '$current_version', tagging.."
  tag="prerelease/$chart_name/$current_version"
  git tag -a "$tag" -m "$tag"
  push_tag_skip_ci
# There is "released" version like `x.y.z` so the new version will be `x.y.(z+1)-pre.1`
elif [[ "$current_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Published release version, not need to tag"
else
  echo "Not supported chart version, you have to fix it first to run the script. Expected x.y.z or x.y.z-pre.v"
  exit 1
fi
