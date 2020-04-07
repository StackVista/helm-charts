#!/usr/bin/env bash

set -e


function usage() {
  echo "Generate a values.yaml for deploying StackState to Kubernetes with Helm 3."
  echo "TODO: more explanation"
}

values_file="values.yaml"

# Parse arguments
while getopts "i:l:s:u:p:v:h" opt; do
  case "$opt" in
  i)  image_pull_secret=$OPTARG ;;
  l)  license_key=$OPTARG ;;
  u)  url=$OPTARG ;;
  p)  admin_password=$OPTARG ;;
  v)  values_file=$OPTARG ;;
  h)  usage; exit;;
  \?) echo "Unknown option: -$OPTARG" >&2; exit 1;;
  :) echo "Missing option argument for -$OPTARG" >&2; exit 1;;
  *) echo "Unimplemented option: -$OPTARG" >&2; exit 1;;
  esac
done

function check_args() {
  error=false

  [ -z "${image_pull_secret}" ] && echo "image_pull_secret (-i) is a required argument." && error=true
  [ -z "${license_key}" ] && echo "License key (-l) is a required argument." && error=true
  [ -z "${url}" ] && echo "The base url (-u) is a required argument." && error=true
  [ -z "${admin_password}" ] && echo "The administrator password (-p) is a required argument." && error=true

  if [ "$error" = true ]; then
    exit 1
  fi

  return 0
}

function check_helm() {
  helm_version=$(helm version --short | cut -d. -f1)
  if [ "${helm_version}" != "v3" ]; then
    echo "Helm version 3 is required, found version ${helm_version}. Please use Helm 3 to deploy StackState."
    exit 1;
  fi
}

function generate_api_key() {
  head -c32 < /dev/urandom | md5sum | cut -c-32
}

function create_admin_password_hash() {
  echo -n "${admin_password}" | md5sum | cut -c-32
}

function generate_values() {
  cat > "${values_file}" <<EOF
stackstate:
  components:
    all:
    image:
      pullSecret: "${image_pull_secret}"
    server:
      extraEnv:
        secret:
          CONFIG_FORCE_stackstate_api_authentication_authServer_stackstateAuthServer_defaultPassword: "$(create_admin_password_hash)"
  receiver:
    baseUrl: "${url}"
    apiKey: "$(generate_api_key)"
  license:
    key: "${license_key}"
EOF
}

function print_helm_command() {
  cat <<EOF
Generated '${values_file}'.

Make sure to store this file in a safe place for usage during upgrades of
StackState. Generating a new file will generate a new API key which
would require updating all running agents and other clients.

Use the following command to install or upgrade StackState into
namespace 'stackstate' using helm release 'stackstate':

helm upgrade --install --namespace stackstate --values values.yaml stackstate stackstate/stackstate
EOF
}

check_helm
check_args
generate_values
print_helm_command
