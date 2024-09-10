#!/bin/bash

set -euo pipefail

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
# shellcheck disable=SC1091
source "$dir/util.sh"

helm dependencies update stable/suse-observability-values

get_secret_values "sops.rancher-helm-credentials.yaml"

.gitlab/package-and-push-chart-for-rancher.sh suse-observability-values
