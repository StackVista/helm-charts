#!/usr/bin/env bash

set -euo pipefail

# shellcheck disable=SC1091
source "$CI_PROJECT_DIR/.gitlab/gpg_utils.sh"
check_git_configuration || configure_git_user

# Fetch remotes branches, we need it to find branches to push commits with updated Chart version
git fetch --all -q
