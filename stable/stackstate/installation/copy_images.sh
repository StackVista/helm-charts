#!/usr/bin/env bash

set -e

# global
is_ecr_re='\.ecr\..*\.amazonaws\.com'
repo_and_tag_re='^([^/:]+/)?([^/:]+/[^/:]+):([^/:]+)$'
nc="\033[0m"
red="\\033[0;31m"
helm_chart=stackstate
helm_repository=https://helm.stackstate.io
helm_values="backup.enabled=true,minio.accessKey=ABCDEFGH,minio.secretKey=ABCDEFGHABCDEFGH,stackstate.baseUrl=http://dummy.stackstate.io,stackstate.admin.authentication.password=dummy,stackstate.authentication.adminPassword=dummy,stackstate.license.key=dummy,global.receiverApiKey=dummy"
dry_run=false

# usage
function usage() {
  cat <<EOF
Copy Docker images needed for StackState chart to another Docker image registry

Environment Variables:
    STS_REGISTRY_USERNAME : StackState repository username (required)
    STS_REGISTRY_PASSWORD : StackState repository password (required)
    DST_REGISTRY_USERNAME : Destination Docker image registry username
    DST_REGISTRY_PASSWORD : Destination Docker image registry password

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
[ -z "$STS_REGISTRY_USERNAME" ] && echo -e "${red}Provide the StackState repository username with the \$STS_REGISTRY_USERNAME environment variable${nc}" && usage && exit 1
[ -z "$STS_REGISTRY_PASSWORD" ] && echo -e "${red}Provide the StackState repository password with the \$STS_REGISTRY_PASSWORD environment variable${nc}" && usage && exit 1

# Create the regctl directory for storing the config between docker runs
CFG_DIR=$(mktemp -d)

docker container run -i --rm --net host -v "${CFG_DIR}:/home/appuser/.regctl/" ghcr.io/regclient/regctl:latest registry login -u "$STS_REGISTRY_USERNAME" -p "$STS_REGISTRY_PASSWORD" "quay.io"

if [ -n "$DST_REGISTRY_USERNAME" ] && [ -n "$DST_REGISTRY_PASSWORD" ]; then
    docker container run -i --rm --net host -v "${CFG_DIR}:/home/appuser/.regctl/" ghcr.io/regclient/regctl:latest registry login -u "$DST_REGISTRY_USERNAME" -p "$DST_REGISTRY_PASSWORD" "$dest_registry"
fi

#
images=()
while IFS='' read -r line; do images+=("$line"); done < <(helm template stackstate "$helm_chart" --repo "$helm_repository" --set "$helm_values" | grep image: | sed -E 's/^.*image: ['\''"]?([^'\''"]*)['\''"]?.*$/\1/')
# Remove duplicates
IFS=" " read -r -a images <<< "$(echo "${images[@]}" | tr ' ' '\n' | sort -u | tr '\n' ' ')"

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
            docker container run -i --rm --net host -v "${CFG_DIR}:/home/appuser/.regctl/" ghcr.io/regclient/regctl:latest image copy "$src_image" "$dest_image"
        fi
    else
        1>&2 echo "Cannot determine repository and tag for $src_image"
        exit 1
    fi
done

rm -rf "${CFG_DIR}"
