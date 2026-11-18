#!/usr/bin/env bash
set -euo pipefail

tag="${1:?agent release tag is required}"
destination="${2:?canonical CRD destination is required}"
if [[ ! "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+-agent$ ]]; then
    echo "CollectionDeclaration adoption requires a released vX.Y.Z-agent image" >&2
    exit 1
fi

source_url="${COLLECTOR_SOURCE_URL:-https://raw.githubusercontent.com/StackVista/sts-opentelemetry-collector}"
temporary="$(mktemp "${destination}.XXXXXX")"
trap 'rm -f "$temporary"' EXIT
curl --fail --silent --show-error --location \
    "${source_url}/${tag%-agent}/receiver/k8sresourcereceiver/collectiondeclaration/collectiondeclaration-crd.yaml" \
    --output "$temporary"
if cmp -s "$temporary" "$destination"; then
    exit 0
fi
chmod 0644 "$temporary"
if [[ "${DRY_RUN:-false}" != "true" ]]; then
    mv "$temporary" "$destination"
fi
printf 'Update CollectionDeclaration CRD from %s\n' "$tag"
