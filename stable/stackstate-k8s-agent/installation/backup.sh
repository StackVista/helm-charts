#!/usr/bin/env bash

set -euo pipefail

# global
repo_and_tag_re='^([^/:]+/)?([^/:]+/[^/:]+):([^/:]+)$'
helm_release=release
helm_chart="stackstate/stackstate-k8s-agent"
helm_values="http-header-injector-webhook.enabled=true,stackstate.apiKey=APIKEY,stackstate.cluster.name=dummy-cluster,stackstate.url=http://dummy.stackstate.io"
dry_run=false

# usage
function usage() {
  cat <<EOF
Back up helm chart images to a tar.gz archive for easy transport via external storage device.

Arguments:
    -c : Helm chart (default: $helm_chart)
    -h : Show this help text
    -t : Dry-run
EOF
}

# parse options
while getopts "c:hr:t" opt; do
  case "$opt" in
  c) helm_chart=$OPTARG ;;
  h) usage; exit ;;
  t) dry_run=true ;;
  \?) echo "Unknown option: -$OPTARG" >&2; exit 1;;
  :) echo "Missing option argument for -$OPTARG" >&2; exit 1;;
  *) echo "Unimplemented option: -$OPTARG" >&2; exit 1;;
  esac
done
shift $((OPTIND -1))

#
images=()
while IFS='' read -r line; do images+=("$line"); done < <(helm template "$helm_release" "$helm_chart" --set "$helm_values" | grep image: | sed -E 's/^.*image: ['\''"]?([^'\''"]*)['\''"]?.*$/\1/' | sort | uniq)
for src_image in "${images[@]}"
do
    if [[ "$src_image" =~ $repo_and_tag_re ]]; then
        organization_and_repo_and_tag=${src_image#*/}     # everything after the first '/' char, not aggressive (strips off quay.io, returns the rest)
        organization=${organization_and_repo_and_tag%/*}  # everything before the first '/' char, not aggressive (strips off stackstate and returns it)
        repo_and_tag=${src_image##*/}                     # everything after the last '/' char, aggressive (strips off quay.io/stackstate, returns the rest)
        repo=${repo_and_tag%%:*}                          # everything before the last ':' char, aggressive (removes tag, returns image name)
        tag=${repo_and_tag##*:}                           # everything after the last ':' char, aggressive (removes image name, returns tag)

        if $dry_run; then
            echo "Backing up $src_image to $organization/${repo}__${tag}.tar (dry-run)"
        else
            if [ ! -d "${organization}" ]; then
              mkdir "${organization}"
            fi
            echo "Backing up $src_image to $organization/${repo}__${tag}.tar"
            docker pull "$src_image"
            if [ -f "$organization/${repo}__${tag}.tar" ]; then
              rm "$organization/${repo}__${tag}.tar"
            fi
            docker save -o "$organization/${repo}__${tag}.tar" "$src_image"
            docker rmi "$src_image"
        fi
    else
        1>&2 echo "Cannot determine repository and tag for $src_image"
        exit 1
    fi
done

if ! $dry_run; then
  tar -zcvf "${organization}.tar.gz" "${organization}"
  rm -rf "${organization}"
fi

echo "Images have been backed up to ${organization}.tar.gz"
