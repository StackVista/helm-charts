#!/usr/bin/env bash

set -euo pipefail

# Verifies if a version of a chart has been "bumped" if not then it returns an error.
# The script is executed when there are some changes in a chart directory

chart="$1"

remote_version=$(git show "helm/${CI_MERGE_REQUEST_TARGET_BRANCH_NAME}:stable/${chart}/Chart.yaml" | yq ".version")
new_version=$(yq ".version" "stable/${chart}/Chart.yaml")

if [ "$(printf '%s\n' "$remote_version" "$new_version" | sort -V | head -n1)" = "$new_version" ]; then
  echo "Version of the chart should be updated"
  exit 1
fi
