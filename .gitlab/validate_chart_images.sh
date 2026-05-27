#!/usr/bin/env bash
# Validates that every docker image referenced by a published chart resolves
# in its registry. Image enumeration is delegated to the chart's own
# installation/o11y-*-get-images.sh helper; existence is checked anonymously
# with `skopeo inspect`. Anonymous-only: private images that 401 will be
# reported as MISSING.

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

case "$(basename "${chartPath}")" in
  suse-observability)
    listImagesScript="${chartPath}/installation/o11y-get-images.sh"
    ;;
  suse-observability-agent)
    listImagesScript="${chartPath}/installation/o11y-agent-get-images.sh"
    ;;
  *)
    echo "Error: '${chartPath}' has no known list-images script wired up in $(basename "$0")" >&2
    exit 2
    ;;
esac

if [ ! -x "${listImagesScript}" ]; then
  echo "Error: list-images script not found or not executable: ${listImagesScript}" >&2
  exit 2
fi

for cmd in helm skopeo xargs; do
  if ! command -v "${cmd}" >/dev/null 2>&1; then
    echo "Error: required command '${cmd}' not found in PATH" >&2
    exit 2
  fi
done

printf "\nEnumerating images from %s via %s...\n" "${chartPath}" "${listImagesScript}"

imagesFile=$(mktemp)
trap 'rm -f "${imagesFile}"' EXIT

if ! "${listImagesScript}" | sort -u > "${imagesFile}"; then
  echo "Error: ${listImagesScript} failed" >&2
  exit 1
fi

imageCount=$(grep -cve '^[[:space:]]*$' "${imagesFile}" || true)
if [ "${imageCount}" -eq 0 ]; then
  echo "Error: ${listImagesScript} produced no images" >&2
  exit 1
fi

printf "Found %s unique image(s). Checking each with 'skopeo inspect' (anonymous)...\n\n" "${imageCount}"

# Parallel per-image check. Each worker prints one line: "OK <image>" or
# "MISSING <image>\t<short error>". Failures also accumulate exit-status info
# via the wrapper's exit code, which xargs propagates to its own exit code.
resultsFile=$(mktemp)
trap 'rm -f "${imagesFile}" "${resultsFile}"' EXIT

checkImage() {
  local image="$1"
  local out
  if out=$(skopeo inspect "docker://${image}" 2>&1 >/dev/null); then
    printf '%b\n' "${GREEN}OK${NO_COLOR}\t${image}"
  else
    local short
    short=$(printf '%s' "${out}" | tr '\n' ' ' | cut -c1-200)
    printf '%b\n' "${RED}MISSING${NO_COLOR}\t${image}\t${short}"
    return 1
  fi
}
export -f checkImage
export GREEN RED NO_COLOR

# xargs exits non-zero (123) if any invocation exited non-zero.
xargsExit=0
# shellcheck disable=SC2016
xargs -a "${imagesFile}" -P 8 -I{} bash -c 'checkImage "$@"' _ {} > "${resultsFile}" || xargsExit=$?

# Stable, readable output: OK lines first (sorted), then MISSING lines (sorted).
grep -E $'^\033\\[0;32m' "${resultsFile}" | sort
missing=$(grep -E $'^\033\\[0;31m' "${resultsFile}" | sort)
if [ -n "${missing}" ]; then
  printf '%s\n' "${missing}"
fi

missingCount=$(printf '%s' "${missing}" | grep -cve '^$' || true)
okCount=$((imageCount - missingCount))

printf "\n"
if [ "${missingCount}" -gt 0 ] || [ "${xargsExit}" -ne 0 ]; then
  printf "%b\n" "${RED}FAIL${NO_COLOR}: ${missingCount}/${imageCount} image(s) failed to resolve."
  exit 1
fi

printf "%b\n" "${GREEN}OK${NO_COLOR}: all ${okCount} image(s) resolved."
