#!/usr/bin/env bash
# Usage: latest-so-tag.sh <image>
#
# Returns the latest x.y.z-soN tag for quay.io/stackstate/<image>.
# "Latest" means: highest x.y.z (compared numerically per component),
# then highest soN (compared numerically).
#
# Rationale: updatecli's built-in versionfilter kinds (regex/semver, latest)
# compare the soN pre-release identifier lexicographically, causing
# "so10" < "so9" (ASCII '1' < '9').  This script sorts all fields numerically.
set -euo pipefail

image="${1?Usage: $0 <image>}"
# QUAY_BASE_URL can be overridden in tests to point at a mock server.
QUAY_BASE_URL="${QUAY_BASE_URL:-https://quay.io}"
page=1
all_tags=""

while true; do
    resp=$(curl -sf \
        "${QUAY_BASE_URL}/api/v1/repository/stackstate/${image}/tag/?onlyActiveTags=true&limit=100&page=${page}")
    page_tags=$(printf '%s' "$resp" | jq -r '.tags[].name')
    all_tags="${all_tags}${page_tags}"$'\n'
    has_more=$(printf '%s' "$resp" | jq -r '.has_additional')
    [ "$has_more" = "true" ] || break
    page=$((page + 1))
done

printf '%s' "$all_tags" \
    | grep -E '^[0-9]+\.[0-9]+\.[0-9]+-so[0-9]+$' \
    | awk -F'[.-]' '{n=$4; sub(/^so/,"",n); printf "%010d%010d%010d%010d\t%s\n",$1,$2,$3,n,$0}' \
    | sort | tail -1 | cut -f2
