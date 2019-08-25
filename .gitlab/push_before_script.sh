#!/usr/bin/env sh

[ -n "${TRACE+x}" ] && set -x

installDependencies() {
  apk -Uuv add bash curl groff less openssl python py-pip
  curl -fSL https://git.io/get_helm.sh | bash -
  pip install awscli
}

configureHelm() {
  helm init --client-only
}

installDependencies
configureHelm
