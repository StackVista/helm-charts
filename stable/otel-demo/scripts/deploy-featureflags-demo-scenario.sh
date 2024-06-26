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

SCENARIO=${1:?First argument must be the scenario name: <failure> or <fix>}
NEW_IMAGE="${2:?Second argument must be the expected image}"
DEPLOYMENT_NAME="${3:?Third argument must be the deployment name of the feature flag service}"

# Read the current deployment image and based on that switch the scenario
IMAGE=$(kubectl get deployment "$DEPLOYMENT_NAME" -o=jsonpath='{$.spec.template.spec.containers[:1].image}')
echo "Current image found is: $IMAGE"

case $SCENARIO in
  "fix")
    if [ "$IMAGE" == "$NEW_IMAGE" ] ; then
      echo "The current image is already the good one. Doing nothing."
    else
      echo "Deploying the good version"
      kubectl patch deployment "$DEPLOYMENT_NAME" --patch-file "$DIR"/featureflags-fix-patch.yaml
    fi
    ;;

  "failure")
    if [ "$IMAGE" == "$NEW_IMAGE" ] ; then
      echo "The current image is already the faulty one. Doing nothing."
    else
      echo "Deploying the faulty version"
      kubectl patch deployment "$DEPLOYMENT_NAME" --patch-file "$DIR"/featureflags-failure-patch.yaml
    fi
  ;;

  *)
    echo "The scenario name<$SCENARIO> is not recognised: must be on of <failure> or <fix>.  Doing nothing."
    exit 1
    ;;
esac
