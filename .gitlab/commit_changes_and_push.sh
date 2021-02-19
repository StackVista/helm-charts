#!/usr/bin/env bash

set -euxo pipefail

COMPONENT="${1?Please specify component}"
COMPONENT_VERSION="${2?Please specify component version}"

if git diff --cached --exit-code; then
  echo "No changes, not committing anything"
else
  git commit -m "chore(stackstate): Upgrade ${COMPONENT} to version ${COMPONENT_VERSION}."
  git push "https://gitlab-ci-token:${gitlab_api_scope_token:?}@gitlab.com/stackvista/devops/helm-charts.git" HEAD:"${CI_COMMIT_BRANCH}"
fi
