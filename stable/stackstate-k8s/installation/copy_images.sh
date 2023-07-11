#!/usr/bin/env bash

set -e

# global
is_ecr_re='\.ecr\..*\.amazonaws\.com'
repo_and_tag_re='^([^/:]+/)?([^/:]+/[^/:]+):([^/:]+)$'
nc="\033[0m"
red="\\033[0;31m"
helm_chart=stackstate-k8s
helm_repository=https://helm.stackstate.io
helm_values="backup.enabled=true,minio.accessKey=ABCDEFGH,minio.secretKey=ABCDEFGHABCDEFGH,stackstate.baseUrl=http://dummy.stackstate.io,stackstate.admin.authentication.password=dummy,stackstate.authentication.adminPassword=dummy,stackstate.license.key=dummy,global.receiverApiKey=dummy"
dry_run=false

# usage
function usage() {
  cat <<EOF
Copy Docker images needed for StackState chart to another Docker image registry

Arguments:
    -c : Helm chart (default: $helm_chart)
    -u : StackState repository username (required)
    -p : StackState repository password (required)
    -d : Destination Docker image registry (required)
    -h : Show this help text
    -r : Helm repository (default: $helm_repository)
    -t : Dry-run
EOF
}


# parse options
while getopts "c:u:p::d:hr:t" opt; do
  case "$opt" in
  c) helm_chart=$OPTARG ;;
  u) src_registry_username=$OPTARG ;;
  p) src_registry_password=$OPTARG ;;
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

[ -z "$src_registry_username" ] && echo -e "${red}Provide the StackState repository username with the -u flag${nc}" && usage && exit 1
[ -z "$src_registry_password" ] && echo -e "${red}Provide the StackState repository password with the -p flag${nc}" && usage && exit 1

docker container run -i --rm --net host -v "$(pwd)/regctl:/home/appuser/.regctl/" ghcr.io/regclient/regctl:latest registry login -u "$src_registry_username" -p "$src_registry_password" "quay.io"


#
images=()
while IFS='' read -r line; do images+=("$line"); done < <(helm template stackstate-k8s "$helm_chart" --repo "$helm_repository" --set "$helm_values" | grep image: | sed -E 's/^.*image: ['\''"]?([^'\''"]*)['\''"]?.*$/\1/')
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
            docker container run -i --rm --net host  -v "$(pwd)/regctl:/home/appuser/.regctl/" ghcr.io/regclient/regctl:latest image copy "$src_image" "$dest_image"
        fi
    else
        1>&2 echo "Cannot determine repository and tag for $src_image"
        exit 1
    fi
done
