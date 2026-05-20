#!/usr/bin/env bash
# Verify that every dependency in <chart>/Chart.yaml that points to a chart
# under local/ is pinned to version "*". Local subcharts are not independently
# versioned; pinning to a specific version defeats the whole purpose of the
# local/ layout.

[[ -n "${TRACE+x}" ]] && set -x

set -euo pipefail

if [ "$#" -ne 1 ] || [ -z "$1" ]; then
  echo "Usage: $(basename "$0") <chart-path>" >&2
  exit 2
fi

chartPath="${1%/}"

if [ ! -f "${chartPath}/Chart.yaml" ]; then
  echo "Error: '${chartPath}' is not a chart directory (no Chart.yaml found)" >&2
  exit 2
fi

repoRoot=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
localRoot="${repoRoot}/local"

if [ ! -d "${localRoot}" ]; then
  exit 0
fi

failures=0

# Emit one dep per line as `name<TAB>repository<TAB>version`.
while IFS=$'\t' read -r name repo version; do
  [ -z "${repo}" ] && continue
  case "${repo}" in
    file://*) ;;
    *) continue ;;
  esac

  relPath="${repo#file://}"
  absDepPath=$(realpath -m "${chartPath}/${relPath}")

  # Only flag direct children of local/, e.g. local/common. Skip nested paths
  # like local/clickhouse/charts/common (the vendored Bitnami common).
  if [ "$(dirname "${absDepPath}")" != "${localRoot}" ]; then
    continue
  fi

  if [ "${version}" != "*" ]; then
    echo "ERROR: ${chartPath}/Chart.yaml dependency '${name}' (repository '${repo}') references a local chart but pins version '${version}'. Local-chart deps must use version: \"*\"." >&2
    failures=$((failures + 1))
  fi
done < <(yq e '.dependencies[]? | [.name, .repository, (.version | tostring)] | @tsv' "${chartPath}/Chart.yaml")

if [ "${failures}" -gt 0 ]; then
  exit 1
fi

printf "OK: all local-chart deps in %s/Chart.yaml are pinned to \"*\"\n" "${chartPath}"
