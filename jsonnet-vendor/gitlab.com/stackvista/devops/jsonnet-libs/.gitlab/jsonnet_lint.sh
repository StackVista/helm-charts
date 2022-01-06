#!/bin/sh

set -e

find . -type f -regex '.*\(jsonnet\|libsonnet\)$' > tmp-file
while IFS= read -r file
do
  echo "Checking ${file}..."
  jsonnet fmt --in-place "${file}"
done < tmp-file
rm tmp-file
