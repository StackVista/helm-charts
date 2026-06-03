#!/usr/bin/env bash

set -e

# Shared helpers sourced by chart bump / tag / release scripts.
#
# Expects git user identity to already be configured (see scripts/ci/configure_git.sh).
# Expects $GH_TOKEN (GitHub App installation token) when pushing.
# Expects $BRANCHES (space-separated) for push_changes / push_changes_skip_ci.

REPO_PUSH_URL="https://x-access-token:${GH_TOKEN:-}@github.com/StackVista/helm-charts-internal.git"

function commit_changes() {
  message=${1:?"Please provide a commit message"}

  if git diff --cached --exit-code; then
    echo "No changes, not committing anything"
  else
    if [[ "${PROMOTION_DRY_RUN}" == 'no' ]]; then
      echo "Committing changes"
      git commit -m "${message}"
    else
      echo "Not committing changes, set PROMOTION_DRY_RUN='no' to commit changes"
      echo "Commit message that would have been used: '${message}'"
    fi
  fi
}

function push_changes_skip_ci() {
  for branch in $BRANCHES; do
    if [[ "${PROMOTION_DRY_RUN}" == 'no' ]]; then
      echo "Pushing changes (skip-ci via [skip ci] commit trailer convention is workflow-level on GitHub)"
      git pull --rebase origin "${branch}"
      git push "${REPO_PUSH_URL}" HEAD:"${branch}"
    else
      echo "Not pushing changes, set PROMOTION_DRY_RUN='no' to commit changes"
    fi
  done
}

function push_tag_skip_ci() {
  tag=$1
  echo "Pushing tag '$tag' (skip-ci handled at the workflow level on GitHub)"
  git push "${REPO_PUSH_URL}" "${tag}"
}

function push_changes() {
  for branch in $BRANCHES; do
    if [[ "${PROMOTION_DRY_RUN}" == 'no' ]]; then
      echo "Pushing changes"
      git pull --rebase origin "${branch}"
      git push "${REPO_PUSH_URL}" HEAD:"${branch}"
    else
      echo "Not pushing changes, set PROMOTION_DRY_RUN='no' to commit changes"
    fi
  done
}

function update_chart_version_in_readme_file() {
  chart=${1:?"Please provide chart name"}
  chart_path="$chart/Chart.yaml"
  readme_path="$chart/README.md"

  current_version=$(yq -r ".version" "$chart_path")
  sed -i -E "s/^(Current chart version is ).*$/\1\`$current_version\`/" "$readme_path"
}

function get_secret_values() {
  # This function extracts credentials, etc and sets them as environment variables.
  secret_file=$1
  eval "$(sops -d "$secret_file" | awk -F ": " '{print $1" "$2}' | while read -r key value; do printf 'export %s=%q\n' "$key" "$value"; done)"
}
