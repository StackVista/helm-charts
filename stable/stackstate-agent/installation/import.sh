#!/usr/bin/env bash

set -euo pipefail

# global
nc="\033[0m"
red="\\033[0;31m"
push=false
dry_run=false

# usage
function usage() {
  cat <<EOF
Import previously exported docker images to an environment with a docker installation, optionally push to new registry.

Arguments:
    -b : Path to backed up images (tar.gz) file
    -d : Destination Docker image registry (required)
    -h : Show this help text
    -p : Push images to (destination) repository
    -t : Dry-run
EOF
}

# parse options
while getopts "b:d:hpt" opt; do
  case "$opt" in
  b) backup=$OPTARG ;;
  d) dest_registry=$OPTARG ;;
  h) usage; exit ;;
  p) push=true ;;
  t) dry_run=true ;;
  \?) echo "Unknown option: -$OPTARG" >&2; exit 1;;
  :) echo "Missing option argument for -$OPTARG" >&2; exit 1;;
  *) echo "Unimplemented option: -$OPTARG" >&2; exit 1;;
  esac
done
shift $((OPTIND -1))

[ -z "$backup" ] && echo -e "${red}Provide the backup archive file with the -b flag${nc}" && usage && exit 1
[ -z "$dest_registry" ] && echo -e "${red}Provide the destination registry with the -d flag${nc}" && usage && exit 1

folder=$(tar tzf "${backup}" | sed -e 's@/.*@@' | uniq)
if [ -d "${folder}" ]; then
  rm -rf "${folder}"
fi

echo "Unzipping archive $backup"
# unzip the backed-up images
tar -zxvf "${backup}"

cd "${folder}" || exit
ARCHIVES=$(ls ./*.tar)
cd .. || exit
#
images=()
while IFS='' read -r line; do images+=("$line"); done < <(echo "${ARCHIVES}")
for src_file in "${images[@]}"
do
    src_file="${src_file#*/}"
    organization="${folder}"
    repo=${src_file%__*}
    tag_and_ext=${src_file#*__}
    tag=${tag_and_ext%%.tar*}

    if $dry_run; then
        echo "Restoring ${organization}/${repo}:${tag} from ${src_file} (dry-run)"
        # assume quay.io
        echo "Imported quay.io/${organization}/${repo}:${tag}"
        echo "Tagged quay.io/${organization}/${repo}:${tag} as ${dest_registry}/${organization}/${repo}:${tag}"
        echo "Untagged: quay.io/${organization}/${repo}:${tag}"
        if $push; then
          echo "Pushed ${dest_registry}/${organization}/${repo}:${tag}"
        fi
    else
        echo "Restoring ${organization}/${repo}:${tag} from ${src_file}"
        something=$(docker load -i "${folder}/${src_file}")
        something_else="${something#*/}"
        something=$(echo "${something#*:}" | awk '{print $1}')
        docker tag "${something}" "${dest_registry}/${something_else}"
        docker rmi "${something}"

        if $push; then
          docker push "${dest_registry}/${something_else}";
        fi
    fi
done

echo "Images have been imported up to $dest_registry"
rm -rf "${folder}"
