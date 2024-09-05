#!/bin/bash

cd stable/suse-observability-agent || exit

echo "Pushing container images to Rancher container registry"
./installation/o11y-agent-get-images.sh -d . > "o11y-agent-images.txt"
./installation/o11y-agent-save-images.sh
DST_REGISTRY_USERNAME=$RANCHER_CONTAINER_REGISTRY_USERNAME DST_REGISTRY_PASSWORD=$RANCHER_CONTAINER_REGISTRY_PASSWORD o11y-agent-load-images.sh -d "$RANCHER_CONTAINER_REGISTRY_URL"

echo "Reconfiguring container images registry in values.yaml"
yq -i -e  '.global.imageRegistry = env(RANCHER_CONTAINER_REGISTRY_URL)' values.yaml
sed -i -r "s|(\s+repository:\s+)stackstate/|\1${RANCHER_CONTAINER_REGISTRY_NAMESPACE}/|g" values.yaml

echo "Packaging and publishing helm chart..."
helm dependencies update .
helm cm-push --username "$RANCHER_HELM_REGISTRY_USERNAME" --password "$RANCHER_HELM_REGISTRY_PASSWORD" . "$RANCHER_HELM_REGISTRY_URL"
