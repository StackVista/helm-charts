#!/usr/bin/env bash

set -e

quay_token=${QUAY_TOKEN:?"Please set the QUAY_TOKEN environment variable"}

export STACKSTATE_VERSION_REGEX="^(?<version>[0-9]+\\.[0-9]+\\.[0-9]+-snapshot.[0-9]+-master.*)$"
# We have two variants of the StackPacks docker image. For StackPacks 2.0, a logical version has been added to the image tag, i.e. `2_0`.
# We only match master build tags without a version suffix (e.g., -2_0-) before the variant.
# Example of matched tag:    20251028112254-master-4e2e7f4-prime-selfhosted
# Example of excluded tag:   20251028112254-master-4e2e7f4-2_0-prime-selfhosted
# This avoids picking up 2.x builds while 1.x remains the default StackPack version.
export STACKPACKS_VERSION_REGEX="^(?<version>[0-9]+-master-[0-9a-f]+)-(prime|community)-(selfhosted|saas)$"

function get_latest_master_version() {
  image=${1:?"Please provide the image name"}
  version_regex=${2:?"Please provide the version regex"}
  curl -s -H "Authorization: Bearer ${quay_token}" -X GET "https://quay.io/api/v1/repository/stackstate/${image}/tag/?limit=100" | jq -r --arg REGEX "$version_regex" '.tags[] | select(.name | test($REGEX)) | .name | capture($REGEX) | .version' | sort -V | tail -n 1
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
