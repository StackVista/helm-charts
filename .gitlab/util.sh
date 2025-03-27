#!/usr/bin/env bash

set -e

# shellcheck disable=SC1091
source "$CI_PROJECT_DIR/.gitlab/gpg_utils.sh"

function commit_changes() {
  message=${1:?"Please provide a commit message"}

  if git diff --cached --exit-code; then
    echo "No changes, not committing anything"
  else
    if [[ "${PROMOTION_DRY_RUN}" == 'no' ]]; then
      echo "Committing changes"
      check_git_configuration || configure_git_user
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
      echo "Pushing changes"
      git pull --rebase origin "${branch}"
      git push "https://gitlab-ci-token:${gitlab_api_scope_token:?}@gitlab.com/stackvista/devops/helm-charts.git" HEAD:"${branch}" -o ci.skip # I have to use option `ci.skip' instead of a commit message otherwise the pipeline isn't trigger for a tag creation
    else
      echo "Not pushing changes, set PROMOTION_DRY_RUN='no' to commit changes"
    fi
  done
}

function push_changes() {
  for branch in $BRANCHES; do
    if [[ "${PROMOTION_DRY_RUN}" == 'no' ]]; then
      echo "Pushing changes"
      git pull --rebase origin "${branch}"
      git push "https://gitlab-ci-token:${gitlab_api_scope_token:?}@gitlab.com/stackvista/devops/helm-charts.git" HEAD:"${branch}"
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
  eval "$(sops -d "$secret_file" | awk -F ": " '{print $1" "$2}' | while read -r key value; do echo export "${key}"="$value"; done)"
}
