#!/bin/bash

# DRY_RUN=true (default) skips the destructive calls (skopeo copy, chart-mutation,
# package-and-push) so the rest of the script — helm deps update, sops decrypt,
# o11y-agent image-list generation — can still run end-to-end. Set DRY_RUN=false to
# actually publish to the Rancher container registry + S3.

release=${1:-"prerelease"}

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
    if [[ "${DRY_RUN:-true}" == "false" ]]; then
      skopeo copy --all --dest-username "$RANCHER_CONTAINER_REGISTRY_USERNAME" --dest-password "$RANCHER_CONTAINER_REGISTRY_PASSWORD" "docker://${image}" "docker://${dest_image}"
      # shellcheck disable=SC2181
      if [ $? -eq 0 ]; then
        echo -e "${GREEN}Successfully copied ${dest_image}${NO_COLOR}"
      else
        echo -e "${RED}Failed to copy ${dest_image}${NO_COLOR}"
      fi
    else
      echo "[DRY_RUN] skipping skopeo copy → ${dest_image}"
    fi
  else
      echo -e "${RED}Image url ${image} is not valid. Skipping...${NO_COLOR}"
  fi
done < "${image_list_file}"

./maintenance/change-image-source.sh -g "$RANCHER_CONTAINER_REGISTRY_URL" -p "$RANCHER_CONTAINER_REGISTRY_NAMESPACE"

cd "${build_root}" || exit

if [[ "$release" = "release" ]]; then
  echo "Making a public release"
else
  echo "Making a prerelease, adding -pre to the chart."
  # Mutates Chart.yaml; only meaningful followed by the package-and-push, so gated together.
  if [[ "${DRY_RUN:-true}" == "false" ]]; then
    scripts/ci/modify_chart_to_prerelease_version.sh stable/suse-observability-agent
  else
    echo "[DRY_RUN] skipping modify_chart_to_prerelease_version.sh (would mutate Chart.yaml)"
  fi
fi

scripts/ci/package-and-push-chart-for-rancher.sh suse-observability-agent
