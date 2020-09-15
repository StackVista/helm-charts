#!/usr/bin/env bash

set -euxo pipefail

sg_version=$(eval "${UPDATE_STACKGRAPH_VERSION}")

if git diff --cached --exit-code; then
  echo "No changes, not committing anything"
else
  git commit -m "chore(stackstate): Upgrade StackGrap to version ${sg_version}."
  git push "https://gitlab-ci-token:${gitlab_api_scope_token:?}@gitlab.com/stackvista/devops/helm-charts.git" HEAD:"${CI_COMMIT_BRANCH}"
fi
