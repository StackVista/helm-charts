#!/usr/bin/env bash
# Exit on error. Append "|| true" if you expect an error.
set -o errexit
# Exit on error inside any functions or subshells.
set -o errtrace
# Do not allow use of undefined vars. Use ${VAR:-} to use an undefined VAR
set -o nounset
# Catch an error in command pipes.
set -o pipefail
# Turn on traces, useful while debugging.
set -o xtrace


DIR=${BASH_SOURCE%/*}
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
if [[ "$DIR" = "." ]]; then DIR="$PWD"; fi
FAULTY_IMAGE="quay.io/stackstate/weaveworksdemo-catalogue:0.3.6"
ORG_IMAGE="quay.io/stackstate/weaveworksdemo-catalogue:0.3.5"

# Read the current deployment image and based on that switch the scenario
IMAGE=$(kubectl get deployment catalogue -o=jsonpath='{$.spec.template.spec.containers[:1].image}')
echo "Current image found is: $IMAGE"

case $IMAGE in
  "${FAULTY_IMAGE}")
    echo "Rolling back the faulty version"
    pwd
    ls -lrt "$DIR"
    kubectl patch deployment catalogue --patch-file "$DIR"/catalogue-fix-patch.yaml
    ;;

  "${ORG_IMAGE}")
    echo "Deploying the faulty version"
    kubectl patch deployment catalogue --patch-file "$DIR"/catalogue-failure-patch.yaml
  ;;

  *)
    echo "Could not recognise the current image. Not doing anything with the scenario"
    exit 1
    ;;
esac
