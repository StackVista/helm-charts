#!/usr/bin/env bash

set -euxo pipefail

chart_path=$1
tag_path_prefix=$2

sg_version="${UPDATE_STACKGRAPH_VERSION}"
values="${chart_path}/values.yaml"

# Check if version changed
current_version=$(yq read "${values}" "${tag_path_prefix}stackgraph.image.tag")

echo "Current StackGraph version: ${current_version}."
echo "New StackGraph version: ${sg_version}."

if ! [[ "${sg_version}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Failed to get valid stackgraph version."
  exit 1
fi

if [ "${sg_version}" == "${current_version}" ]; then
  echo "No change in StackGraph version, skipping update."
else
  # Update StackGraph version
  new_values=".values.yaml"
  yq w "${values}" "${tag_path_prefix}stackgraph.image.tag" "${sg_version}" > "${new_values}"
  mv -f "${new_values}" "${values}"

  # update Helm chart versions
  chart="${chart_path}/Chart.yaml"
  updated_chart=".chart.yaml"

  awk 'match($0,/^(version: .*\.)([0-9]+)$/,a) {a[2]++; $0=a[1] a[2]} { print }' "${chart}" > "${updated_chart}"
  mv -f "${updated_chart}" "${chart}"

  # update Readme
  readme="${chart_path}/README.md"
  new_readme=".readme.md"
  chart_version=$(yq r "${chart}" version)
  sed -E "s/^Current chart version is .*$/Current chart version is \`${chart_version}\`/" "${readme}" | \
  sed -E "s/${tag_path_prefix}stackgraph\.image\.tag \| string \| \`.*\` \|/${tag_path_prefix}stackgraph.image.tag | string | \`\"${sg_version}\"\` |/" > "${new_readme}"

  mv "${new_readme}" "${readme}"

  git add "${values}" "${chart_path}" "${readme}"
fi
