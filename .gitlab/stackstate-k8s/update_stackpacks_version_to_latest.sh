#!/bin/bash

set -euxo pipefail

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# shellcheck disable=SC1091
source "${dir}/util.sh"
# shellcheck disable=SC1091
source "${dir}/../util.sh"

full_stackpacks_master_tag=$(get_latest_master_version stackpacks "$STACKPACKS_VERSION_REGEX")
values_path="stable/stackstate-k8s/values.yaml"
readme_path="stable/stackstate-k8s/README.md"

echo "Latest stackpack master tag: $full_stackpacks_master_tag"

updateChartValue "stackstate.stackpacks.image.tag" "$full_stackpacks_master_tag" "$values_path" "$readme_path"

git add "$values_path" "$readme_path"
commit_changes "Updating stackpacks version to $full_stackpacks_master_tag"
push_changes
