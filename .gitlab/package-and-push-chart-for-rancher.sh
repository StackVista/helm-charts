#!/bin/bash

set -euo pipefail

chart="$1"
chart_repo_name="suse-observability"
build_dir="./build"

# Using the s3 URL mitigates cache issues when building multiple charts in parallel
chart_repo_url="https://s3.amazonaws.com/${RANCHER_HELM_REGISTRY_BUCKET}"
chart_repo_s3="s3://${RANCHER_HELM_REGISTRY_BUCKET}"

cd stable || exit

mkdir -p "${build_dir}"

echo "Packaging the chart ${chart} for repository ${chart_repo_name} at url ${chart_repo_url}"
helm package "${chart}" -d "${build_dir}"

echo "Downloading the index.yaml from the upstream"
if ! curl -slf "${chart_repo_url}/index.yaml" > upstream-index.yaml; then
  echo "ERROR: failed to download the index.yaml from the upstream. Please check if the upstream is alive and the index.yaml file exists"
  rm upstream-index.yaml
  exit 1
fi

echo "Updating the index.yaml"
helm repo index --merge upstream-index.yaml "${build_dir}"

echo "Uploading chart"
AWS_ACCESS_KEY_ID="${RANCHER_HELM_REGISTRY_USERNAME}" AWS_SECRET_ACCESS_KEY="${RANCHER_HELM_REGISTRY_PASSWORD}" aws s3 cp --recursive "${build_dir}" "${chart_repo_s3}"

echo "Invalidating CloudFront Distribution"
AWS_ACCESS_KEY_ID="${RANCHER_HELM_REGISTRY_USERNAME}" AWS_SECRET_ACCESS_KEY="${RANCHER_HELM_REGISTRY_PASSWORD}" aws cloudfront create-invalidation --distribution-id "${RANCHER_HELM_REGISTRY_DISTRIBUTION_ID}" --paths "/*"
