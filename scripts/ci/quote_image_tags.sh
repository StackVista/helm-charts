#!/usr/bin/env bash
# Rewrites a chart's top-level values.yaml so every image tag / version is a
# *quoted* YAML scalar. Idempotent: lines that are already quoted, empty, or
# null are left untouched, and only the affected value lines are modified
# (indentation, key, and any trailing inline comment are preserved).
#
# Why this exists: updatecli's yaml plugin re-serialises the value it bumps and
# only quotes what go-yaml deems ambiguous (e.g. plain integers). A short git
# SHA in `<digits>e<digits>` form (e.g. 10000e90) is left unquoted, and helm
# then coerces it to a float (1e+94), producing a wrong image reference. Running
# this as a `shell` target after the tag targets re-quotes the bumped key so the
# committed values.yaml stays string-typed. The counterpart check is
# scripts/ci/validate_quoted_tags.sh.
#
# Implemented with bash + grep only (no awk) so it runs in the minimal CI image.

[[ -n "${TRACE+x}" ]] && set -x

set -uo pipefail

GREEN='\033[0;32m'
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

# Matches "<indent>(tag|imageTag|version): <value...>" with the key anchored
# right after the indentation (so apiVersion etc. never match) and at least one
# space before the value.
keyRe='^([[:space:]]*(tag|imageTag|version):[[:space:]]+)([^[:space:]#].*)$'

tmpFile="$(mktemp)"
trap 'rm -f "${tmpFile}"' EXIT

changed=0
while IFS= read -r line || [ -n "${line}" ]; do
  if [[ "${line}" =~ ${keyRe} ]]; then
    prefix="${BASH_REMATCH[1]}"
    rest="${BASH_REMATCH[3]}"

    # Split a trailing inline comment off the value (tags never contain '#').
    comment=""
    case "${rest}" in
      *"#"*)
        value="${rest%%#*}"
        comment="#${rest#*#}"
        ;;
      *)
        value="${rest}"
        ;;
    esac
    # Trim trailing whitespace from the value.
    value="${value%"${value##*[![:space:]]}"}"

    first="${value%"${value#?}"}"
    case "${value}" in
      '' | '~' | 'null') ;;                 # nothing to quote
      *)
        if [ "${first}" != '"' ] && [ "${first}" != "'" ]; then
          if [ -n "${comment}" ]; then
            printf '%s"%s" %s\n' "${prefix}" "${value}" "${comment}" >> "${tmpFile}"
          else
            printf '%s"%s"\n' "${prefix}" "${value}" >> "${tmpFile}"
          fi
          changed=1
          continue
        fi
        ;;
    esac
  fi
  printf '%s\n' "${line}" >> "${tmpFile}"
done < "${valuesFile}"

if [ "${changed}" -eq 1 ]; then
  cat "${tmpFile}" > "${valuesFile}"
  printf "%b\n" "${GREEN}Quoted${NO_COLOR} image tags/versions in ${valuesFile}."
else
  printf "%b\n" "${GREEN}OK${NO_COLOR}: image tags/versions in ${valuesFile} already quoted."
fi
