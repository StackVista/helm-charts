#!/usr/bin/env bash
# Builds the file:// dependencies of <chart> recursively, deepest first. Helm
# vendors only one level, so without this a nested local chart ships without its
# own dependencies. <chart> itself is left to the caller's `helm dependencies`.

[[ -n "${TRACE+x}" ]] && set -x

set -euo pipefail

if [ "$#" -ne 1 ] || [ -z "$1" ]; then
  echo "Usage: $(basename "$0") <chart-path>" >&2
  exit 2
fi

declare -A built=()

fileDeps() {
  yq e '.dependencies[]? | select(.repository == "file*").repository | sub("^file://","")' "$1/Chart.yaml"
}

buildTree() {
  local chart
  chart=$(realpath "$1")
  [ -n "${built[$chart]+x}" ] && return 0
  built[$chart]=1

  local deps
  deps=$(fileDeps "$chart")
  [ -z "$deps" ] && return 0

  local dep
  while IFS= read -r dep; do
    buildTree "$chart/$dep"
  done <<< "$deps"

  helm dependencies build --skip-refresh "$chart"
}

root="${1%/}"
while IFS= read -r dep; do
  [ -n "$dep" ] && buildTree "$root/$dep"
done < <(fileDeps "$root")
