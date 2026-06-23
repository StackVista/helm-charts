#!/bin/bash

set -e

DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
if [[ "$DIR" = "." ]]; then DIR="$PWD"; fi

BASEPATH="$DIR"
function printUsage() {
  cat << USAGE

Usage: ./install.sh [-n <namespace>] --setup|--destroy|--list|--help [deployment]

This script takes all deployments from this example directory and allows easy install and uninstall of those examples. Options:
  -n/--namespace -- Override the default namespace (which is the deployment name)
  -s/--setup     -- Setup the provided deployment
  -d/--destroy   -- Destroy the provided deployment
  -l/--list      -- List available deployments
  -h/--help      -- Print this help page

USAGE
}

NAMESPACE=""
DEPLOYMENT=""
ACTION=""

while [[ $# -gt 0 ]]; do
  case $1 in
    -n|--namespace)
      NAMESPACE="$2"
      shift
      shift
    ;;
    -s|--setup)
      ACTION="setup"
      shift
    ;;
    -d|--destroy)
      ACTION="destroy"
      shift
    ;;
    -l|--list)
      ACTION="list"
      shift
    ;;
    -h|--help)
      ACTION="help"
      shift
    ;;
    *)
      DEPLOYMENT="$1"
      shift # past argument
      break
      ;;
  esac
done

DEPLOYMENTPATH=""

function resolveDeploymentPathAndNamespace() {
  # xargs removes spurious whitespace. WHAT?!
  matchesCount=$(find "$BASEPATH" -name "$DEPLOYMENT.yaml" | wc -l | xargs)

  if [[ "$matchesCount" = "0" ]]; then
    echo "Deployment $DEPLOYMENT could not be found"
    exit 1
  elif [[ "$matchesCount" = "1" ]]; then
    DEPLOYMENTPATH=$(find "$BASEPATH" -name "$DEPLOYMENT.yaml")
    if [[ "$NAMESPACE" = "" ]]; then
      NAMESPACE=$DEPLOYMENT
    fi
    echo "Using deployment at $DEPLOYMENTPATH"
  else
    echo "Found $matchesCount deployments with name $DEPLOYMENT.yaml. I can't choose!"
    exit 1
  fi
}

if [[ "$ACTION" = "setup" ]]; then
  resolveDeploymentPathAndNamespace

  echo "Assuring namespace $NAMESPACE"
  kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

  echo "Installing demo deployment into $NAMESPACE"
  kubectl apply --namespace "$NAMESPACE" -f "$DEPLOYMENTPATH"

elif [[ "$ACTION" = "destroy" ]]; then
  resolveDeploymentPathAndNamespace

  echo "Deleting http-error-monitor from $NAMESPACE"
  kubectl delete --namespace "$NAMESPACE" -f "$DEPLOYMENTPATH"

  echo "Removing namespace $NAMESPACE"
  kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl delete -f -

elif [[ "$ACTION" = "list" ]]; then
  echo "Listing available deployments."

  find "$BASEPATH" -name \*.yaml | rev | cut -d '/' -f1 | rev | cut -d '.' -f1
elif [[ "$ACTION" = "help" ]]; then
  printUsage
elif [[ "$ACTION" = "" ]]; then
  echo "No action specified, please specify -s/-d/-l/-h"
  printUsage
  exit 1
else
  echo "Unknown action $ACTION"
  printUsage
  exit 1
fi
