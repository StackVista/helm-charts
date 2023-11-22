#!/usr/bin/env bash

set -e

quay_token=${QUAY_TOKEN:?"Please set the QUAY_TOKEN environment variable"}

export STACKSTATE_VERSION_REGEX="^[0-9]+\\.[0-9]+\\.[0-9]+-snapshot.[0-9]+-master.*$"
export STACKPACKS_VERSION_REGEX="^[0-9]+-master.*-(selfhosted|saas)$"

function get_latest_master_version() {
  image=${1:?"Please provide the image name"}
  version_regex=${2:?"Please provide the version regex"}
  curl -s -H "Authorization: Bearer ${quay_token}" -X GET "https://quay.io/api/v1/repository/stackstate/${image}/tag/?limit=100" | jq -r --arg REGEX "$version_regex" '.tags[] | select(.name | test($REGEX)) | .name' | sort -V | tail -n 1
}
