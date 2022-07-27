#!/usr/bin/env bash

set -e

# global
is_ecr_re='\.ecr\..*\.amazonaws\.com'
repo_and_tag_re='^([^/:]+/)?([^/:]+/[^/:]+):([^/:]+)$'
nc="\033[0m"
red="\\033[0;31m"
helm_release=release
helm_chart="stackstate/cluster-agent"
helm_repository=https://helm.stackstate.io
helm_values="stackstate.apiKey=APIKEY,stackstate.cluster.name=THISCLUSTERNAME,stackstate.url=http://dummy.stackstate.io"
dry_run=false

# usage
function usage() {
  cat <<EOF
Copy Docker images needed for StackState chart to another Docker image registry

Arguments:
    -c : Helm chart (default: $helm_chart)
    -d : Destination Docker image registry (required)
    -h : Show this help text
    -r : Helm repository (default: $helm_repository)
    -t : Dry-run
EOF
}


# parse options
while getopts "c:d:hr:t" opt; do
  case "$opt" in
  c) helm_chart=$OPTARG ;;
  d) dest_registry=$OPTARG ;;
  h) usage; exit ;;
  r) helm_repository=$OPTARG ;;
  t) dry_run=true ;;
  \?) echo "Unknown option: -$OPTARG" >&2; exit 1;;
  :) echo "Missing option argument for -$OPTARG" >&2; exit 1;;
  *) echo "Unimplemented option: -$OPTARG" >&2; exit 1;;
  esac
done
shift $((OPTIND -1))

[ -z "$dest_registry" ] && echo -e "${red}Provide the destination registry with the -d flag${nc}" && usage && exit 1

#
images=()
while IFS='' read -r line; do images+=("$line"); done < <(helm template "$helm_release" "$helm_chart" --set "$helm_values" | grep image: | sed -E 's/^.*image: ['\''"]?([^'\''"]*)['\''"]?.*$/\1/')
for src_image in "${images[@]}"
do
    if [[ "$src_image" =~ $repo_and_tag_re ]]; then
        repo=${BASH_REMATCH[2]}
        tag=${BASH_REMATCH[3]}
        dest_image="$dest_registry/$repo:$tag"
        if $dry_run; then
            echo "Copying $src_image to $dest_image (dry-run)"
        else
            if [[ "$dest_registry" =~ $is_ecr_re ]]; then
                echo "Ensuring ECR repository $repo exists"
                aws ecr describe-repositories --repository-names "$repo" >/dev/null 2>/dev/null || aws ecr create-repository --repository-name "$repo" > /dev/null
            fi
            echo "Copying $src_image to $dest_image"
            docker pull "$src_image"
            docker tag "$src_image" "$dest_image"
            docker push "$dest_image"
            docker rmi "$dest_image"
        fi
    else
        1>&2 echo "Cannot determine repository and tag for $src_image"
        exit 1
    fi
done
