#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -x

readonly AWS_BUCKET_STABLE="s3://helm.stackstate.io"
readonly HELM_TARBALL="helm-v2.14.2-linux-amd64.tar.gz"
readonly HELM_URL="https://storage.googleapis.com/kubernetes-helm"
readonly STABLE_REPO_URL="https://helm.stackstate.io/"

main() {
  setup_helm_client

  if ! sync_repo stable "$AWS_BUCKET_STABLE" "$STABLE_REPO_URL"; then
    log_error "Not all stable charts could be packaged and synced!"
  fi
}

setup_helm_client() {
  echo "Setting up Helm client..."

  curl --user-agent curl-ci-sync -sSL -o "$HELM_TARBALL" "$HELM_URL/$HELM_TARBALL"
  tar xzfv "$HELM_TARBALL"

  PATH="$(pwd)/linux-amd64/:$PATH"

  helm init --client-only
}

setup_helm_repositories() {
  echo "Setting up Helm repositories..."

  helm repo add bitnami https://charts.bitnami.com/
  helm repo add gitlab https://charts.gitlab.io/
  helm repo add incubator https://kubernetes-charts-incubator.storage.googleapis.com/
  helm repo add jetstack https://charts.jetstack.io/
  helm repo add jfrog https://charts.jfrog.io/

  echo "Updating Helm repository indexes..."
  helm repo update
}

sync_repo() {
  local repo_dir="${1?Specify repo dir}"
  local bucket="${2?Specify repo bucket}"
  local repo_url="${3?Specify repo url}"
  local sync_dir="${repo_dir}-sync"
  local index_dir="${repo_dir}-index"

  echo "Syncing repo '$repo_dir'..."

  mkdir -p "$sync_dir"
  if ! aws s3 cp "$bucket/index.yaml" "$index_dir/index.yaml"; then
    log_error "Exiting because unable to copy index locally. Not safe to proceed."
    exit 1
  fi

  local exit_code=0

  for dir in "$repo_dir"/*; do
    if helm dependency build "$dir"; then
      helm package --destination "$sync_dir" "$dir"
    else
      log_error "Problem building dependencies. Skipping packaging of '$dir'."
      exit_code=1
    fi
  done

  if helm repo index --url "$repo_url" --merge "$index_dir/index.yaml" "$sync_dir"; then
    # Move updated index.yaml to sync folder so we don't push the old one again
    mv -f "$sync_dir/index.yaml" "$index_dir/index.yaml"

    aws s3 sync "$sync_dir" "$bucket"

    # Make sure index.yaml is synced last
    aws s3 cp "$index_dir/index.yaml" "$bucket"
  else
    log_error "Exiting because unable to update index. Not safe to push update."
    exit 1
  fi

  ls -l "$sync_dir"

  return "$exit_code"
}

log_error() {
  printf '\e[31mERROR: %s\n\e[39m' "$1" >&2
}

main
