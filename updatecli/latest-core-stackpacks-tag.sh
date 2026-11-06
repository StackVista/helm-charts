#!/usr/bin/env bash
set -euo pipefail

branch="${STACKPACKS_CORE_SOURCE_BRANCH:?STACKPACKS_CORE_SOURCE_BRANCH is required}"
quay_base_url="${QUAY_BASE_URL:-https://quay.io}"
page=1
all_tags=""

while true; do
    response=$(curl -sf \
        "${quay_base_url}/api/v1/repository/stackstate/core-stackpacks/tag/?onlyActiveTags=true&limit=100&page=${page}")
    page_tags=$(printf '%s' "$response" | jq -r '.tags[].name')
    all_tags="${all_tags}${page_tags}"$'\n'
    has_more=$(printf '%s' "$response" | jq -r '.has_additional')
    [ "$has_more" = "true" ] || break
    page=$((page + 1))
done

printf '%s' "$all_tags" |
    jq -Rrs --arg branch "$branch" '
        split("\n")
        | map(
            . as $tag
            | capture("^(?<timestamp>[0-9]{14})-(?<suffix>.+)$")?
            | select(.suffix | startswith($branch + "-"))
            | .timestamp as $timestamp
            | (.suffix | ltrimstr($branch + "-")
                | capture("^[0-9a-f]+(-retry-(?<retry>[0-9]+))?$")?)
            | {
                tag: $tag,
                timestamp: ($timestamp | tonumber),
                retry: ((.retry // "0") | tonumber)
            }
        )
        | if length == 0 then error("No core StackPack tags found for branch " + $branch)
          else max_by([.timestamp, .retry, .tag]).tag
          end
    '
