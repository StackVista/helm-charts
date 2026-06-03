#!/usr/bin/env bash
# Squash-merge an updatecli working branch into TARGET_BRANCH as a single
# commit, then delete the working branch on origin.
#
# Runs as a follow-up to a CI job that invoked `updatecli apply` and pushed
# its per-target commits to <working-branch> (default behaviour of updatecli's
# github SCM with workingbranch=true).
#
# Usage: updatecli_squash_push.sh <working-branch> <target-branch> <commit-message>
#
# Required env: GH_TOKEN (GitHub App installation token with contents:write).
# Optional env: PUSH_CLONE_DIRECTORY (defaults to /tmp/updatecli-push).
set -euo pipefail

WORKING_BRANCH="${1:?working branch required}"
TARGET_BRANCH="${2:?target branch required}"
COMMIT_MESSAGE="${3:?commit message required}"

: "${GH_TOKEN:?GH_TOKEN must be set}"

PUSH_CLONE_DIRECTORY="${PUSH_CLONE_DIRECTORY:-/tmp/updatecli-push}"
REPO_URL="https://x-access-token:${GH_TOKEN}@github.com/StackVista/helm-charts-internal.git"

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

# Plain commit (unsigned). When pushed via the App installation token the
# commit is attributed to the App's bot-author on GitHub.
git commit -m "$COMMIT_MESSAGE"

# Plain push (no --force). If TARGET_BRANCH moved during the run we fail and
# re-run rather than silently overwriting whoever else pushed.
git push origin "HEAD:$TARGET_BRANCH"

# Clean up the working branch on origin. Non-fatal: if it fails (race with
# another run, perms, etc.) the next updatecli run will just force-push over
# it anyway.
git push origin --delete "$WORKING_BRANCH" || true
