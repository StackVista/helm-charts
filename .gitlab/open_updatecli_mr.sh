#!/usr/bin/env bash
# Ensure a merge request exists for the updatecli docker-images branch.
# Called after updatecli jobs so the MR is created regardless of whether
# the finalize pipeline pushed (updatecli's action only runs when it pushes).
set -euo pipefail

SOURCE_BRANCH="${1:-updatecli-master-docker-images}"
TARGET_BRANCH="${2:-master}"
TITLE="${3:-[master] Bump helm chart docker images}"

if [ -z "${GITLAB_TOKEN:-}" ]; then
  echo "ERROR: GITLAB_TOKEN not set"
  exit 1
fi

# Project ID: use CI_PROJECT_ID in GitLab CI (numeric), or URL-encoded path
PROJECT_ID="${CI_PROJECT_ID:-stackvista%2Fdevops%2Fhelm-charts}"
API_URL="https://gitlab.com/api/v4/projects/${PROJECT_ID}/merge_requests"

# Check if source branch exists on remote
git fetch origin "$SOURCE_BRANCH" --quiet 2>/dev/null || true
if ! git rev-parse "origin/$SOURCE_BRANCH" >/dev/null 2>&1; then
  echo "Branch origin/$SOURCE_BRANCH does not exist; skipping MR create"
  exit 0
fi

# Find existing open MR for this source branch
EXISTING=$(curl -s -f --header "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  "${API_URL}?state=opened&source_branch=${SOURCE_BRANCH}&target_branch=${TARGET_BRANCH}" \
  | jq -r 'if type == "array" then .[0].iid else empty end' 2>/dev/null || echo "")

if [ -n "$EXISTING" ] && [ "$EXISTING" != "null" ]; then
  echo "Merge request !${EXISTING} already exists for ${SOURCE_BRANCH} -> ${TARGET_BRANCH}"
  exit 0
fi

# Create the MR
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
  --header "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  --header "Content-Type: application/json" \
  "${API_URL}" \
  --data "{\"source_branch\":\"${SOURCE_BRANCH}\",\"target_branch\":\"${TARGET_BRANCH}\",\"title\":\"${TITLE}\"}")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "201" ]; then
  MR_IID=$(echo "$BODY" | jq -r '.iid')
  MR_URL=$(echo "$BODY" | jq -r '.web_url')
  echo "Created merge request !${MR_IID}: ${MR_URL}"
else
  echo "ERROR: Failed to create MR (HTTP ${HTTP_CODE}): ${BODY}" >&2
  exit 1
fi
