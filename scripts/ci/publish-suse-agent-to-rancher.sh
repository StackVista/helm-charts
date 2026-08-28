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

# copySignature moves both cosign signature forms that skopeo copy leaves behind:
# the legacy tag-based sha256-<digest>.sig manifest and the newer
# OCI-referrers-linked bundle. Both are mandatory; a missing or failed copy of
# either aborts the release (exit 1) rather than shipping an unsigned image.
function copySignature() {
  local src_image=$1
  local dst_image=$2
  local status=0

  local digest
  digest=$(skopeo inspect "docker://${src_image}" | jq -r '.Digest')
  local sig_suffix="${digest/sha256:/sha256-}.sig"
  local src_repo="${src_image%:*}"
  local dst_repo="${dst_image%:*}"
  local src_sig_ref="${src_repo}:${sig_suffix}"
  local dst_sig_ref="${dst_repo}:${sig_suffix}"

  # Legacy: an ordinary tagged manifest, so skopeo copies it like any image.
  echo "Copying legacy signature ${src_sig_ref} to ${dst_sig_ref}"
  if ! skopeo copy --dest-username "$RANCHER_CONTAINER_REGISTRY_USERNAME" --dest-password "$RANCHER_CONTAINER_REGISTRY_PASSWORD" "docker://${src_sig_ref}" "docker://${dst_sig_ref}"; then
    echo -e "${RED}Missing or failed legacy signature ${src_sig_ref} for ${src_image}${NO_COLOR}" >&2
    status=1
  fi

  # Referrer-linked: reachable only via the digest's referrers, which oras reads.
  # --from-* left unset so the public quay.io source stays anonymous; -r also
  # re-copies the subject manifest (same bytes skopeo already pushed), as oras
  # has no referrers-only copy mode.
  echo "Copying referrer-linked signatures for ${src_repo}@${digest} to ${dst_repo}"
  if ! oras copy -r "${src_repo}@${digest}" "${dst_repo}" \
    --to-username "$RANCHER_CONTAINER_REGISTRY_USERNAME" \
    --to-password "$RANCHER_CONTAINER_REGISTRY_PASSWORD"; then
    echo -e "${RED}Missing or failed referrer-linked signatures for ${src_repo}@${digest}${NO_COLOR}" >&2
    status=1
  fi

  if [ "$status" -ne 0 ]; then
    echo -e "${RED}Signature copy failed for ${src_image}: both a legacy and a referrer-linked signature are required. Aborting release.${NO_COLOR}" >&2
    exit 1
  fi
}

# A successful copy does not prove the destination finished linking the
# referrer manifest; only a real cosign verify against Rancher catches a
# half-pushed bundle.
function verifySignatureCopy() {
  local dst_image=$1

  echo "Verifying copied signature for ${dst_image}"
  if ! cosign verify \
    --certificate-oidc-issuer=https://token.actions.githubusercontent.com \
    --certificate-identity-regexp='^https://github.com/StackVista/.*' \
    --registry-username "$RANCHER_CONTAINER_REGISTRY_USERNAME" \
    --registry-password "$RANCHER_CONTAINER_REGISTRY_PASSWORD" \
    "${dst_image}" >/dev/null; then
    echo -e "${RED}Signature verification failed for ${dst_image} after copying to Rancher${NO_COLOR}" >&2
    exit 1
  fi
}

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
      if skopeo copy --all --dest-username "$RANCHER_CONTAINER_REGISTRY_USERNAME" --dest-password "$RANCHER_CONTAINER_REGISTRY_PASSWORD" "docker://${image}" "docker://${dest_image}"; then
        echo -e "${GREEN}Successfully copied ${dest_image}${NO_COLOR}"
      else
        echo -e "${RED}Failed to copy ${dest_image}${NO_COLOR}" >&2
        exit 1
      fi
      copySignature "${image}" "${dest_image}"
      verifySignatureCopy "${dest_image}"
    else
      echo "[DRY_RUN] skipping skopeo copy + signature copy/verify → ${dest_image}"
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
