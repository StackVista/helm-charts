#!/usr/bin/env bash
# Open a PR from $1 (source branch) into $2 (target branch) with title $3.
# Expects env: GH_TOKEN (GitHub App installation token with pull-requests:write).
set -euo pipefail

source_branch="${1:?source branch required}"
target_branch="${2:?target branch required}"
title="${3:?PR title required}"

gh pr create \
  --base "$target_branch" \
  --head "$source_branch" \
  --title "$title" \
  --body "Automated PR opened by updatecli."
