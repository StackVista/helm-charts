#!/usr/bin/env bash
# Open a PR from $1 (source branch) into $2 (target branch) with title $3.
#
# Idempotent: if an open PR for the same head→base pair already exists, log its
# URL and exit 0 instead of erroring. Matches the legacy GitLab equivalent
# (.gitlab/open_updatecli_mr.sh) behaviour — the `gh pr create` predecessor of
# this script lacked it.
#
# Uses the REST API directly via curl + jq because the stackstate-devops CI
# container image does not ship the `gh` binary.
#
# Expects env:
#   GH_TOKEN           — GitHub App installation token with pull-requests:write
#   GITHUB_REPOSITORY  — owner/repo (set by GH Actions); falls back to this repo
set -euo pipefail

source_branch="${1:?source branch required}"
target_branch="${2:?target branch required}"
title="${3:?PR title required}"

repo="${GITHUB_REPOSITORY:-StackVista/helm-charts-internal}"
owner="${repo%%/*}"
api="https://api.github.com/repos/${repo}/pulls"

gh_headers=(
  -H "Authorization: Bearer ${GH_TOKEN}"
  -H "Accept: application/vnd.github+json"
  -H "X-GitHub-Api-Version: 2022-11-28"
)

# Idempotency check. `head` must be qualified as `owner:branch` for cross-repo
# safety; for same-repo PRs the qualified form still works.
existing=$(curl -sS -G "${gh_headers[@]}" \
  --data-urlencode "state=open" \
  --data-urlencode "head=${owner}:${source_branch}" \
  --data-urlencode "base=${target_branch}" \
  "$api" \
  | jq -r 'if length > 0 then .[0].html_url else empty end')

if [ -n "$existing" ]; then
  echo "PR already exists for ${source_branch} -> ${target_branch}: ${existing}"
  exit 0
fi

payload=$(jq -n \
  --arg title  "$title" \
  --arg head   "$source_branch" \
  --arg base   "$target_branch" \
  --arg body   "Automated PR opened by updatecli." \
  '{title: $title, head: $head, base: $base, body: $body}')

response=$(curl -sS -w "\n%{http_code}" -X POST "${gh_headers[@]}" \
  -H "Content-Type: application/json" \
  --data "$payload" \
  "$api")

http_code=$(printf '%s\n' "$response" | tail -n1)
body=$(printf '%s\n' "$response" | sed '$d')

if [ "$http_code" = "201" ]; then
  url=$(echo "$body" | jq -r '.html_url')
  echo "Created PR: ${url}"
else
  echo "ERROR: failed to create PR (HTTP ${http_code}): ${body}" >&2
  exit 1
fi
