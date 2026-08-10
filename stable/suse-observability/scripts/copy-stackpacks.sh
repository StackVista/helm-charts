#!/bin/sh

set -e

target_dir=$1

if [ -z "${target_dir}" ]; then
  echo "Usage: $0 <target_dir> [--clear]"
  exit 1
fi

mkdir -p "${target_dir}"

if [ "$#" -ge 2 ] && [ "$2" = "--clear" ]; then
  echo "Cleaning up current StackPacks in ${target_dir}/"
  rm -rf "${target_dir:?}/"* || true
fi

echo "Copying StackPacks from /stackpacks to ${target_dir}/"
cp /stackpacks/*.sts "${target_dir}/" 2>/dev/null || true
