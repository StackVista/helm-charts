#!/usr/bin/env bash

set -euo pipefail

# Verifies if a version of a chart has been "bumped" if not then it returns an error.
# The script is executed when there are some changes in a chart directory.
#
# Required env: TARGET_BRANCH (the PR base branch). In GitHub Actions, set this
# from ${{ github.base_ref }} in the workflow step.

chart="$1"
: "${TARGET_BRANCH:?TARGET_BRANCH must be set (e.g. from github.base_ref)}"

# Gate: skip if THIS BRANCH didn't modify stable/<chart>/. We use the merge-base
# (not origin/${TARGET_BRANCH}'s tip) because if master moves forward after this
# branch was cut — e.g. someone else legitimately bumps a chart on master — the
# tip comparison would report our branch as "having changes" for charts we
# never touched, then the version-bump check below would falsely demand a bump.
# Comparing to the merge-base reflects only what this branch introduced.
#
# Note: the VERSION comparison below intentionally still uses the tip — if we
# *did* modify the chart, we want to fail loudly if master has bumped past us
# in the meantime (concurrent bumps would otherwise silently downgrade on
# merge or force an awkward Chart.yaml conflict).
base=$(git merge-base "origin/${TARGET_BRANCH}" HEAD 2>/dev/null || echo "")
if [ -z "$base" ]; then
  echo "Cannot determine merge-base with origin/${TARGET_BRANCH}; running version check against tip"
elif git diff --quiet "$base" HEAD -- "stable/${chart}/"; then
  echo "stable/${chart}/ has no changes on this branch (since merge-base ${base}); skipping version-bump check"
  exit 0
fi

check_if_helm_chart_exists=$(git ls-tree -r "origin/${TARGET_BRANCH}" -- "stable/${chart}/Chart.yaml")
if [ -z "$check_if_helm_chart_exists" ]; then
  echo "Chart.yaml doesnt exist in the target branch, probably it is the first commit for the chart. Skipping version check!!!"
  exit 0
fi

remote_version=$(git show "origin/${TARGET_BRANCH}:stable/${chart}/Chart.yaml" | yq ".version")
new_version=$(yq ".version" "stable/${chart}/Chart.yaml")

if [ "$(printf '%s\n' "$remote_version" "$new_version" | sort -V | head -n1)" = "$new_version" ]; then
  echo "Version of the chart should be updated"
  exit 1
fi
