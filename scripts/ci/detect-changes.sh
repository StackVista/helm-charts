#!/usr/bin/env bash
# Detect which stable / local charts are affected by the current diff and emit
# a JSON matrix. Writes to $GITHUB_OUTPUT if set, else stdout.
#
# Inputs (env):
#   BASE_SHA              optional override for the diff base
#   GITHUB_EVENT_BEFORE   set by GH Actions on push events
#   GITHUB_OUTPUT         set by GH Actions; receives the four outputs
#
# Outputs:
#   stable                   JSON array of chart names under stable/ (independently published)
#   local                    JSON array of chart names under local/ (consumed only as file:// deps)
#   resource_usage_changed   "true" / "false"
#
# Note: updatecli changes are detected at the workflow level instead (via a
# `push: paths: ['updatecli/**']` filter in .github/workflows/updatecli.yml).

set -euo pipefail

# Trust the workspace path regardless of who owns it on disk. In GitHub Actions
# `container:` jobs we run as root (--user 0:0 to bypass /__w/_temp EACCES) but
# the workspace directory was created by the runner pod's fsGroup user, so git
# would otherwise refuse with "dubious ownership in repository".
git config --global --add safe.directory '*'

# Tools this script needs that aren't pre-installed in sts-ci-images:stackstate-helm-test:
# - jsonnet (to render the chart list from .libsonnet)
# - jq      (to slice the rendered JSON into matrix arrays)
# Install whichever ones are missing. Alpine package names match the binary names;
# on macOS for local dev `brew install jsonnet jq` covers it.
need=()
for t in jsonnet jq; do
  command -v "$t" >/dev/null 2>&1 || need+=("$t")
done
if [ ${#need[@]} -gt 0 ]; then
  if command -v apk >/dev/null 2>&1; then
    apk add --no-cache "${need[@]}" >/dev/null
  elif command -v apt-get >/dev/null 2>&1; then
    apt-get update -qq && apt-get install -y --no-install-recommends "${need[@]}" >/dev/null
  else
    echo "Missing tools (${need[*]}) and no supported package manager (apk/apt-get) available." >&2
    echo "Install them first (e.g. 'brew install ${need[*]}' on macOS)." >&2
    exit 1
  fi
fi

# --- 1. Determine diff base ---
base="${BASE_SHA:-${GITHUB_EVENT_BEFORE:-}}"
if [[ -z "$base" || "$base" =~ ^0+$ ]] || ! git merge-base --is-ancestor "$base" HEAD 2>/dev/null; then
  base="$(git rev-parse origin/master 2>/dev/null || git rev-parse HEAD)"
fi

# --- 2. Compute changed files (three-dot: merge-base..HEAD) ---
mapfile -t changed < <(git diff --name-only "$base"...HEAD || true)

# --- 3. Render chart list from jsonnet vars (single source of truth) ---
# Output keys `stable` / `local` mirror the on-disk directory names (`stable/`,
# `local/`). The jsonnet input uses GitLab's older `public_charts`/`internal_charts`
# nomenclature; both categories are independently published, so we union them.
charts_json="$(jsonnet -J jsonnet-vendor -e '
  local v = import ".jsonnet-libs/extras/helm_chart_repo/variables.libsonnet";
  {
    stable: std.objectFields(v.helm.public_charts) + std.objectFields(v.helm.internal_charts),
    local_charts: std.set(std.flattenArrays(
      [v.helm.public_charts[k] for k in std.objectFields(v.helm.public_charts)] +
      [v.helm.internal_charts[k] for k in std.objectFields(v.helm.internal_charts)]
    )),
    deps: {[k]: v.helm.public_charts[k] for k in std.objectFields(v.helm.public_charts)}
        + {[k]: v.helm.internal_charts[k] for k in std.objectFields(v.helm.internal_charts)},
  }
')"

mapfile -t all_stable < <(echo "$charts_json" | jq -r '.stable[]')
mapfile -t all_local  < <(echo "$charts_json" | jq -r '.local_charts[]')

# --- 4. CI-wide rebuild rule ---
ci_wide=false
for f in "${changed[@]}"; do
  case "$f" in
    .github/workflows/*|scripts/ci/*|.jsonnet-libs/*) ci_wide=true; break ;;
  esac
done

# --- 5. Map changed files to affected charts ---
affected_stable=()
affected_local=()

if $ci_wide || [ "${#changed[@]}" -eq 0 ]; then
  # Over-fire on CI-wide change OR empty diff (first-run / base==HEAD).
  affected_stable=("${all_stable[@]}")
  affected_local=("${all_local[@]}")
else
  for chart in "${all_stable[@]}"; do
    mapfile -t deps < <(echo "$charts_json" | jq -r --arg c "$chart" '.deps[$c][]?')
    for f in "${changed[@]}"; do
      if [[ "$f" == stable/$chart/* ]]; then affected_stable+=("$chart"); break; fi
      hit=false
      for dep in "${deps[@]}"; do
        if [[ "$f" == local/$dep/* ]]; then hit=true; break; fi
      done
      if $hit; then affected_stable+=("$chart"); break; fi
    done
  done
  for chart in "${all_local[@]}"; do
    for f in "${changed[@]}"; do
      if [[ "$f" == local/$chart/* ]]; then affected_local+=("$chart"); break; fi
    done
  done
fi

# --- 6. Ancillary scopes ---
resource_usage_changed=false
for f in "${changed[@]}"; do
  case "$f" in
    test/*|stable/suse-observability-values/*) resource_usage_changed=true ;;
  esac
done
for c in "${affected_stable[@]}"; do
  [[ "$c" == "suse-observability" ]] && resource_usage_changed=true
done

# --- 7. Emit outputs ---
# Convert bash arrays -> JSON arrays. jq -R reads raw lines, jq -s slurps to array,
# then we filter out the empty-string entry that arises from "${arr[@]:-}" on empty arrays.
stable_json="$(printf '%s\n' "${affected_stable[@]:-}" | jq -R . | jq -sc 'map(select(length > 0))')"
local_json="$(printf '%s\n' "${affected_local[@]:-}"  | jq -R . | jq -sc 'map(select(length > 0))')"

emit() {
  if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
    {
      echo "stable=$stable_json"
      echo "local=$local_json"
      echo "resource_usage_changed=$resource_usage_changed"
    } >> "$GITHUB_OUTPUT"
  else
    echo "stable=$stable_json"
    echo "local=$local_json"
    echo "resource_usage_changed=$resource_usage_changed"
  fi
}
emit
