#!/usr/bin/env sh

set -o errexit
set -o nounset

readonly repo_dir="stable"
readonly aws_bucket="${AWS_BUCKET:-"s3://helm.stackstate.io"}"
readonly repo_url="${REPO_URL:-"https://helm.stackstate.io/"}"
readonly sync_dir="${repo_dir}-sync"
readonly index_dir="${repo_dir}-index"

# Copy builds from all charts directory to `stable-sync` dir
copy_builds () {
  mkdir -p "${sync_dir}"

  for dir in "${repo_dir}"/*; do
    if [ -d "${dir}/build" ]; then
      echo "Preparing chart ${dir} ..."
      cp -a "${dir}"/build/* "${sync_dir}"
    fi
  done
}

# Updates index.yaml and then copies builds and the newly updated index.yaml to the S3
sync_charts() {
  if ! aws s3 cp "${aws_bucket}/index.yaml" "${index_dir}/index.yaml"; then
    log_error "Exiting because unable to copy index locally. Not safe to proceed."
    exit 1
  fi

  if helm repo index --url "${repo_url}" --merge "${index_dir}/index.yaml" "${sync_dir}"; then
    #Move updated index.yaml to sync folder so we don't push the old one again
    mv -f "${sync_dir}/index.yaml" "${index_dir}/index.yaml"

    aws s3 sync "${sync_dir}" "${aws_bucket}"

    # Make sure index.yaml is synced last
    aws s3 cp "${index_dir}/index.yaml" "${aws_bucket}"
  else
    log_error "Exiting because unable to update index. Not safe to push update."
    exit 1
  fi
}

log_error() {
  printf '\e[31mERROR: %s\n\e[39m' "$1" >&2
}

copy_builds
sync_charts
