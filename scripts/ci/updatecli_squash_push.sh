#!/usr/bin/env bash
# Squash-merge an updatecli working branch into a single GitHub-signed commit.
# By default the commit lands on TARGET_BRANCH and the working branch is deleted.
# When OUTPUT_BRANCH is provided, the commit lands on OUTPUT_BRANCH instead;
# this is used for review PRs that still need signed commits.
#
# Runs as a follow-up to a CI job that invoked `updatecli apply` and pushed
# its per-target commits to <working-branch> (default behaviour of updatecli's
# github SCM with workingbranch=true).
#
# The squashed change is committed via GitHub's GraphQL `createCommitOnBranch`
# mutation rather than `git commit` + `git push`. Commits created via the API
# by a GitHub App are auto-signed by GitHub on the App's behalf, so the
# resulting commit lands as Verified and satisfies branch protection's
# `require_signed_commits` rule. (The older `git commit` + `git push` flow via
# the App's HTTPS token produced unsigned commits — GPG/SSH signing was
# deliberately retired during the GitLab → GitHub port; we don't ship a key to
# CI.) We still clone the target branch and run `git merge --squash` locally
# because we need git to compute the squashed file set; the resulting staged
# diff is then translated into GraphQL fileChanges (additions + deletions).
#
# Usage: updatecli_squash_push.sh <working-branch> <target-branch> <commit-message> [output-branch]
#
# Required env: GH_TOKEN (GitHub App installation token with contents:write).
# Optional env: PUSH_CLONE_DIRECTORY (defaults to /tmp/updatecli-push),
#               GITHUB_REPOSITORY  (defaults to StackVista/helm-charts-internal).
set -euo pipefail

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# shellcheck disable=SC1091
source "$dir/util.sh"  # for create_commit_on_branch_with_retry

WORKING_BRANCH="${1:?working branch required}"
TARGET_BRANCH="${2:?target branch required}"
COMMIT_MESSAGE="${3:?commit message required}"
OUTPUT_BRANCH="${4:-$TARGET_BRANCH}"

: "${GH_TOKEN:?GH_TOKEN must be set}"

PUSH_CLONE_DIRECTORY="${PUSH_CLONE_DIRECTORY:-/tmp/updatecli-push}"
REPO="${GITHUB_REPOSITORY:-StackVista/helm-charts-internal}"
REPO_URL="https://x-access-token:${GH_TOKEN}@github.com/${REPO}.git"

rm -rf "$PUSH_CLONE_DIRECTORY"
git clone --branch "$TARGET_BRANCH" --single-branch "$REPO_URL" "$PUSH_CLONE_DIRECTORY"
cd "$PUSH_CLONE_DIRECTORY"

if ! git fetch origin "$WORKING_BRANCH" 2>/dev/null; then
  echo "Working branch '$WORKING_BRANCH' does not exist on origin; updatecli made no changes — nothing to push"
  exit 0
fi

# Uses curl + jq instead of `gh api graphql` because the stackstate-devops CI
# container does not ship `gh`. $-variables in the query string are GraphQL
# refs (bound via the variables block below), not bash vars.
# shellcheck disable=SC2016
graphql_query='
mutation(
  $repo: String!, $branch: String!, $message: String!, $sha: GitObjectID!,
  $additions: [FileAddition!], $deletions: [FileDeletion!]
) {
  createCommitOnBranch(input: {
    branch:          { repositoryNameWithOwner: $repo, branchName: $branch }
    message:         { headline: $message }
    expectedHeadOid: $sha
    fileChanges:     { additions: $additions, deletions: $deletions }
  }) { commit { url oid } }
}'

# Build the payload by squashing the working branch onto the live target tip
# (re-run per attempt; expectedHeadOid = that tip). `must` aborts on failure —
# critically, a `git merge --squash` conflict must not build from a conflicted index.
build_squash_payload() {
  local payload_file=$1

  must git fetch --quiet origin "$TARGET_BRANCH"
  must git reset --hard --quiet FETCH_HEAD
  local head_sha
  head_sha=$(git rev-parse HEAD)
  must git fetch --quiet origin "$WORKING_BRANCH"
  must git merge --squash FETCH_HEAD

  if git diff --cached --quiet; then
    # The working-branch change is already present on the target (e.g. a
    # concurrent run landed it). Nothing left to commit — clean up and finish.
    echo "Working branch has no effective changes vs ${TARGET_BRANCH}; cleaning up" >&2
    git push origin --delete "$WORKING_BRANCH" || true
    exit 0
  fi

  if [ "$OUTPUT_BRANCH" != "$TARGET_BRANCH" ]; then
    # Seed the review branch at the target head before adding the signed commit;
    # the unsigned updatecli commits must not remain in PR history under
    # require_signed_commits.
    must git push origin "HEAD:refs/heads/${OUTPUT_BRANCH}" --force
  fi

  # Translate the staged diff into GraphQL fileChanges. `-z` keeps paths with
  # spaces/newlines intact; rename/copy records carry two paths, others one.
  # Each entry is written to its own file under $tmp and slurped via `jq -s` /
  # `--slurpfile` so the base64 content never hits argv (ARG_MAX on large READMEs).
  local tmp add_count del_count
  tmp=$(mktemp -d)
  mkdir -p "$tmp/add" "$tmp/del"
  add_count=0
  del_count=0

  emit_add() {
    jq -n --arg p "$1" --rawfile c "$1" \
          '{path: $p, contents: ($c | @base64)}' \
          > "$tmp/add/$(printf '%06d' "$add_count").json"
    add_count=$((add_count + 1))
  }
  emit_del() {
    jq -n --arg p "$1" '{path: $p}' \
          > "$tmp/del/$(printf '%06d' "$del_count").json"
    del_count=$((del_count + 1))
  }

  local status path1 path2
  while IFS= read -r -d '' status; do
    IFS= read -r -d '' path1
    case "$status" in
      R*|C*) IFS= read -r -d '' path2 ;;
      *)     path2='' ;;
    esac
    case "$status" in
      A|M|T) emit_add "$path1" ;;
      D)     emit_del "$path1" ;;
      R*)    emit_add "$path2"; emit_del "$path1" ;;
      C*)    emit_add "$path2" ;;
      *)
        echo "ERROR: unsupported git diff status '$status' for path '$path1'" >&2
        rm -rf "$tmp"
        exit 1
        ;;
    esac
  done < <(git diff --cached --name-status -z)

  # `jq -s` (slurp) reads all argument files and emits a single array containing
  # each file's parsed JSON value. With no input files jq emits nothing, so
  # branch on the counter and fall back to an empty array literal.
  if [ "$add_count" -gt 0 ]; then
    jq -s '.' "$tmp"/add/*.json > "$tmp/additions.json"
  else
    echo '[]' > "$tmp/additions.json"
  fi
  if [ "$del_count" -gt 0 ]; then
    jq -s '.' "$tmp"/del/*.json > "$tmp/deletions.json"
  else
    echo '[]' > "$tmp/deletions.json"
  fi

  # additions.json/deletions.json each hold one array, so `--slurpfile` wraps it
  # one level extra — peel with `$additions[0]`.
  jq -n \
    --arg       query     "$graphql_query" \
    --arg       repo      "$REPO" \
    --arg       branch    "$OUTPUT_BRANCH" \
    --arg       message   "$COMMIT_MESSAGE" \
    --arg       sha       "$head_sha" \
    --slurpfile additions "$tmp/additions.json" \
    --slurpfile deletions "$tmp/deletions.json" \
    '{query: $query, variables: {repo: $repo, branch: $branch, message: $message,
      sha: $sha, additions: $additions[0], deletions: $deletions[0]}}' \
    > "$payload_file"

  rm -rf "$tmp"
}

commit_url=$(create_commit_on_branch_with_retry build_squash_payload) || exit 1

# Empty commit_url means build_squash_payload took its no-op `exit 0` path (which,
# in this subshell, only ended the subshell). It already cleaned up — just stop.
if [ -z "$commit_url" ]; then
  echo "No commit created: working branch had no effective changes vs ${TARGET_BRANCH}."
  exit 0
fi
echo "Created commit: ${commit_url}"

# Clean up the raw working branch on origin when it is no longer the output
# branch. Non-fatal: if it fails (race with another run, perms, etc.) the next
# updatecli run will just force-push over it anyway. Branch deletion is not a
# commit, so plain `git push --delete` over the App token works fine even with
# `require_signed_commits` enabled.
if [ "$OUTPUT_BRANCH" != "$WORKING_BRANCH" ]; then
  git push origin --delete "$WORKING_BRANCH" || true
fi
