#!/usr/bin/env bash

set -e

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

function push_changes() {
  if [[ "${PROMOTION_DRY_RUN}" == 'no' ]]; then
    echo "Pushing changes"
    git pull --rebase origin "${CI_COMMIT_BRANCH}"
    git push "https://gitlab-ci-token:${gitlab_api_scope_token:?}@gitlab.com/stackvista/devops/saas-tenants.git" HEAD:"${CI_COMMIT_BRANCH}"
  else
    echo "Not pushing changes, set PROMOTION_DRY_RUN='no' to commit changes"
  fi
}
