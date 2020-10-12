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
    -d : Password that will be set for the default StackState 'admin' user (required)
    -a : Password that will be set for StackState's admin api, access should be restricted (Dev)Ops (required)
    -v : Name of generated values file (default: values.yaml)
    -k : Install the StackState Kubernetes agent, StackState and the agent will refer to the cluster by thisn name (for convenience we suggest to use the same name as in your kube context) (optional)
    -n : Non-interactive mode
    -h : Show this help text
EOF
}

values_file="values.yaml"

stackstate_cluster_agent_enabled=false
interactive=true

# Parse arguments
while getopts "u:p:l:b:a:v:d:k:nh" opt; do
  case "$opt" in
  u)  image_pull_credentials_username=$OPTARG ;;
  p)  image_pull_credentials_password=$OPTARG ;;
  l)  license_key=$OPTARG ;;
  b)  url=$OPTARG ;;
  a)  admin_api_password=$OPTARG ;;
  d)  default_admin_password=$OPTARG ;;
  v)  values_file=$OPTARG ;;
  n)  interactive=false ;;
  k)  k8s_cluster_name=$OPTARG; stackstate_cluster_agent_enabled=true ;;
  h)  usage; exit;;
  \?) echo "Unknown option: -$OPTARG" >&2; exit 1;;
  :) echo "Missing option argument for -$OPTARG" >&2; exit 1;;
  *) echo "Unimplemented option: -$OPTARG" >&2; exit 1;;
  esac
done

function check_args() {
  [ -z "${license_key}" ] && $interactive  && read -r -p "Please provide the license key (-l): " license_key
  [ -z "${license_key}" ] && echo -e "${red}License key (-l) is a required argument.${nc}" && exit 1

  [ -z "${image_pull_credentials_username}" ]  && $interactive  && read -r -p "Please provide the username for pulling StackState Docker images (-u): " image_pull_credentials_username
  [ -z "${image_pull_credentials_username}" ] && echo -e "${red}Username for pulling StackState Docker images (-u) is a required argument.${nc}" && exit 1

  if [ -z "${image_pull_credentials_password}" ] && $interactive; then
    read -sr -p "Please provide the password for pulling StackState Docker images (-p): " image_pull_credentials_password
    echo ""
    read -s -r -p "Please repeat the password for confirmation: " image_pull_credentials_password_confirm
    echo ""
    [ "${image_pull_credentials_password}" != "${image_pull_credentials_password_confirm}" ] && echo -e "${red}Passwords mismatch.${nc}" && exit 1
  fi
  [ -z "${image_pull_credentials_password}" ] && echo -e "${red}Password for pulling StackState Docker images (-p) is a required argument.${nc}" && exit 1

  [ -z "${url}" ] && $interactive && read -r -p "Please provide the base URL for StackState, for example https://my.stackstate.host (-b): " url
  [ -z "${url}" ] && echo -e "${red}The base url (-b) is a required argument.${nc}" && exit 1

  if [ -z "${default_admin_password}" ] && $interactive; then
    read -s -r -p "Please provide the password for the 'admin' user that will be set for StackState (-d): " default_admin_password
    echo ""
    read -s -r -p "Please repeat the password for confirmation: " default_admin_password_confirm
    echo ""
    [ "${default_admin_password}" != "${default_admin_password_confirm}" ] && echo -e "${red}Passwords mismatch.${nc}" && exit 1
  fi
  [ -z "${default_admin_password}" ] && echo -e "${red}The 'admin' password (-d) is a required argument.${nc}" && exit 1

  if [ -z "${admin_api_password}" ] && $interactive; then
    read -s -r -p "Please provide the password that will be set for StackState's admin api (-a): " admin_api_password
    echo ""
    read -s -r -p "Please repeat the password for confirmation: " admin_api_password_confirm
    echo ""
    [ "${admin_api_password}" != "${admin_api_password_confirm}" ] && echo -e "${red}Passwords mismatch.${nc}" && exit 1
  fi
  [ -z "${admin_api_password}" ] && echo -e "${red}The password for StackState's admin api (-a) is a required argument.${nc}" && exit 1

  if [ -z "${k8s_cluster_name}" ] && $interactive; then
    echo "StackState has Kubernetes support for which an agent needs to be installed on the cluster."
    echo "See the Kubernetes Stackpack at https://docs.stackstate.com/ for the details."
    while true; do
      read -r -p "Do you want to install the StackState Kubernetes Agent on the cluster? [yn] " yn
      case $yn in
        [Yy]* )
          stackstate_cluster_agent_enabled=true
          printf "\n\n"
          echo "StackState and the agent use a cluster name to refer to this Kubernetes cluster."
          echo "For convenience we suggest to use the same name as in your kube context."
          read -r -p "Please provide a name for the kubernetes cluster : " k8s_cluster_name
          [ -z "${k8s_cluster_name}" ] && echo -e "${red}The Kubernetes cluster name is required when installing the agent.${nc}" && exit 1
          break;;
        [Nn]* )
          stackstate_cluster_agent_enabled=false
          break;;
        * ) echo "Please answer yes or no.";;
      esac
    done
  fi

  return 0
}

function check_md5() {
  # Check if md5sum is installed as default on OSX its called md5
  if [ ! -z "$(command -v md5sum 2>/dev/null)" ] ; then
     md5sum_command=$(command -v md5sum)
  else
     md5sum_command=$(command -v md5)
  fi
}

function check_helm() {
  helm_version=$(helm version --short | cut -d. -f1)
  if [ "${helm_version}" != "v3" ]; then
    echo "Helm version 3 is required, found version ${helm_version}. Please use Helm 3 to deploy StackState."
    exit 1;
  fi
}

function generate_api_key() {
  head -c32 < /dev/urandom | $md5sum_command | cut -c-32
}

function create_admin_api_password_hash() {
  echo -n "${admin_api_password}" | $md5sum_command | cut -c-32
}

function create_default_admin_password_hash() {
  echo -n "${default_admin_password}" | $md5sum_command  | cut -c-32
}

function create_agent_auth_token() {
  head -c32 < /dev/urandom | $md5sum_command  | cut -c-32
}

function generate_values() {
  cat > "${values_file}" <<EOF
global:
  receiverApiKey: "$(generate_api_key)"
hbase:
  all:
    image:
      pullSecretUsername: "${image_pull_credentials_username}"
      pullSecretPassword: "${image_pull_credentials_password}"
stackstate:
  components:
    all:
      image:
        pullSecretUsername: "${image_pull_credentials_username}"
        pullSecretPassword: "${image_pull_credentials_password}"
  receiver:
    baseUrl: "${url}"
  license:
    key: "${license_key}"
  authentication:
    adminPassword: "$(create_default_admin_password_hash)"
  admin:
    authentication:
      password: "$(create_admin_api_password_hash)"
EOF

  if ${stackstate_cluster_agent_enabled}; then
  cat >> "${values_file}" <<EOF
  stackpacks:
    installed:
      - name: kubernetes
        configuration:
          kubernetes_cluster_name: "${k8s_cluster_name}"
cluster-agent:
  enabled: true
  stackstate:
    cluster:
      name: "${k8s_cluster_name}"
      authToken: "$(create_agent_auth_token)"
EOF
  fi
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

check_md5
check_helm
check_args
generate_values
print_helm_command
