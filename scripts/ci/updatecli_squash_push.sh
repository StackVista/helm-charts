#!/usr/bin/env bash
# Squash-merge an updatecli working branch into TARGET_BRANCH as a single
# commit, then delete the working branch on origin.
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
# Usage: updatecli_squash_push.sh <working-branch> <target-branch> <commit-message>
#
# Required env: GH_TOKEN (GitHub App installation token with contents:write).
# Optional env: PUSH_CLONE_DIRECTORY (defaults to /tmp/updatecli-push),
#               GITHUB_REPOSITORY  (defaults to StackVista/helm-charts-internal).
set -euo pipefail

WORKING_BRANCH="${1:?working branch required}"
TARGET_BRANCH="${2:?target branch required}"
COMMIT_MESSAGE="${3:?commit message required}"

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

# Stage every change from the working branch as if it were a single change on
# top of TARGET_BRANCH.
git merge --squash FETCH_HEAD

if git diff --cached --quiet; then
  echo "Working branch had no effective changes vs $TARGET_BRANCH; cleaning up"
  git push origin --delete "$WORKING_BRANCH" || true
  exit 0
fi

# expectedHeadOid gives optimistic-concurrency: if TARGET_BRANCH moves between
# this read and the mutation, the API rejects with a clear error rather than
# overwriting. Re-run the job to retry.
head_sha=$(git rev-parse HEAD)

# Translate the staged diff into GraphQL fileChanges. `-z` makes name-status
# records NUL-terminated and per-field NUL-separated, so paths with spaces or
# newlines survive intact. Rename/copy records carry TWO paths
# (oldpath NUL newpath NUL); plain A/M/D records carry one (path NUL).
#
# Each per-file addition/deletion JSON object is written to its own file under
# $tmp, then assembled into arrays via `jq -s` and consumed below via
# `jq --slurpfile`. Carrying the JSON in shell variables (and passing it via
# `--argjson`) hits ARG_MAX as soon as one file's base64 content + earlier
# accumulated entries exceed ~128KB — the auto-generated README.md alone is
# enough. `--rawfile` and `--slurpfile` read from disk, so jq never sees the
# bulk content via argv.
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
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

# `--slurpfile` binds the variable to the array of JSON values read from the
# file; additions.json / deletions.json each contain ONE value (an array), so
# the result is wrapped one level extra — peel it back with `$additions[0]`.
# Writing the assembled payload to disk (rather than capturing it in a shell
# variable) keeps the base64 content off the next command line.
jq -n \
  --arg       query     "$graphql_query" \
  --arg       repo      "$REPO" \
  --arg       branch    "$TARGET_BRANCH" \
  --arg       message   "$COMMIT_MESSAGE" \
  --arg       sha       "$head_sha" \
  --slurpfile additions "$tmp/additions.json" \
  --slurpfile deletions "$tmp/deletions.json" \
  '{query: $query, variables: {repo: $repo, branch: $branch, message: $message,
    sha: $sha, additions: $additions[0], deletions: $deletions[0]}}' \
  > "$tmp/payload.json"

# Feed the payload to curl from disk (`--data-binary @path`) — never via argv
# or a shell variable expansion that would blow past ARG_MAX on large diffs.
response=$(curl -sS -w "\n%{http_code}" -X POST \
  -H "Authorization: Bearer ${GH_TOKEN}" \
  -H "Accept: application/vnd.github+json" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  -H "Content-Type: application/json" \
  --data-binary "@$tmp/payload.json" \
  https://api.github.com/graphql)

http_code=$(printf '%s\n' "$response" | tail -n1)
body=$(printf '%s\n' "$response" | sed '$d')

# GitHub's GraphQL endpoint returns HTTP 200 even when the mutation errors at
# the GraphQL layer (e.g. expectedHeadOid mismatch, missing permissions).
# Check both transport status AND `.errors` in the body.
if [ "$http_code" != "200" ]; then
  echo "ERROR: GraphQL HTTP ${http_code}: ${body}" >&2
  exit 1
fi

if [ "$(echo "$body" | jq 'has("errors") and (.errors | length > 0)')" = "true" ]; then
  echo "ERROR: GraphQL returned errors:" >&2
  echo "$body" | jq '.errors' >&2
  exit 1
fi

commit_url=$(echo "$body" | jq -r '.data.createCommitOnBranch.commit.url')
echo "Created commit: ${commit_url}"

# Clean up the working branch on origin. Non-fatal: if it fails (race with
# another run, perms, etc.) the next updatecli run will just force-push over
# it anyway. Branch deletion isn't signed, so plain `git push --delete` over
# the App token works fine even with `require_signed_commits` enabled.
git push origin --delete "$WORKING_BRANCH" || true
