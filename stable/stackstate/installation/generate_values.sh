#!/usr/bin/env bash

set -e

#colors

nc="\033[0m"
red="\\033[0;31m"
green="\\033[0;32m"

function usage() {
  cat <<EOF
Generate a values.yaml for deploying StackState to Kubernetes with Helm 3.
If any of the required arguments are missing they will be asked interactively.

Arguments:
    -i : Image pull secret .dockerConfigJson property (required)
    -l : StackState license key (required)
    -u : StackState base URL; i.e. public (outside of the Kubernetes cluster) URL of StackState (required)
    -p : Administrator password (required)
    -v : Name of generated values file (default: values.yaml)
    -h : Show this help text
EOF
}

values_file="values.yaml"

# Parse arguments
while getopts "i:l:s:u:p:v:h" opt; do
  case "$opt" in
  i)  image_pull_secret_docker_config_json=$OPTARG ;;
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
  [ -z "${license_key}" ] && read -r -p "Please provide the license key (-l): " license_key
  [ -z "${license_key}" ] && echo -e "${red}License key (-l) is a required argument.${nc}" && exit 1

  [ -z "${image_pull_secret_docker_config_json}" ] && read -r -p "Please provide the image pull secret json (-i): " image_pull_secret_docker_config_json
  [ -z "${image_pull_secret_docker_config_json}" ] && echo -e "${red}image pul secret json (-i) is a required argument.${nc}" && exit 1

  [ -z "${url}" ] && read -r -p "Please provide the base URLfor StackState (-u): " url
  [ -z "${url}" ] && echo -e "${red}The base url (-u) is a required argument.${nc}" && exit 1

  [ -z "${admin_password}" ] && read -r -p "Please provide the administrator password (-p): " admin_password
  [ -z "${admin_password}" ] && echo -e "${red}The administrator password (-p) is a required argument.${nc}" && exit 1

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
      pullSecretDockerConfigJson: "${image_pull_secret_docker_config_json}"
    server:
      extraEnv:
        secret:
          CONFIG_FORCE_stackstate_api_authentication_authServer_stackstateAuthServer_defaultPassword: "$(create_admin_password_hash)"
  receiver:
    baseUrl: "${url}"
    apiKey: "$(generate_api_key)"
  license:
    key: "${license_key}"
hbase:
  components:
    all:
      image:
        pullSecretDockerConfigJson: "${image_pull_secret_docker_config_json}"
EOF
}

function print_helm_command() {
  echo -e "Generated ${green}'${values_file}'${nc}."
  echo -e ""
  echo -e "Make sure to ${green}store this file in a safe place${nc} for usage during upgrades of"
  echo -e "StackState. Generating a new file will generate a new API key which"
  echo -e "would require updating all running agents and other clients."
  echo -e ""
  echo -e "Now for first time installation create a namespace for StackState first:"
  echo -e ""
  echo -e "${green}kubectl create namespace stackstate${nc}"
  echo -e ""
  echo -e "Use the following command to install or upgrade StackState into"
  echo -e "namespace 'stackstate' using helm release 'stackstate':"
  echo -e ""
  echo -e "${green}helm upgrade --install --namespace stackstate --values values.yaml stackstate stackstate/stackstate${nc}"
  echo -e ""
}

check_helm
check_args
generate_values
print_helm_command
