#!/bin/bash

repository=$1

latest_version=$(helm search repo --devel  "$repository/stackstate" -o json | jq '.[] | select (.name == "$repository/stackstate") | .version' -r)
new_version=$(awk 'match($0,/^(.*\.)([0-9]+)$/,a) {a[2]++; $0=a[1] a[2]} { print }' <<< "$latest_version")

sed -i -E -e "s/^version: [0-9]+\.[0-9]+\.[0-9]+-snapshot\.[0-9]+/version: ${new_version}/" stable/stackstate/Chart.yaml
