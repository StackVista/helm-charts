#!/bin/bash
# Update helm dependencies for all charts that depend on suse-observability-sizing.
# Run this after modifying templates in suse-observability-sizing to ensure
# dependent charts pick up the changes.

set -euo pipefail

dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
cd "${dir}/../.." || exit

./scripts/bump-chart-version/bump_chart_version.py -v suse-observability-sizing --skip suse-observability
