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

# Run a command, aborting on failure. Needed inside the build-payload functions:
# they run in a command substitution where `set -e` is ignored (and can't be
# re-enabled), so e.g. a `git merge --squash` conflict wouldn't otherwise abort.
function must() {
  "$@" || { echo "FATAL: command failed: $*" >&2; exit 1; }
}

# POST a createCommitOnBranch payload (JSON file) to GitHub; echo the commit URL
# on success. Returns 0 committed / 2 stale-head conflict (retry) / 1 fatal.
# GraphQL errors come back under HTTP 200, so check both status and `.errors`.
function _gh_create_commit_post() {
  local payload_file=${1:?payload file required}
  : "${GH_TOKEN:?GH_TOKEN must be set}"

  local response http_code body
  response=$(curl -sS -w "\n%{http_code}" -X POST \
    -H "Authorization: Bearer ${GH_TOKEN}" \
    -H "Accept: application/vnd.github+json" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    -H "Content-Type: application/json" \
    --data-binary "@${payload_file}" \
    https://api.github.com/graphql)
  http_code=$(printf '%s\n' "$response" | tail -n1)
  body=$(printf '%s\n' "$response" | sed '$d')

  if [ "$http_code" != "200" ]; then
    echo "ERROR: GraphQL HTTP ${http_code}: ${body}" >&2
    return 1
  fi

  if [ "$(echo "$body" | jq 'has("errors") and (.errors | length > 0)')" = "true" ]; then
    # Retry only when EVERY error is a stale-head conflict (type STALE_DATA, or
    # the canonical "...point to <oid> but it did not" wording — matched narrowly
    # so a validation error isn't retried). `all(…)` is true on empty, hence the
    # length guard.
    if echo "$body" | jq -e '
        (.errors | length > 0) and all(.errors[];
          (((.type // "") | ascii_upcase) == "STALE_DATA")
          or (((.message // "")) | test("point to .*but it did not|is at .*but expected"; "i")))
        ' >/dev/null; then
      echo "Conflict: target branch head moved since expectedHeadOid was read." >&2
      return 2
    fi
    echo "ERROR: GraphQL returned errors:" >&2
    echo "$body" | jq '.errors' >&2
    return 1
  fi

  # A success must carry a commit URL; a 200 with none (non-JSON body, null data)
  # is a failure, not an empty success.
  local url
  url=$(printf '%s' "$body" | jq -r '.data.createCommitOnBranch.commit.url // empty' 2>/dev/null || true)
  if [ -z "$url" ]; then
    echo "ERROR: unexpected GraphQL response (no commit URL): ${body}" >&2
    return 1
  fi
  printf '%s\n' "$url"
  return 0
}

# Drive _gh_create_commit_post with stale-head retries; echo the commit URL.
#   create_commit_on_branch_with_retry <build_payload_fn>
# <build_payload_fn> takes a path to write the payload to, and MUST re-derive the
# change + expectedHeadOid from the live branch tip on each call — that's what
# makes a retry re-apply onto whatever just landed instead of replaying a stale
# payload. Env: COMMIT_RETRY_MAX_ATTEMPTS (default 5).
function create_commit_on_branch_with_retry() {
  local build_fn=${1:?build-payload function name required}
  local max_attempts=${COMMIT_RETRY_MAX_ATTEMPTS:-5}
  local payload attempt=1 rc backoff
  payload=$(mktemp)
  # shellcheck disable=SC2064
  trap "rm -f '${payload}'" RETURN

  while : ; do
    # Build to stderr: the caller captures our stdout, so only the URL that
    # _gh_create_commit_post echoes may go there (git plumbing prints to stdout).
    "$build_fn" "$payload" >&2
    # `|| rc=$?` captures the real status (a plain `if` would mask it) and keeps
    # `set -e` from firing.
    rc=0
    _gh_create_commit_post "$payload" || rc=$?
    if [ "$rc" -eq 0 ]; then
      return 0
    fi
    if [ "$rc" -eq 2 ] && [ "$attempt" -lt "$max_attempts" ]; then
      backoff=$(( attempt < 5 ? attempt : 5 ))
      echo "Re-deriving on the new branch head and retrying in ${backoff}s (attempt $((attempt + 1))/${max_attempts})..." >&2
      sleep "$backoff"
      attempt=$((attempt + 1))
      continue
    fi
    return "$rc"
  done
}
