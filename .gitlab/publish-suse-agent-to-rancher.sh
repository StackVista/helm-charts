#!/bin/bash

set -euo pipefail

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# shellcheck disable=SC1091
source "$dir/util.sh"

build_root=$(pwd)

GREEN='\033[0;32m'
RED='\033[0;31m'
NO_COLOR='\033[0m'

cd stable/suse-observability-agent || exit

echo "Pushing container images to Rancher container registry"
helm dependencies update .

image_list_file="o11y-agent-images.txt"

./installation/o11y-agent-get-images.sh -d . > "${image_list_file}"

get_secret_values "${build_root}/sops.rancher-helm-credentials.yaml"

# Pull and push the images from the list
while IFS= read -r image; do
  # Simple check if the image is in the format <registry>/<namespace>/<repository>:<tag>
  if [[ $image =~ ^([^/]+)/([^/]+)/(.*):([^:]+)$ ]]; then
    repository_and_tag=$(echo "${image}" | cut -d'/' -f3-)
    dest_image="$RANCHER_CONTAINER_REGISTRY_URL/$RANCHER_CONTAINER_REGISTRY_NAMESPACE/${repository_and_tag}"
    echo "Copying docker://${image} to docker://${dest_image}"
    skopeo copy --all --dest-username "$RANCHER_CONTAINER_REGISTRY_USERNAME" --dest-password "$RANCHER_CONTAINER_REGISTRY_PASSWORD" "docker://${image}" "docker://${dest_image}"
    # shellcheck disable=SC2181
    if [ $? -eq 0 ]; then
      echo -e "${GREEN}Successfully copied ${dest_image}${NO_COLOR}"
    else
      echo -e "${RED}Failed to copy ${dest_image}${NO_COLOR}"
    fi
  else
      echo -e "${RED}Image url ${image} is not valid. Skipping...${NO_COLOR}"
  fi
done < "${image_list_file}"

./maintenance/change-image-source.sh -g "$RANCHER_CONTAINER_REGISTRY_URL" -p "$RANCHER_CONTAINER_REGISTRY_NAMESPACE"

cd "${build_root}" || exit

.gitlab/package-and-push-chart-for-rancher.sh suse-observability-agent
