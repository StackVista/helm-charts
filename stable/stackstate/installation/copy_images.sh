#!/usr/bin/env bash

set -e

# global
is_ecr_re='\.ecr\..*\.amazonaws\.com'
repo_and_tag_re='/([^/]*/[^/:]*):([^/:]*)$'
nc="\033[0m"
red="\\033[0;31m"

# usage
function usage() {
  cat <<EOF
Copy Docker images needed for StackState chart to another Docker image registry

Arguments:
    -d : Destination Docker image registry (required)
    -h : Show this help text
EOF
}

# parse options
while getopts "d:h" opt; do
  case "$opt" in
  d) dest_registry=$OPTARG ;;
  h) usage; exit;;
  \?) echo "Unknown option: -$OPTARG" >&2; exit 1;;
  :) echo "Missing option argument for -$OPTARG" >&2; exit 1;;
  *) echo "Unimplemented option: -$OPTARG" >&2; exit 1;;
  esac
done
shift $((OPTIND -1))

[ -z "$dest_registry" ] && echo -e "${red}Provide the destination registry with the -d flag${nc}" && exit 1

#
images=()
while IFS='' read -r line; do images+=("$line"); done < <(helm template stackstate stackstate --repo https://helm.stackstate.io --set stackstate.license.key=dummy | grep image: | sed -E 's/^.*image: "?([^"]*)"?.*$/\1/')
for src_image in "${images[@]}"
do
    if [[ "$src_image" =~ $repo_and_tag_re ]]; then
        repo=${BASH_REMATCH[1]}
        tag=${BASH_REMATCH[2]}
        dest_image="$dest_registry/$repo:$tag"
        if [[ "$dest_registry" =~ $is_ecr_re ]]; then
            echo "Ensuring ECR repository $repo exists"
            aws ecr describe-repositories --repository-names "$repo" >/dev/null 2>/dev/null || aws ecr create-repository --repository-name "$repo" > /dev/null
        fi
        echo "Copying $src_image to $dest_image"
        docker pull "$src_image"
        docker tag "$src_image" "$dest_image"
        docker push "$dest_image"
        docker rmi "$dest_image"
    else
        1>&2 echo "Cannot determine repository and tag for $src_image"
        exit 1
    fi
done
