#!/bin/bash
# Update helm dependencies for all charts that depend on suse-observability-sizing.
# Run this after modifying templates in suse-observability-sizing to ensure
# dependent charts pick up the changes.

set -euo pipefail

dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
cd "${dir}/../.." || exit

dependent_charts=()
for chart_yaml in stable/*/Chart.yaml; do
  chart_dir=$(dirname "$chart_yaml") chart_name=$(basename "$chart_dir")

  if [[ "$chart_name" == "suse-observability-sizing" ]]; then
    continue
  fi

  if grep -q 'suse-observability-sizing' "$chart_yaml"; then
    if [[ "$chart_name" == "suse-observability" ]]; then
      continue
    fi
    dependent_charts+=("$chart_dir")
  fi
done

for chart_dir in "${dependent_charts[@]}"; do
  echo "Updating dependencies for $(basename "$chart_dir")..."
  if [[ -z "$chart_dir" || "$chart_dir" != stable/* ]]; then
    echo "ERROR: unexpected chart_dir '$chart_dir', skipping" >&2
    continue
  fi
  rm -f "$chart_dir"/charts/suse-observability-sizing-*.tgz
  helm dependency update "$chart_dir"
done
