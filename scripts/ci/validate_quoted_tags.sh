#!/usr/bin/env bash
# Validates that every image tag / version in a chart's values.yaml is written
# as a *quoted* YAML scalar.
#
# Why this matters: updatecli writes image tags into values.yaml as plain
# scalars. A short git SHA that happens to be all digits (e.g. 88123456) or in
# scientific form (e.g. 60e12345) is then coerced by helm's YAML parser into a
# number/float (8.8123456e+07), producing a wrong image reference. Keeping the
# value quoted forces it to stay a string, and updatecli preserves the existing
# quote style on subsequent bumps.
#
# The check is line-based on the chart's top-level values.yaml: any non-empty
# inline value for a `tag:`, `imageTag:` or `version:` key must start with a
# single or double quote. Implemented with grep + bash only (no awk) so it runs
# in the minimal CI image.

[[ -n "${TRACE+x}" ]] && set -x

set -uo pipefail

GREEN='\033[0;32m'
RED='\033[0;31m'
NO_COLOR='\033[0m'

usage() {
  echo "Usage: $(basename "$0") <chart-path>" >&2
  echo "Example: $(basename "$0") stable/suse-observability" >&2
  exit 2
}

if [ "$#" -ne 1 ] || [ -z "$1" ]; then
  usage
fi

chartPath="${1%/}"

if [ ! -f "${chartPath}/Chart.yaml" ]; then
  echo "Error: '${chartPath}' is not a chart directory (no Chart.yaml found)" >&2
  exit 2
fi

valuesFile="${chartPath}/values.yaml"
if [ ! -f "${valuesFile}" ]; then
  echo "Error: values file not found: ${valuesFile}" >&2
  exit 2
fi

printf "\nValidating quoted image tags/versions in %s...\n\n" "${valuesFile}"

# Candidate lines: a tag/imageTag/version key (anchored right after indentation,
# so keys like apiVersion are not matched) followed by an inline value. Comment
# lines never match. grep -n prefixes each with "LINENO:".
candidates=$(grep -nE '^[[:space:]]*(tag|imageTag|version):[[:space:]]' "${valuesFile}" || true)

violations=""
while IFS= read -r record; do
  [ -z "${record}" ] && continue
  lineno="${record%%:*}"
  content="${record#*:}"

  # Value is everything after the first colon.
  value="${content#*:}"
  # Trim leading whitespace.
  value="${value#"${value%%[![:space:]]*}"}"
  # Strip a trailing inline comment (" #...").
  case "${value}" in
    *" #"*) value="${value%% #*}" ;;
  esac
  # Trim trailing whitespace.
  value="${value%"${value##*[![:space:]]}"}"

  # Empty / null values are fine (overridable defaults).
  case "${value}" in
    '' | '~' | 'null') continue ;;
  esac

  # First character: a quote means it is already a string scalar.
  first="${value%"${value#?}"}"
  case "${first}" in
    '"' | "'") ;;
    *) violations="${violations}${lineno}:${content}"$'\n' ;;
  esac
done <<EOF
${candidates}
EOF

if [ -n "${violations}" ]; then
  printf "%b\n" "${RED}FAIL${NO_COLOR}: found unquoted image tag/version value(s) in ${valuesFile}:"
  while IFS= read -r v; do
    [ -z "${v}" ] && continue
    printf "  %s:%s\n" "${valuesFile}" "${v%%:*}"
  done <<EOF
${violations}
EOF
  printf '\n'
  while IFS= read -r v; do
    [ -z "${v}" ] && continue
    printf "    %s\n" "${v#*:}"
  done <<EOF
${violations}
EOF
  printf "\nQuote the value(s) above (e.g. tag: \"04404825\") so YAML keeps them as strings.\n"
  exit 1
fi

printf "%b\n" "${GREEN}OK${NO_COLOR}: all image tags/versions in ${valuesFile} are quoted."
