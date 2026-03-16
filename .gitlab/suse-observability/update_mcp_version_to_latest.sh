#!/bin/bash

set -euxo pipefail

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# shellcheck disable=SC1091
source "${dir}/util.sh"
# shellcheck disable=SC1091
source "${dir}/../util.sh"

mcp_version=$(get_latest_master_version suse-observability-mcp "$MCP_VERSION_REGEX")
values_path="stable/suse-observability/values.yaml"
readme_path="stable/suse-observability/README.md"

echo "Latest MCP version: $mcp_version"

updateChartValue "stackstate.components.mcp.image.tag" "$mcp_version" "$values_path" "$readme_path"

git add "$values_path" "$readme_path"
commit_changes "Updating MCP version to $mcp_version"
push_changes
