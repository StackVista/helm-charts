#!/usr/bin/env bash

set -euo pipefail

# This script will use the
#
# STACKSTATE_SYSTEM_USER_NAME
# STACKSTATE_SYSTEM_USER_EMAIL
#

# Configure user matching the gitlab account and gpg key
git config --global user.email "$STACKSTATE_SYSTEM_USER_EMAIL"
git config --global user.name "$STACKSTATE_SYSTEM_USER_NAME"
