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

# Updates a helm chart value to new new one and also updates README file.
function updateChartValue() {
  value_path=${1:?"Please provide path to value"}
  new_value=${2:?"Please provide a new value"}
  values_path=${3:?"Please provide a path to values.yaml file"}
  readme_path=${4:?"Please provide a path to README.md file"}

  # checks if value exists in the README file
  set +e
  grep -q "$value_path" "$readme_path"
  value_exists_readme=$?
  set -e
  if ! [[ "$value_exists_readme" -eq "0" ]]; then
    echo "not found $value_path in the $readme_path file"
    return 1
  fi

  # checks if value exists in the values.yaml file
  if [[ "$(yq ".$value_path" "$values_path")" = "null" ]]; then
    echo "not found $value_path in the $values_path file"
    return 1
  fi

  yq -i eval ".${value_path}=\"${new_value}\"" "$values_path"
  # Replaces value in the README file to the new value, e.g.
  # | stackstate.stackpacks.image.tag | string | `"20231129143410-master-630ae63-selfhosted"` | Tag used for the `stackpacks` Docker image; |
  #                                                ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ - will be replaced to the new value
  sed -i -E "s/^(\| ${value_path} \| string \| \`\").*(\"\` \|.*)$/\1${new_value}\2/" "$readme_path"
}
