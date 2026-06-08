#!/bin/bash

set -euxo pipefail

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# shellcheck disable=SC1091
source "$dir/util.sh"  # for update_chart_version_in_readme_file

# The script "bumps" helm chart version and pushes the change back to master via the GitHub
# GraphQL `createCommitOnBranch` mutation. Commits created via the API by a GitHub App are
# auto-signed by GitHub on the App's behalf, so the resulting commit lands as Verified and
# satisfies branch protection's `require_signed_commits` rule. (The older `git commit` +
# `git push` flow via the App's HTTPS token produced unsigned commits — GPG/SSH signing was
# deliberately retired during the GitLab → GitHub port; we don't ship a key to CI.)
#
# Usage:  ./scripts/ci/bump_suse_chart_pre_master_version.sh <chart-dir>
#
# Required env:
#   GH_TOKEN     — GitHub App installation token (already set in the ci.yml after-script step)
#
# Behaviour:
#   - x.y.z-pre.N  →  x.y.z-pre.(N+1)     (increment the prerelease counter)
#   - x.y.z        →  x.y.(z+1)-pre.1     (start a new prerelease series)

chart=$1
chart_path="$chart/Chart.yaml"
readme_path="$chart/README.md"
current_version=$(yq ".version" "$chart_path")

if [[ "$current_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+-pre\.[0-9]+$ ]]; then
  current_pre_counter=$(sed -E 's/^[0-9]+\.[0-9]+\.[0-9]+-pre\.([0-9]+)$/\1/' <<< "$current_version")
  new_pre_counter=$((current_pre_counter+1))
  sed -i -E "s/^(version: [0-9]+\.[0-9]+\.[0-9]+-pre\.)[0-9]+$/\1${new_pre_counter}/" "$chart_path"
elif [[ "$current_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  current_patch_counter=$(sed -E 's/^[0-9]+\.[0-9]+\.([0-9])$/\1/' <<< "$current_version")
  new_patch_counter=$((current_patch_counter+1))
  sed -i -E "s/^(version: [0-9]+\.[0-9]+\.)[0-9]+$/\1${new_patch_counter}-pre.1/" "$chart_path"
else
  echo "Not supported chart version, you have to fix it first to run the script. Expected x.y.z or x.y.z-pre.v"
  exit 1
fi

new_version=$(yq ".version" "$chart_path")
update_chart_version_in_readme_file "$chart"

# Read the two file contents we want to commit. Base64 them per GraphQL Base64String input
# requirements. `-w0` keeps each blob on a single line (no line wrapping); fine on Linux,
# unavailable on macOS — but this script only runs in CI on Linux containers.
chart_b64=$(base64 -w0 "$chart_path")
readme_b64=$(base64 -w0 "$readme_path")

# expectedHeadOid gives optimistic-concurrency: if master moves between this read and the
# mutation, the API rejects with a clear error rather than overwriting. Re-run the job to
# retry; in practice the only thing pushing to master is this job itself, so collisions are
# very rare. Fail loudly on conflict.
head_sha=$(git rev-parse HEAD)

# Include `[skip ci]` in the headline so GitHub Actions natively skips ALL workflows on the
# resulting commit (avoids re-triggering ci.yml's build/test path on a Chart.yaml-only diff,
# and avoids re-entering push_to_internal which already has its own `[skip ci]` guard).
# shellcheck disable=SC2016  # $-variables in the query are GraphQL refs, not bash; gh api binds them via the -f flags below.
gh api graphql \
  -f query='
mutation(
  $repo: String!, $branch: String!, $message: String!, $sha: GitObjectID!,
  $chart_path: String!, $chart_b64: Base64String!,
  $readme_path: String!, $readme_b64: Base64String!
) {
  createCommitOnBranch(input: {
    branch:          { repositoryNameWithOwner: $repo, branchName: $branch }
    message:         { headline: $message }
    expectedHeadOid: $sha
    fileChanges: {
      additions: [
        { path: $chart_path,  contents: $chart_b64 }
        { path: $readme_path, contents: $readme_b64 }
      ]
    }
  }) { commit { url oid } }
}' \
  -f repo="${GITHUB_REPOSITORY:-StackVista/helm-charts-internal}" \
  -f branch="${BRANCHES:-master}" \
  -f message="Updating '$chart' helm chart version to $new_version [skip ci]" \
  -f sha="$head_sha" \
  -f chart_path="$chart_path" \
  -f chart_b64="$chart_b64" \
  -f readme_path="$readme_path" \
  -f readme_b64="$readme_b64"
