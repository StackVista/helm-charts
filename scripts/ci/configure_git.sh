#!/usr/bin/env bash
set -euo pipefail

: "${GH_APP_SLUG:?GH_APP_SLUG must be set}"
: "${GH_TOKEN:?GH_TOKEN must be set}"

# Trust the workspace directory regardless of who owns it on disk. In
# `container:` jobs the workspace was created by the runner pod (fsGroup 1001)
# but we run as root (--user 0:0), so git would otherwise refuse with
# "dubious ownership in repository".
git config --global --add safe.directory '*'

git config --global user.name "${GH_APP_SLUG}[bot]"
git config --global user.email "${GH_APP_SLUG}[bot]@users.noreply.github.com"

# Authenticate git push via the App installation token. GitHub records the
# pusher as the App bot-author for any commit pushed through this remote.
git config --global "url.https://x-access-token:${GH_TOKEN}@github.com/.insteadOf" "https://github.com/"
