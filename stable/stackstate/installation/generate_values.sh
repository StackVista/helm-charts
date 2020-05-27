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
    -u : Username for Docker image pulling (required)
    -p : Password for Docker image pulling (required)
    -l : StackState license key (required)
    -b : StackState base URL, externally (outside of the Kubernetes cluster) visible url of the StackState endpoints (required)
         The exact value depends on your ingress setup. An example: https://my.stackstate.host
    -a : Administrator password that will be set for StackState (required)
    -v : Name of generated values file (default: values.yaml)
    -h : Show this help text
EOF
}

values_file="values.yaml"

# Parse arguments
while getopts "u:p:l:b:a:v:h" opt; do
  case "$opt" in
  u)  image_pull_credentials_username=$OPTARG ;;
  p)  image_pull_credentials_password=$OPTARG ;;
  l)  license_key=$OPTARG ;;
  b)  url=$OPTARG ;;
  a)  admin_password=$OPTARG ;;
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

  [ -z "${image_pull_credentials_username}" ] && read -r -p "Please provide the username for pulling StackState Docker images (-u): " image_pull_credentials_username
  [ -z "${image_pull_credentials_username}" ] && echo -e "${red}Username for pulling StackState Docker images (-u) is a required argument.${nc}" && exit 1

  if [ -z "${image_pull_credentials_password}" ]; then
    read -sr -p "Please provide the password for pulling StackState Docker images (-p): " image_pull_credentials_password
    [ -z "${image_pull_credentials_password}" ] && echo -e "${red}Password for pulling StackState Docker images (-p) is a required argument.${nc}" && exit 1
    echo ""
    read -s -r -p "Please repeat the password for confirmation: " image_pull_credentials_password_confirm
    echo ""
    [ "${image_pull_credentials_password}" != "${image_pull_credentials_password_confirm}" ] && echo -e "${red}Passwords mismatch.${nc}" && exit 1
  fi

  [ -z "${url}" ] && read -r -p "Please provide the base URL for StackState, for example https://my.stackstate.host (-b): " url
  [ -z "${url}" ] && echo -e "${red}The base url (-b) is a required argument.${nc}" && exit 1

  if [ -z "${admin_password}" ]; then
    read -s -r -p "Please provide the administrator password that will be set for StackState (-a): " admin_password
    [ -z "${admin_password}" ] && echo -e "${red}The administrator password (-a) is a required argument.${nc}" && exit 1
    echo ""
    read -s -r -p "Please repeat the password for confirmation: " admin_password_confirm
    echo ""
    [ "${admin_password}" != "${admin_password_confirm}" ] && echo -e "${red}Passwords mismatch.${nc}" && exit 1
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
global:
  receiverApiKey: "$(generate_api_key)"
stackstate:
  components:
    all:
      image:
        pullSecretUsername: "${image_pull_credentials_username}"
        pullSecretPassword: "${image_pull_credentials_password}"
    server:
      extraEnv:
        secret:
          CONFIG_FORCE_stackstate_api_authentication_authServer_stackstateAuthServer_defaultPassword: "$(create_admin_password_hash)"
  receiver:
    baseUrl: "${url}"
  license:
    key: "${license_key}"
hbase:
  all:
    image:
      pullSecretUsername: "${image_pull_credentials_username}"
      pullSecretPassword: "${image_pull_credentials_password}"
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
