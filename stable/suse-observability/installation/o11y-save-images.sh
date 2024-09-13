#!/usr/bin/env bash
set -o pipefail

GREEN='\033[0;32m'
RED='\033[0;31m'
NO_COLOR='\033[0m'

image_list_file="o11y-images.txt"
image_archive="o11y-images.tar.gz"

function usage() {
  cat <<EOF
Save Docker images needed for the Rancher Prime Observability chart to a TGZ archive

Usage:
  $0 -i o11y-images.txt -f o11y-images.tar.gz

Arguments:
    -i : File with the list of images (default: ${image_list_file})
    -f : TGZ Archive to save images to (default: ${image_archive})
    -h : Show this help text
EOF
}

# Parse options
while getopts ":f:i:h" opt; do
  case ${opt} in
    f)
      image_archive=${OPTARG}
      ;;
    i)
      image_list_file=${OPTARG}
      ;;
    h)
      usage
      exit 0
      ;;
    :)
      echo -e "${RED}Option -${OPTARG} requires an argument.${NO_COLOR}" >&2
      usage
      exit 1
      ;;
    *)
      echo -e "${RED}Unimplemented option: -${OPTARG}${NO_COLOR}" >&2
      usage
      exit 1
      ;;
  esac
done

# Check if the file with image list exists
if [ ! -f "${image_list_file}" ]; then
  echo -e "${RED}The file with the list of images is not found${NO_COLOR}: ${image_list_file}"
  exit 1
fi

# Pull the images from the list
while IFS= read -r image; do
  if docker pull "${image}" > /dev/null 2>&1; then
    echo -e "${GREEN}Image pull success${NO_COLOR}: ${image}"
    pulled="${pulled} ${image}"
  else
    echo -e "${RED}Image pull failed${NO_COLOR}: ${image}"
  fi
done < "${image_list_file}"


echo -e "Creating ${image_archive} with $(echo "${pulled}" | wc -w | tr -d '[:space:]') images"
# shellcheck disable=SC2086
docker save ${pulled} | gzip --stdout > "${image_archive}"
# shellcheck disable=SC2181
if [ $? -eq 0 ]; then
    echo -e "${GREEN}Images saved to ${image_archive}${NO_COLOR}"
else
    echo -e "${RED}Failed to save images to ${image_archive}${NO_COLOR}"
fi
