#!/usr/bin/env sh

set -o errexit
set -o nounset

readonly awsBucket="${AWS_BUCKET:-"s3://helm.stackstate.io"}"
readonly repoUrl="${REPO_URL:-"https://helm.stackstate.io/"}"

main() {
  if ! sync_repo stable "${awsBucket}" "${repoUrl}"; then
    log_error "Not all stable charts could be packaged and synced!"
  fi
}

sync_repo() {
  repo_dir="${1?Specify repo dir}"
  bucket="${2?Specify repo bucket}"
  repo_url="${3?Specify repo url}"
  sync_dir="${repo_dir}-sync"
  index_dir="${repo_dir}-index"

  echo "Syncing repo '${repo_dir}'..."

  mkdir -p "${sync_dir}"
  if ! aws s3 cp "${bucket}/index.yaml" "${index_dir}/index.yaml"; then
    log_error "Exiting because unable to copy index locally. Not safe to proceed."
    exit 1
  fi

  exit_code=0

  for dir in "${repo_dir}"/*; do
    if helm dependency build "${dir}"; then
      helm package --destination "${sync_dir}" "${dir}"
    else
      log_error "Problem building dependencies. Skipping packaging of '${dir}'."
      exit_code=1
    fi
  done

  if helm repo index --url "${repo_url}" --merge "${index_dir}/index.yaml" "${sync_dir}"; then
    # Move updated index.yaml to sync folder so we don't push the old one again
    mv -f "${sync_dir}/index.yaml" "${index_dir}/index.yaml"

    aws s3 sync "${sync_dir}" "${bucket}"

    # Make sure index.yaml is synced last
    aws s3 cp "${index_dir}/index.yaml" "${bucket}"
  else
    log_error "Exiting because unable to update index. Not safe to push update."
    exit 1
  fi

  ls -l "${sync_dir}"

  return "${exit_code}"
}

log_error() {
  printf '\e[31mERROR: %s\n\e[39m' "$1" >&2
}

main
