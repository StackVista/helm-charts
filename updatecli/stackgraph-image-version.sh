#!/usr/bin/env bash
set -euo pipefail

tag="${1:?Usage: stackgraph-image-version.sh <published-server-tag>}"
version="$(skopeo inspect --no-tags --format '{{ index .Labels "com.stackstate.stackgraph.version" }}' \
  "docker://quay.io/stackstate/stackstate-server:${tag}")"
if [[ -z "$version" || "$version" == "<no value>" || "$version" == "null" ]]; then
  echo "com.stackstate.stackgraph.version label missing on stackstate-server image" >&2
  exit 1
fi

# StackGraph #84 changes only image OS packages, not Java artifacts or protocols.
# Keep its published fix when the backend still embeds the 8.3.13 client.
# Do not apply a version floor to other releases or branch builds.
case "$version" in
  8.3.13) version=8.3.14 ;;
esac
printf '%s' "$version"
