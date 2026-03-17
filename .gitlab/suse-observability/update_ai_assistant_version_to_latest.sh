#!/bin/bash

set -euxo pipefail

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# shellcheck disable=SC1091
source "${dir}/util.sh"
# shellcheck disable=SC1091
source "${dir}/../util.sh"

ai_assistant_version=$(get_latest_master_version suse-observability-borg "$AI_ASSISTANT_VERSION_REGEX")
values_path="stable/suse-observability/values.yaml"
readme_path="stable/suse-observability/README.md"

echo "Latest AI Assistant version: $ai_assistant_version"

updateChartValue "stackstate.components.aiAssistant.image.tag" "$ai_assistant_version" "$values_path" "$readme_path"

git add "$values_path" "$readme_path"
commit_changes "Updating AI Assistant version to $ai_assistant_version"
push_changes
