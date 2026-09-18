#!/usr/bin/env bash
# Usage: latest-so-tag.sh <image>
#
# Returns the latest x.y.z-soN or x.y.z.w-soN tag for quay.io/stackstate/<image>.
# "Latest" means: highest upstream version (compared numerically per component),
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
    | jq -Rrs '
        split("\n")
        | map(select(test("^[0-9]+\\.[0-9]+\\.[0-9]+(\\.[0-9]+)?-so[0-9]+$")))
        | if length == 0 then error("No release tags found")
          else max_by(
            split("-so") as $tag
            | ($tag[0] | split(".") | map(tonumber)) as $version
            | [$version[0], $version[1], $version[2], ($version[3] // 0), ($tag[1] | tonumber)]
          )
          end
      '
