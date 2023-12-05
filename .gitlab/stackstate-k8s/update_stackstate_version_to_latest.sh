#!/bin/bash

set -euxo pipefail

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# shellcheck disable=SC1091
source "${dir}/util.sh"
# shellcheck disable=SC1091
source "${dir}/../util.sh"

stackstate_master_tag=$(get_latest_master_version stackstate-server "$STACKSTATE_VERSION_REGEX")
values_path="stable/stackstate-k8s/values.yaml"
chart_path="stable/stackstate-k8s/Chart.yaml"
readme_path="stable/stackstate-k8s/README.md"

echo "Latest stackstate master tag: $stackstate_master_tag"

updateChartValue "stackstate.components.all.image.tag" "$stackstate_master_tag" "$values_path" "$readme_path"
yq -i eval ".appVersion=\"$stackstate_master_tag\"" "$chart_path"

git add "$chart_path" "$values_path" "$readme_path"
commit_changes "Updating stackstate version to $stackstate_master_tag"
push_changes
