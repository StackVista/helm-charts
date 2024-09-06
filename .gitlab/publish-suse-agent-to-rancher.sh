#!/bin/bash

cd stable/suse-observability-agent || exit

echo "Pushing container images to Rancher container registry"
helm dependencies update .

image_list_file="o11y-agent-images.txt"

./installation/o11y-agent-get-images.sh -d . > "${image_list_file}"

# Login to the rancher registry
if [ -n "$RANCHER_CONTAINER_REGISTRY_USERNAME" ] && [ -n "$RANCHER_CONTAINER_REGISTRY_PASSWORD" ]; then
  docker login -u "$RANCHER_CONTAINER_REGISTRY_USERNAME" -p "$RANCHER_CONTAINER_REGISTRY_PASSWORD" "$RANCHER_CONTAINER_REGISTRY_URL"
fi

# Pull the images from the list
while IFS= read -r image; do
  # Simple check if the image is in the format <registry>/<namespace>/<repository>:<tag>
  if [[ $image =~ ^([^/]+)/([^/]+)/(.*):([^:]+)$ ]]; then
    if docker pull "${image}" > /dev/null 2>&1; then
      echo -e "${GREEN}Image pull success${NO_COLOR}: ${image}"
      pulled="${pulled} ${image}"
    else
      echo -e "${RED}Image pull failed${NO_COLOR}: ${image}"
    fi

    repository_and_tag=$(echo "${image}" | cut -d'/' -f3-)
    dest_image="$RANCHER_CONTAINER_REGISTRY_URL/$RANCHER_CONTAINER_REGISTRY_NAMESPACE/${repository_and_tag}"
    docker tag "${image}" "${dest_image}"
    docker push "${dest_image}"
    # shellcheck disable=SC2181
    if [ $? -eq 0 ]; then
      echo -e "${GREEN}Successfully pushed ${dest_image}${NO_COLOR}"
    else
      echo -e "${RED}Failed to push ${dest_image}${NO_COLOR}"
    fi

    # Remove the image again to not fill up all diskspace on the runner
    docker rmi "${image}"
  else
      echo -e "${RED}Image url ${image} is not valid. Skipping...${NO_COLOR}"
  fi
done < "${image_list_file}"

echo "Reconfiguring container images registry in values.yaml"
yq -i -e  '.global.imageRegistry = env(RANCHER_CONTAINER_REGISTRY_URL)' values.yaml
sed -i -r "s|(\s+repository:\s+)stackstate/|\1${RANCHER_CONTAINER_REGISTRY_NAMESPACE}/|g" values.yaml

echo "Packaging and publishing helm chart..."
helm cm-push --username "$RANCHER_HELM_REGISTRY_USERNAME" --password "$RANCHER_HELM_REGISTRY_PASSWORD" . "$RANCHER_HELM_REGISTRY_URL"
