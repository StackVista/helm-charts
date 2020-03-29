#!/usr/bin/env sh

[ -n "${TRACE+x}" ] && set -x

set -e

installDependencies() {
  apk -Uuv add bash curl groff less openssl python py-pip
  curl -fSL https://git.io/get_helm.sh | bash -s -- --version "${HELM_VERSION}"
  pip install awscli
}

installDependencies
