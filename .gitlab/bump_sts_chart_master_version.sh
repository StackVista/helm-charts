#!/bin/bash

set -euxo pipefail

repository=$1
chart=$2

latest_version=$(helm search repo --devel  "$repository/$chart" -o json | jq --arg chart "$repository/$chart" '.[] | select (.name == $chart) | .version' -r)
latest_version_counter=$(sed -E -e 's/^(.*\.)([0-9]+)(-.*)?$/\2/' <<< "$latest_version")
new_version_counter=$((latest_version_counter+1))

sed -i -E -e "s/^(version: [0-9]+\.[0-9]+\.[0-9]+-snapshot\.)[0-9]+(-.*)?/\1${new_version_counter}/" "stable/$chart/Chart.yaml"
