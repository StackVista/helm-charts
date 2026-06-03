#!/usr/bin/env sh

[ -n "${TRACE+x}" ] && set -x

set -e

installDependencies() {
  apk -Uuv add bash curl groff less openssl yq
  curl -fSL "https://github.com/yannh/kubeconform/releases/download/v0.4.12/kubeconform-linux-amd64.tar.gz" | tar -C /usr/local/bin -xvz
  chmod +x /usr/local/bin/kubeconform
}

# Trust the workspace directory regardless of who owns it on disk. In
# `container:` jobs the workspace was created by the runner pod (fsGroup 1001)
# but we run as root (--user 0:0), so git would otherwise refuse with
# "dubious ownership in repository". `verify_versions_bumped.sh` (called by
# check_chart_version) runs git diff against origin/${TARGET_BRANCH}.
git config --global --add safe.directory '*'

installDependencies
