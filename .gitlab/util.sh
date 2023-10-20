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
  # tags don't have CI_COMMIT_BRANCH, so I fetches the current branch(s) for current HEAD (HEAD points to a detached commit)
  # but there may be multiple branches so I iterate all of them and push a commit to each branch
  branches=${CI_COMMIT_BRANCH:-$(git for-each-ref --format='%(objectname) %(refname:short)' refs/remotes/origin | awk -v branch="$(git rev-parse HEAD)" '$1==branch && $2!="origin" {print $2}' | sed -E 's/^origin\/(.*)$/\1/')}

  for branch in $branches; do
    if [[ "${PROMOTION_DRY_RUN}" == 'no' ]]; then
      echo "Pushing changes"
      git pull --rebase origin "${branch}"
      git push "https://gitlab-ci-token:${gitlab_api_scope_token:?}@gitlab.com/stackvista/devops/helm-charts.git" HEAD:"${branch}"
    else
      echo "Not pushing changes, set PROMOTION_DRY_RUN='no' to commit changes"
    fi
  done
}
