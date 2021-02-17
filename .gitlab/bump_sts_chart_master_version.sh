#!/bin/bash

set -euxo pipefail

repository=$1

latest_version=$(helm search repo --devel  "$repository/stackstate" -o json | jq --arg chart "$repository/stackstate" '.[] | select (.name == $chart) | .version' -r)
new_version_counter=$(awk 'match($0,/^(.*\.)([0-9]+)$/,a) {a[2]++; $0=a[2]} { print }' <<< "$latest_version")

sed -i -E -e "s/^(version: [0-9]+\.[0-9]+\.[0-9]+-snapshot\.)[0-9]+/\1${new_version_counter}/" stable/stackstate/Chart.yaml
