#!/bin/bash

set -euxo pipefail

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# shellcheck disable=SC1091
source "${dir}/util.sh"
# shellcheck disable=SC1091
source "${dir}/../util.sh"

stackpacks_master_version=$(get_latest_master_version stackpacks "$STACKPACKS_VERSION_REGEX")
values_path="stable/suse-observability/values.yaml"
readme_path="stable/suse-observability/README.md"

echo "Latest stackpack master version: $stackpacks_master_version"

updateChartValue "stackstate.stackpacks.image.version" "$stackpacks_master_version" "$values_path" "$readme_path"

git add "$values_path" "$readme_path"
commit_changes "Updating stackpacks version to $stackpacks_master_version"
push_changes
