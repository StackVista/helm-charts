#!/usr/bin/env bash
# Fails if any (sub)chart in a packaged chart has an unvendored dependency.
# Fleet runs a dependency update on every chart directory and rejects those.

[[ -n "${TRACE+x}" ]] && set -x

set -euo pipefail

if [ "$#" -ne 1 ] || [ ! -f "$1" ]; then
  echo "Usage: $(basename "$0") <chart.tgz>" >&2
  exit 2
fi

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
tar -xzf "$1" -C "$work"

trimTrailing() {
  printf '%s' "${1%"${1##*[! ]}"}"
}

failures=0
while IFS= read -r chartYaml; do
  chart=$(dirname "$chartYaml")
  [ "$(yq e '.dependencies | length' "$chartYaml")" = "0" ] && continue
  # Captured rather than piped so that a helm failure stops the script.
  list=$(helm dependency list "$chart")
  # Columns are tab-separated and space-padded, and REPOSITORY may be empty.
  # Lines without a tab (header aside) are helm WARNINGs, not dependencies.
  while IFS= read -r line; do
    [[ "$line" == *$'\t'* ]] || continue
    name=$(trimTrailing "${line%%$'\t'*}")
    status=$(trimTrailing "${line##*$'\t'}")
    case "$status" in
      STATUS|ok|unpacked) ;;
      *)
        echo "ERROR: ${chart#"$work"/} dependency '${name}' is ${status}" >&2
        failures=$((failures + 1))
        ;;
    esac
  done <<< "$list"
done < <(find "$work" -name Chart.yaml)

if [ "$failures" -gt 0 ]; then
  exit 1
fi

echo "OK: all dependencies are vendored in $(basename "$1")"
