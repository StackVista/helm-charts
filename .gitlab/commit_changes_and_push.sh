#!/usr/bin/env bash

set -euo pipefail

COMPONENT="${1?Please specify component}"
COMPONENT_VERSION="${2?Please specify component version}"

# shellcheck disable=SC1091
source "$CI_PROJECT_DIR/.gitlab/gpg_utils.sh"
check_git_configuration || configure_git_user

if git diff --cached --exit-code; then
  echo "No changes, not committing anything"
else
  git commit -m "chore(stackstate): Upgrade ${COMPONENT} to version ${COMPONENT_VERSION}."
  git push "https://gitlab-ci-token:${gitlab_api_scope_token:?}@gitlab.com/stackvista/devops/helm-charts.git" HEAD:"${CI_COMMIT_BRANCH}"
fi
